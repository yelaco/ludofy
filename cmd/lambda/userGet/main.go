package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

var storageClient *storage.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	userId := auth.MustAuth(event.RequestContext.Authorizer)
	targetId := event.PathParameters["id"]
	if targetId == "" {
		targetId = userId
	}

	userProfile, err := storageClient.GetUserProfile(ctx, targetId)
	if err != nil {
		if errors.Is(err, storage.ErrUserProfileNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, fmt.Errorf("failed to get user profile: %w", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get user profile: %w", err)
	}

	userRating, err := storageClient.GetUserRating(ctx, targetId)
	if err != nil {
		if errors.Is(err, storage.ErrUserRatingNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, fmt.Errorf("failed to get user rating: %w", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get user rating: %w", err)
	}

	// If users request their own information, return in full
	var getFull bool
	if userId == targetId {
		getFull = true
	}
	user := dtos.UserResponseFromEntities(userProfile, userRating, getFull)
	userJson, err := json.Marshal(user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal response: %w", err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(userJson),
	}, nil
}

func main() {
	lambda.Start(handler)
}
