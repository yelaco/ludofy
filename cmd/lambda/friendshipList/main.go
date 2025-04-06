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
	friendships, lastEvalKey, err := storageClient.FetchFriendships(
		ctx,
		userId,
		startKey,
		limit,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to fetch friendships: %w", err)
	}

	resp := dtos.FriendshipListResponseFromEntities(friendships)
	if lastEvalKey != nil {
		resp.NextPageToken = &dtos.NextFriendshipPageToken{
			FriendId: lastEvalKey["FriendId"].(*types.AttributeValueMemberS).Value,
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
	limit := 10
	if limitStr, ok := params["limit"]; ok {
		limitInt64, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid limit: %v", err)
		}
		limit = int(limitInt64)
	}

	// Check for startKey (optional)
	var startKey map[string]types.AttributeValue
	if startKeyStr, ok := params["startKey"]; ok {
		var nextPageToken dtos.NextFriendshipPageToken
		if err := json.Unmarshal(
			[]byte(startKeyStr),
			&nextPageToken,
		); err != nil {
			return nil, 0, err
		}
		startKey = map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
			"FriendId": &types.AttributeValueMemberS{
				Value: nextPageToken.FriendId,
			},
		}
	}

	return startKey, int32(limit), nil
}

func main() {
	lambda.Start(handler)
}
