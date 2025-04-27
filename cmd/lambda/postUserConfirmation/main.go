package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mafredri/go-trueskill"
	"github.com/yelaco/ludofy/internal/aws/storage"
	"github.com/yelaco/ludofy/internal/domains/entities"
)

var (
	storageClient   *storage.Client
	ratingAlgorithm = os.Getenv("RATING_ALGORITHM")
)

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

	switch ratingAlgorithm {
	case "glicko":
		initialRatingStr := os.Getenv("INITIAL_RATING")
		initialRating, err := strconv.ParseFloat(initialRatingStr, 64)
		if err != nil {
			return event, fmt.Errorf("invalid initial rating: %w", err)
		}
		err = storageClient.PutUserRating(ctx, entities.UserRating{
			UserId:       userId,
			Rating:       initialRating,
			RD:           200,
			PartitionKey: "UserRatings",
		})
		if err != nil {
			return event, fmt.Errorf("failed to put user rating: %w", err)
		}
	case "trueskill":
		err = storageClient.PutUserRating(ctx, entities.UserRating{
			UserId:       userId,
			PartitionKey: "UserRatings",
			Rating:       trueskill.DefaultMu,
			Sigma:        trueskill.DefaultSigma,
		})
		if err != nil {
			return event, fmt.Errorf("failed to put user rating: %w", err)
		}
	}

	return event, nil
}

func main() {
	lambda.Start(handler)
}
