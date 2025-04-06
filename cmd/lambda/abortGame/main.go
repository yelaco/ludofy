package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

var storageClient *storage.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handler(ctx context.Context, event json.RawMessage) error {
	var req dtos.MatchAbortRequest
	if err := json.Unmarshal(event, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	err := storageClient.DeleteActiveMatch(ctx, req.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete active match: %w", err)
	}

	err = storageClient.DeleteSpectatorConversation(ctx, req.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete spectator conversation: %w", err)
	}

	for _, playerId := range req.PlayerIds {
		err = storageClient.DeleteUserMatch(ctx, playerId)
		if err != nil {
			return fmt.Errorf(
				"failed to delete user match: [userId: %s] - %w",
				playerId,
				err,
			)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
