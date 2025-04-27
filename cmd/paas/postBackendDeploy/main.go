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
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/yelaco/ludofy/internal/paas/aws/storage"
	"github.com/yelaco/ludofy/internal/paas/domains/dtos"
	"github.com/yelaco/ludofy/internal/paas/domains/entities"
)

var (
	storageClient *storage.Client
	batchClient   *batch.Client
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(
		dynamodb.NewFromConfig(cfg),
		s3.NewFromConfig(cfg),
	)
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

	deployment, err := storageClient.GetDeployment(ctx, deploymentId)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	var status string
	switch event.Status {
	case "SUCCEEDED":
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
	case "FAILED":
		prefix := fmt.Sprintf("%s/%s/", deployment.UserId, deployment.BackendId)
		err := storageClient.RemoveTemplates(ctx, prefix)
		if err != nil {
			return fmt.Errorf("failed to remove templates: %w", err)
		}
		status = "failed"
	case "RUNNING":
		status = "deploying"
	default:
		return fmt.Errorf("unknown status: %s", event.Status)
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
