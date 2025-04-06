package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/chess-vn/slchess/internal/aws/analysis"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

var analysisClient *analysis.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	analysisClient = analysis.NewClient(
		athena.NewFromConfig(cfg),
		nil,
	)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	auth.MustAuth(event.RequestContext.Authorizer)
	puzzleId := event.PathParameters["id"]

	puzzle, err := analysisClient.GetPuzzle(ctx, puzzleId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get puzzle: %w", err)
	}

	resp := dtos.PuzzleResponseFromEntity(puzzle)
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
