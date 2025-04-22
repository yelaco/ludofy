package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/paas/aws/auth"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
)

var (
	storageClient *storage.Client
	cfClient      *cloudformation.Client
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(
		dynamodb.NewFromConfig(cfg),
		nil,
	)
	cfClient = cloudformation.NewFromConfig(cfg)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	userId := auth.MustAuth(event.RequestContext.Authorizer)
	backendId := event.PathParameters["id"]

	backend, err := storageClient.GetBackend(ctx, backendId)
	if err != nil {
		if errors.Is(err, storage.ErrBackendNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get backend: %w", err)
	}
	if backend.UserId != userId {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
		}, fmt.Errorf("invalid user id")
	}

	_, err = cfClient.DeleteStack(ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(backend.StackName),
		RoleARN:   aws.String(os.Getenv("DEPLOY_JOB_ROLE_ARN")),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to delete stack: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
