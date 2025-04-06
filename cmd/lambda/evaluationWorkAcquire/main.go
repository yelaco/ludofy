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
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/chess-vn/slchess/internal/aws/analysis"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

var analysisClient *analysis.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	analysisClient = analysis.NewClient(
		nil,
		sqs.NewFromConfig(cfg),
	)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	evaluationWork, err := analysisClient.AcquireEvaluationWork(ctx)
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

	resp := dtos.EvaluationWorkResponseFromEntity(evaluationWork)
	respJson, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal response: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusAccepted,
		Body:       string(respJson),
	}, nil
}

func main() {
	lambda.Start(handler)
}
