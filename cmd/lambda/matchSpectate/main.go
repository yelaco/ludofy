package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

var storageClient *storage.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	auth.MustAuth(event.RequestContext.Authorizer)
	matchId := event.PathParameters["id"]

	spectatorConversation, err := storageClient.GetSpectatorConversation(
		ctx,
		matchId,
	)
	if err != nil {
		if errors.Is(err, storage.ErrSpectatorConversationNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get spectator conversation: %w", err)
	}

	matchStates, lastEvaluatedKey, err := storageClient.FetchMatchStates(
		ctx,
		matchId,
		nil,
		20,
		false,
	)
	if err != nil {
		if !errors.Is(err, storage.ErrMatchStateNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get match state: %w", err)
		}
	}

	resp := dtos.NewMatchSpectateResponse(
		matchStates,
		spectatorConversation.ConversationId,
	)
	if lastEvaluatedKey != nil {
		resp.MatchStates.NextPageToken = &dtos.NextMatchStatePageToken{
			Id:  lastEvaluatedKey["Id"].(*types.AttributeValueMemberS).Value,
			Ply: lastEvaluatedKey["Ply"].(*types.AttributeValueMemberN).Value,
		}
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal response: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(respJson),
	}, nil
}

func main() {
	lambda.Start(handler)
}
