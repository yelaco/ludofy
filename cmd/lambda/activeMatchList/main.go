package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/yelaco/ludofy/internal/aws/auth"
	"github.com/yelaco/ludofy/internal/aws/storage"
	"github.com/yelaco/ludofy/internal/domains/dtos"
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
	gameMode, startKey, limit, err := extractParameters(event.QueryStringParameters)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to extract paramters: %w", err)
	}
	activeMatches, lastEvalKey, err := storageClient.FetchActiveMatches(
		ctx,
		gameMode,
		startKey,
		limit,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to fetch active matches: %w", err)
	}

	resp := dtos.ActiveMatchListResponseFromEntities(activeMatches)
	if lastEvalKey != nil {
		resp.NextPageToken = &dtos.NextActiveMatchPageToken{
			CreatedAt: lastEvalKey["CreatedAt"].(*types.AttributeValueMemberS).Value,
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

func extractParameters(params map[string]string) (
	string,
	map[string]types.AttributeValue,
	int32,
	error,
) {
	gameMode := params["gameMode"]

	var limit int32 = 10
	limitStr, ok := params["limit"]
	if ok {
		limitInt64, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return "", nil, 0, fmt.Errorf("invalid limit: %v", err)
		}
		limit = int32(limitInt64)
	}

	// Check for startKey (optional)
	var startKey map[string]types.AttributeValue
	if startKeyStr, ok := params["startKey"]; ok {
		var nextPageToken dtos.NextActiveMatchPageToken
		if err := json.Unmarshal(
			[]byte(startKeyStr),
			&nextPageToken,
		); err != nil {
			return "", nil, 0, err
		}
		startKey = map[string]types.AttributeValue{
			"CreatedAt": &types.AttributeValueMemberS{
				Value: nextPageToken.CreatedAt,
			},
		}
	}

	return gameMode, startKey, limit, nil
}

func main() {
	lambda.Start(handler)
}
