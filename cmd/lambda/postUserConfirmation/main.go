package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var storageClient *storage.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handler(
	ctx context.Context,
	event events.CognitoEventUserPoolsPostConfirmation,
) (
	events.CognitoEventUserPoolsPostConfirmation,
	error,
) {
	userId := event.Request.UserAttributes["sub"]
	username := event.UserName

	// Default user profile
	err := storageClient.PutUserProfile(ctx, entities.UserProfile{
		UserId:     userId,
		Username:   username,
		Membership: "guest",
		CreatedAt:  time.Now(),
	})
	if err != nil {
		return event, fmt.Errorf("failed to put user profile: %w", err)
	}

	// Initial puzzle profile
	err = storageClient.PutPuzzleProfile(ctx, entities.PuzzleProfile{
		UserId: userId,
		Rating: 300,
	})
	if err != nil {
		return event, fmt.Errorf("failed to put puzzle profile: %w", err)
	}

	// Default user rating
	err = storageClient.PutUserRating(ctx, entities.UserRating{
		UserId:       userId,
		Rating:       1200,
		RD:           200,
		PartitionKey: "UserRatings",
	})
	if err != nil {
		return event, fmt.Errorf("failed to put user rating: %w", err)
	}

	return event, nil
}

func main() {
	lambda.Start(handler)
}
