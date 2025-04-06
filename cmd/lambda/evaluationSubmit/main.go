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
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/chess-vn/slchess/internal/aws/analysis"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

var (
	apigatewayClient *apigatewaymanagementapi.Client
	analysisClient   *analysis.Client
	storageClient    *storage.Client

	region            = os.Getenv("AWS_REGION")
	websocketApiId    = os.Getenv("WEBSOCKET_API_ID")
	websocketApiStage = os.Getenv("WEBSOCKET_API_STAGE")
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	apiEndpoint := fmt.Sprintf(
		"https://%s.execute-api.%s.amazonaws.com/%s",
		websocketApiId,
		region,
		websocketApiStage,
	)
	apigatewayClient = apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
		BaseEndpoint: aws.String(apiEndpoint),
		Region:       region,
		Credentials:  cfg.Credentials,
	})
	analysisClient = analysis.NewClient(
		nil,
		sqs.NewFromConfig(cfg),
	)
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	stop := extractParameters(event.QueryStringParameters)

	var submission dtos.EvaluationSubmission
	if err := json.Unmarshal([]byte(event.Body), &submission); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	eval := dtos.EvaluationResultToEntity(submission.Evaluation)
	evalJson, err := json.Marshal(dtos.EvaluationResponseFromEntity(eval))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal evaluation: %w", err)
	}
	_, err = apigatewayClient.PostToConnection(
		ctx,
		&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(submission.ConnectionId),
			Data:         evalJson,
		},
	)
	if err != nil {
		var goneErr *types.GoneException
		if !errors.As(err, &goneErr) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to post to connect: %w", err)
		}
	}

	err = storageClient.PutEvaluation(ctx, eval, 24*time.Hour)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to put evaluation: %w", err)
	}

	err = analysisClient.RemoveEvaluationWork(ctx, submission.ReceiptHandle)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to remove evaluation work: %w", err)
	}

	if stop {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}, nil
	}

	// Assign more work
	newWork, err := analysisClient.AcquireEvaluationWork(ctx)
	if err != nil {
		if errors.Is(err, analysis.ErrEvaluationWorkNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNoContent,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to acquire work : %w", err)
	}
	resp := dtos.EvaluationWorkResponseFromEntity(newWork)
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
) bool {
	var stop bool
	if stopStr, ok := params["stop"]; ok {
		if stopStr == "true" {
			stop = true
		}
	}

	return stop
}

func main() {
	lambda.Start(handler)
}
