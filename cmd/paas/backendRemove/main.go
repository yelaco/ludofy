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
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/paas/aws/auth"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
)

var (
	storageClient *storage.Client
	batchClient   *batch.Client

	batchJobName       string
	batchJobQueue      string
	batchJobDefinition string
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(
		dynamodb.NewFromConfig(cfg),
		nil,
	)
	batchClient = batch.NewFromConfig(cfg)

	batchJobName = os.Getenv("BATCH_JOB_NAME")
	batchJobQueue = os.Getenv("BATCH_JOB_QUEUE")
	batchJobDefinition = os.Getenv("BATCH_JOB_DEFINITION")
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

	jobInput := &batch.SubmitJobInput{
		JobName:       aws.String(batchJobName),
		JobQueue:      aws.String(batchJobQueue),
		JobDefinition: aws.String(batchJobDefinition),
		Tags: map[string]string{
			"backendId": backendId,
		},

		// Optional: Pass environment variables or overrides
		ContainerOverrides: &types.ContainerOverrides{
			Environment: []types.KeyValuePair{
				{
					Name:  aws.String("STACK_NAME"),
					Value: aws.String(backend.StackName),
				},
				{
					Name:  aws.String("USER_ID"),
					Value: aws.String(userId),
				},
				{
					Name:  aws.String("BACKEND_ID"),
					Value: aws.String(backendId),
				},
			},
		},
	}
	_, err = batchClient.SubmitJob(ctx, jobInput)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to submit remove job: %w", err)
	}

	opts := storage.BackendUpdateOptions{
		Status: aws.String("delete-in-progress"),
	}
	err = storageClient.UpdateBackend(ctx, backendId, opts)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to update backend: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
