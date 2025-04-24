package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/compute"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
	"github.com/chess-vn/slchess/pkg/utils"
)

var (
	storageClient    *storage.Client
	computeClient    *compute.Client
	apigatewayClient *apigatewaymanagementapi.Client

	clusterName       = os.Getenv("SERVER_CLUSTER_NAME")
	serviceName       = os.Getenv("SERVER_SERVICE_NAME")
	region            = os.Getenv("AWS_REGION")
	websocketApiId    = os.Getenv("WEBSOCKET_API_ID")
	websocketApiStage = os.Getenv("WEBSOCKET_API_STAGE")
	deploymentStage   = os.Getenv("DEPLOYMENT_STAGE")

	ErrNoMatchFound       = errors.New("failed to matchmaking")
	ErrInvalidGameMode    = errors.New("invalid game mode")
	ErrServerNotAvailable = errors.New("server not available")

	matchSize   = 2
	apiEndpoint = fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", websocketApiId, region, websocketApiStage)
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
	computeClient = compute.NewClient(
		ecs.NewFromConfig(cfg),
		ec2.NewFromConfig(cfg),
	)
	apigatewayClient = apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
		BaseEndpoint: aws.String(apiEndpoint),
		Region:       region,
		Credentials:  cfg.Credentials,
	})
	matchSizeStr := os.Getenv("MATCH_SIZE")
	matchSize, _ = strconv.Atoi(matchSizeStr)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	userId := auth.MustAuth(event.RequestContext.Authorizer)

	// Start game server beforehand if none available
	err := computeClient.CheckAndStartTask(ctx, clusterName, serviceName)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to start game server: %w", err)
	}

	// Extract and validate matchmaking ticket
	var matchmakingReq dtos.MatchmakingRequest
	err = json.Unmarshal([]byte(event.Body), &matchmakingReq)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to validate request: %w", err)
	}

	ticket := dtos.MatchmakingRequestToEntity(userId, matchmakingReq)
	if matchmakingReq.IsRanked {
		userRating, err := storageClient.GetUserRating(ctx, userId)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get user rating: %w", err)
		}
		ticket.UserRating = userRating.Rating
		if err := ticket.Validate(); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
			}, fmt.Errorf("invalid ticket: %w", err)
		}
	}

	// Check if user already in a activeMatch
	activeMatch, err := storageClient.CheckForActiveMatch(ctx, userId)
	if err != nil {
		if !errors.Is(err, storage.ErrUserMatchNotFound) &&
			!errors.Is(err, storage.ErrActiveMatchNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to check for active match: %w", err)
		}
	} else {
		var serverIp string
		for range 5 {
			serverIp, err = computeClient.CheckAndGetNewServerIp(
				ctx,
				clusterName,
				serviceName,
				activeMatch.Server,
			)
			if err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get server ip: %w", err)
		}
		activeMatch.Server = serverIp
		matchResp := dtos.ActiveMatchResponseFromEntity(activeMatch)
		matchRespJson, _ := json.Marshal(matchResp)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(matchRespJson),
		}, nil
	}

	// Attempt matchmaking
	playerIds, err := findMatchingPlayers(ctx, ticket)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to find opponents: %w", err)
	}

	// If no match found, queue the player by caching the matchmaking ticket
	if len(playerIds) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusAccepted,
			Body:       "Queued",
		}, nil
	}

	// Retrieve ip address of an available server
	var serverIp string
	for range 5 {
		serverIp, err = computeClient.GetServerIp(ctx, clusterName, serviceName)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get server ip: %w", err)
	}

	// Try to create new match
	playerIds = append(playerIds, userId)
	match, err := createMatch(
		ctx,
		playerIds,
		ticket.GameMode,
		serverIp,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to create match: %w", err)
	}
	matchResp := dtos.ActiveMatchResponseFromEntity(match)
	matchRespJson, err := json.Marshal(matchResp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal response: %w", err)
	}

	// Notify the other players about the match
	for _, playerId := range playerIds {
		err = notifyQueueingUser(ctx, playerId, matchRespJson)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to notify queueing user: %w", err)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(matchRespJson),
	}, nil
}

// Matchmaking function using go-redis commands
func findMatchingPlayers(
	ctx context.Context,
	ticket entities.MatchmakingTicket,
) (
	[]string,
	error,
) {
	tickets, err := storageClient.ScanMatchmakingTickets(ctx, ticket, matchSize-1)
	if err != nil {
		return nil, fmt.Errorf("failed to scan matchmaking tickets: %w", err)
	}

	var opponentIds []string
	if len(tickets) == matchSize-1 {
		for _, opTicket := range tickets {
			if opTicket.UserId == ticket.UserId {
				continue
			}
			opponentIds = append(opponentIds, opTicket.UserId)
		}
	} else {
		// Not enough players, add the user ticket to the pool
		storageClient.PutMatchmakingTickets(ctx, ticket)
	}

	return opponentIds, nil
}

func createMatch(
	ctx context.Context,
	playerIds []string,
	gameMode string,
	serverIp string,
) (
	entities.ActiveMatch,
	error,
) {
	match := entities.ActiveMatch{
		MatchId:        utils.GenerateUUID(),
		ConversationId: utils.GenerateUUID(),
		PartitionKey:   "ActiveMatches",
		GameMode:       gameMode,
		Server:         serverIp,
		CreatedAt:      time.Now(),
	}

	for _, playerId := range playerIds {
		match.Players = append(match.Players, entities.Player{
			Id: playerId,
		})
	}

	// Save match information
	if err := storageClient.TransactCreateMatch(ctx, match); err != nil {
		return entities.ActiveMatch{}, fmt.Errorf("failed to transact create match: %w", err)
	}

	// Create a conversation for spectators
	err := storageClient.PutSpectatorConversation(
		ctx,
		entities.SpectatorConversation{
			MatchId:        match.MatchId,
			ConversationId: utils.GenerateUUID(),
		},
	)
	if err != nil {
		return entities.ActiveMatch{}, fmt.Errorf("failed to put spectator conversation: %w", err)
	}

	return match, nil
}

func notifyQueueingUser(ctx context.Context, userId string, data []byte) error {
	connection, err := storageClient.GetConnectionByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrConnectionNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get connection: %w", err)
	}

	_, err = apigatewayClient.PostToConnection(
		ctx,
		&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connection.Id),
			Data:         data,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to post to connect: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
