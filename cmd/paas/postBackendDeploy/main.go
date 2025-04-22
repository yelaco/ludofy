package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
	"github.com/chess-vn/slchess/internal/paas/domains/dtos"
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
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
func handler(ctx context.Context, event dtos.BackendDeployEvent) error {
	output, err := batchClient.DescribeJobs(ctx, &batch.DescribeJobsInput{
		Jobs: []string{event.JobId},
	})
	if err != nil {
		return fmt.Errorf("failed to describe job: %w", err)
	}
	if len(output.Jobs) == 0 {
		return fmt.Errorf("unknown job with id: %s", event.JobId)
	}

	deploymentId, ok := output.Jobs[0].Tags["deploymentId"]
	if !ok {
		return fmt.Errorf("deployment id not found for job with id: %s", event.JobId)
	}

	status := "failed"
	if event.Status == "SUCCEEDED" {
		deployment, err := storageClient.GetDeployment(ctx, deploymentId)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}
		err = storageClient.PutBackend(ctx, entities.Backend{
			Id:        deployment.BackendId,
			UserId:    deployment.UserId,
			StackName: deployment.Input.StackName,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to put backend: %w", err)
		}
		status = "successful"
	}

	opts := storage.DeploymentUpdateOptions{
		Status: aws.String(status),
	}
	err = storageClient.UpdateDeployment(ctx, deploymentId, opts)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
