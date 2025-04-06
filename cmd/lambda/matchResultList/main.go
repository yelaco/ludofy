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
	targetId, startKey, limit, err := extractScanParameters(
		userId,
		event.QueryStringParameters,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest},
			fmt.Errorf("failed to extract parameters: %w", err)
	}
	matchResults, lastEvalKey, err := storageClient.FetchMatchResults(
		ctx,
		targetId,
		startKey,
		limit,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to fetch match results: %w", err)
	}

	resp := dtos.MatchResultListResponseFromEntities(matchResults)
	if lastEvalKey != nil {
		resp.NextPageToken = &dtos.NextMatchResultPageToken{
			Timestamp: lastEvalKey["Timestamp"].(*types.AttributeValueMemberS).Value,
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

func extractScanParameters(
	userId string,
	params map[string]string,
) (
	string,
	map[string]types.AttributeValue,
	int32,
	error,
) {
	var targetId string
	if userIdStr, ok := params["userId"]; ok {
		targetId = userIdStr
	} else {
		targetId = userId
	}

	limit := 10
	if limitStr, ok := params["limit"]; ok {
		limitInt64, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return "", nil, 0, fmt.Errorf("invalid limit: %v", err)
		}
		limit = int(limitInt64)
	}

	// Check for startKey (optional)
	var startKey map[string]types.AttributeValue
	if startKeyStr, ok := params["startKey"]; ok {
		var nextPageToken dtos.NextMatchResultPageToken
		if err := json.Unmarshal(
			[]byte(startKeyStr),
			&nextPageToken,
		); err != nil {
			return "", nil, 0, err
		}
		startKey = map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
			"Timestamp": &types.AttributeValueMemberS{
				Value: nextPageToken.Timestamp,
			},
		}
	}

	return targetId, startKey, int32(limit), nil
}

func main() {
	lambda.Start(handler)
}
