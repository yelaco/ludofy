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

const (
	ASC  = true
	DESC = false
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
	startKey, limit, order, err := extractScanParameters(
		matchId,
		event.QueryStringParameters,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to extract parameters: %w", err)
	}
	matchStates, lastEvalKey, err := storageClient.FetchMatchStates(
		ctx,
		matchId,
		startKey,
		limit,
		order,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to fetch match states: %w", err)
	}

	resp := dtos.MatchStateListResponseFromEntities(matchStates)
	if lastEvalKey != nil {
		resp.NextPageToken = &dtos.NextMatchStatePageToken{
			Id:  lastEvalKey["Id"].(*types.AttributeValueMemberS).Value,
			Ply: lastEvalKey["Ply"].(*types.AttributeValueMemberN).Value,
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
	matchId string,
	params map[string]string,
) (
	map[string]types.AttributeValue,
	int32,
	bool,
	error,
) {
	limit := 20
	if limitStr, ok := params["limit"]; ok {
		limitInt64, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return nil, 0, false, fmt.Errorf("invalid limit: %v", err)
		}
		limit = int(limitInt64)
	}

	// Check for startKey (optional)
	var startKey map[string]types.AttributeValue
	if startKeyStr, ok := params["startKey"]; ok {
		var nextPageToken dtos.NextMatchStatePageToken
		err := json.Unmarshal([]byte(startKeyStr), &nextPageToken)
		if err != nil {
			return nil, 0, false, err
		}
		startKey = map[string]types.AttributeValue{
			"Id":      &types.AttributeValueMemberS{Value: nextPageToken.Id},
			"MatchId": &types.AttributeValueMemberS{Value: matchId},
			"Ply":     &types.AttributeValueMemberN{Value: nextPageToken.Ply},
		}
	}

	var order bool
	if orderStr, ok := params["order"]; ok {
		if orderStr == "asc" {
			order = ASC
		}
	}

	return startKey, int32(limit), order, nil
}

func main() {
	lambda.Start(handler)
}
