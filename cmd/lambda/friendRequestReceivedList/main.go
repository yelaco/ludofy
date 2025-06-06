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
	userId := auth.MustAuth(event.RequestContext.Authorizer)

	startKey, limit, err := extractParameters(
		userId,
		event.QueryStringParameters,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to extract parameters: %w", err)
	}
	friendRequests, lastEvalKey, err := storageClient.FetchReceivedFriendRequests(
		ctx,
		userId,
		startKey,
		limit,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to fetch friend requests: %w", err)
	}

	resp := dtos.FriendRequestListResponseFromEntities(friendRequests)
	if lastEvalKey != nil {
		resp.NextPageToken = &dtos.NextFriendRequestPageToken{
			SenderId: lastEvalKey["SenderId"].(*types.AttributeValueMemberS).Value,
		}
		fmt.Println(lastEvalKey)
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

func extractParameters(
	userId string,
	params map[string]string,
) (
	map[string]types.AttributeValue,
	int32,
	error,
) {
	var limit int32 = 10
	if limitStr, ok := params["limit"]; ok {
		limitInt64, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid limit: %v", err)
		}
		limit = int32(limitInt64)
	}

	// Check for startKey (optional)
	var startKey map[string]types.AttributeValue
	if startKeyStr, ok := params["startKey"]; ok {
		var nextPageToken dtos.NextFriendRequestPageToken
		if err := json.Unmarshal(
			[]byte(startKeyStr),
			&nextPageToken,
		); err != nil {
			return nil, 0, err
		}
		startKey = map[string]types.AttributeValue{
			"SenderId": &types.AttributeValueMemberS{
				Value: nextPageToken.SenderId,
			},
			"ReceiverId": &types.AttributeValueMemberS{
				Value: userId,
			},
		}
	}

	return startKey, limit, nil
}

func main() {
	lambda.Start(handler)
}
