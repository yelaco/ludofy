package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
	"github.com/chess-vn/slchess/internal/paas/domains/dtos"
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

var storageClient *storage.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg), nil)
}

// Handler function
func handler(ctx context.Context, event events.CloudWatchEvent) error {
	var deployEvent dtos.BackendDeployEvent
	if err := json.Unmarshal(event.Detail, &deployEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	deployment, err := storageClient.GetDeployment(ctx, deployEvent.DeploymentId)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	err = storageClient.PutBackend(ctx, entities.Backend{
		Id:        deployment.BackendId,
		UserId:    deployment.UserId,
		StackName: deployment.StackName,
	})
	if err != nil {
		return fmt.Errorf("failed to put backend: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
