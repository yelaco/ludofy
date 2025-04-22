package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
	"github.com/chess-vn/slchess/internal/paas/domains/dtos"
)

var (
	storageClient *storage.Client
	batchClient   *batch.Client
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg), nil)
	batchClient = batch.NewFromConfig(cfg)
}

// Handler function
func handler(ctx context.Context, event dtos.BackendRemoveEvent) error {
	output, err := batchClient.DescribeJobs(ctx, &batch.DescribeJobsInput{
		Jobs: []string{event.JobId},
	})
	if err != nil {
		return fmt.Errorf("failed to describe job: %w", err)
	}
	if len(output.Jobs) == 0 {
		return fmt.Errorf("unknown job with id: %s", event.JobId)
	}

	backendId, ok := output.Jobs[0].Tags["backendId"]
	if !ok {
		return fmt.Errorf("deployment id not found for job with id: %s", event.JobId)
	}

	if event.Status == "SUCCEEDED" {
		err = storageClient.DeleteBackend(ctx, backendId)
		if err != nil {
			return fmt.Errorf("failed to put backend: %w", err)
		}
	} else {
		opts := storage.BackendUpdateOptions{
			Status: aws.String("delete-failed"),
		}
		err = storageClient.UpdateBackend(ctx, backendId, opts)
		if err != nil {
			return fmt.Errorf("failed to update backend: %w", err)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
