package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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

	timeLayout  = "2006-01-02 15:04:05.999999999 -0700 MST"
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
	userRating, err := storageClient.GetUserRating(ctx, userId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get user rating: %w", err)
	}
	ticket := dtos.MatchmakingRequestToEntity(userRating, matchmakingReq)
	if err := ticket.Validate(); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("invalid ticket: %w", err)
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
	opponentIds, err := findOpponents(ctx, ticket)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to find opponents: %w", err)
	}

	// If no match found, queue the player by caching the matchmaking ticket
	if len(opponentIds) == 0 {
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
	for _, opponentId := range opponentIds {
		match, err := createMatch(
			ctx,
			userRating,
			opponentId,
			ticket.GameMode,
			serverIp,
		)
		if err != nil {
			continue
		}
		matchResp := dtos.ActiveMatchResponseFromEntity(match)
		matchRespJson, err := json.Marshal(matchResp)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to marshal response: %w", err)
		}

		// Notify the opponent about the match
		err = notifyQueueingUser(ctx, opponentId, matchRespJson)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to notify queueing user: %w", err)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(matchRespJson),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
	}, nil
}

// Matchmaking function using go-redis commands
func findOpponents(
	ctx context.Context,
	ticket entities.MatchmakingTicket,
) (
	[]string,
	error,
) {
	tickets, err := storageClient.ScanMatchmakingTickets(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to scan matchmaking tickets: %w", err)
	}

	var opponentIds []string
	if len(tickets) > 0 {
		for _, opTicket := range tickets {
			if opTicket.UserId == ticket.UserId {
				continue
			}
			opponentIds = append(opponentIds, opTicket.UserId)
		}
	} else {
		// No match found, add the user ticket to the queue
		storageClient.PutMatchmakingTickets(ctx, ticket)
	}

	return opponentIds, nil
}

func createMatch(
	ctx context.Context,
	userRating entities.UserRating,
	opponentId,
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

	// Associate the players with created match to kind of mark them as matched
	err := storageClient.PutUserMatch(ctx, entities.UserMatch{
		UserId:  opponentId,
		MatchId: match.MatchId,
	})
	if err != nil {
		return entities.ActiveMatch{}, err
	}

	err = storageClient.PutUserMatch(ctx, entities.UserMatch{
		UserId:  userRating.UserId,
		MatchId: match.MatchId,
	})
	if err != nil {
		return entities.ActiveMatch{}, err
	}

	// Pre-calculate players' rating in each possible outcome
	opponentRating, err := storageClient.GetUserRating(ctx, opponentId)
	if err != nil {
		return entities.ActiveMatch{},
			fmt.Errorf("failed to get user rating: %w", err)
	}

	newUserRatings, newUserRDs, err := calculateNewRatings(
		ctx,
		userRating,
		opponentRating,
	)
	if err != nil {
		return entities.ActiveMatch{}, err
	}

	newOpponentRatings, newOpponentRatingsRDs, err := calculateNewRatings(
		ctx,
		opponentRating,
		userRating,
	)
	if err != nil {
		return entities.ActiveMatch{}, err
	}

	match.Player1 = entities.Player{
		Id:         userRating.UserId,
		Rating:     userRating.Rating,
		RD:         userRating.RD,
		NewRatings: newUserRatings,
		NewRDs:     newUserRDs,
	}
	match.Player2 = entities.Player{
		Id:         opponentRating.UserId,
		Rating:     opponentRating.Rating,
		RD:         opponentRating.RD,
		NewRatings: newOpponentRatings,
		NewRDs:     newOpponentRatingsRDs,
	}
	match.AverageRating = (match.Player1.Rating + match.Player2.Rating) / 2

	// Save match information
	storageClient.PutActiveMatch(ctx, match)

	// Match created, remove opponent ticket from the queue
	err = storageClient.DeleteMatchmakingTickets(ctx, opponentId)
	if err != nil {
		return entities.ActiveMatch{},
			fmt.Errorf(
				"failed to delete matchmaking ticket: [userId: %s] %w",
				opponentId,
				err,
			)
	}

	// Create a conversation for spectators
	err = storageClient.PutSpectatorConversation(
		ctx,
		entities.SpectatorConversation{
			MatchId:        match.MatchId,
			ConversationId: utils.GenerateUUID(),
		},
	)
	if err != nil {
		return entities.ActiveMatch{}, err
	}

	return match, nil
}

func notifyQueueingUser(ctx context.Context, userId string, data []byte) error {
	// Get user ID from DynamoDB
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

	_, err = apigatewayClient.DeleteConnection(
		ctx,
		&apigatewaymanagementapi.DeleteConnectionInput{
			ConnectionId: aws.String(connection.Id),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
