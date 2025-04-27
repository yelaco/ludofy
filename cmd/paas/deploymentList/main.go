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
	"github.com/yelaco/ludofy/internal/paas/aws/auth"
	"github.com/yelaco/ludofy/internal/paas/aws/storage"
	"github.com/yelaco/ludofy/internal/paas/domains/dtos"
)

var storageClient *storage.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg), nil)
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
		event.QueryStringParameters,
		userId,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to extract parameters: %w", err)
	}
	deployments, lastEvalKey, err := storageClient.FetchDeployments(
		ctx,
		userId,
		startKey,
		limit,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to fetch deployments: %w", err)
	}

	resp := dtos.DeploymentListResponseFromEntities(deployments)
	if lastEvalKey != nil {
		resp.NextPageToken = &dtos.NextDeploymentPageToken{
			Id:        lastEvalKey["Id"].(*types.AttributeValueMemberS).Value,
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

func extractParameters(
	params map[string]string,
	userId string,
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
		var nextPageToken dtos.NextDeploymentPageToken
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
			"Id": &types.AttributeValueMemberS{
				Value: nextPageToken.Id,
			},
			"CreatedAt": &types.AttributeValueMemberS{
				Value: nextPageToken.CreatedAt,
			},
		}
	}

	return startKey, limit, nil
}

func main() {
	lambda.Start(handler)
}
