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
	auth.MustAuth(event.RequestContext.Authorizer)
	startKey, limit, err := extractScanParameters(event.QueryStringParameters)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to extract parameters: %w", err)
	}
	userRatings, lastEvalKey, err := storageClient.FetchUserRatings(
		ctx,
		startKey,
		limit,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to fetch user ratings: %w", err)
	}

	resp := dtos.UserRatingListResponseFromEntities(userRatings)
	if lastEvalKey != nil {
		resp.NextPageToken = &dtos.NextUserRatingPageToken{
			Rating: lastEvalKey["Rating"].(*types.AttributeValueMemberS).Value,
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

func extractScanParameters(params map[string]string) (
	map[string]types.AttributeValue,
	int32,
	error,
) {
	limitStr, ok := params["limit"]
	if !ok {
		return nil, 0, fmt.Errorf("missing parameter: limit")
	}

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid limit: %v", err)
	}

	// Check for startKey (optional)
	var startKey map[string]types.AttributeValue
	if startKeyStr, ok := params["startKey"]; ok {
		var nextPageToken dtos.NextUserRatingPageToken
		err := json.Unmarshal([]byte(startKeyStr), &nextPageToken)
		if err != nil {
			return nil, 0, err
		}
		startKey = map[string]types.AttributeValue{
			"Rating": &types.AttributeValueMemberS{
				Value: nextPageToken.Rating,
			},
		}
	}

	return startKey, int32(limit), nil
}

func main() {
	lambda.Start(handler)
}
