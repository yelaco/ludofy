package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/notification"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/entities"
	"github.com/chess-vn/slchess/pkg/utils"
)

var (
	storageClient *storage.Client
	notiClient    *notification.Client
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
	notiClient = notification.NewClient(sns.NewFromConfig(cfg))
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	receiverId := auth.MustAuth(event.RequestContext.Authorizer)
	senderId := event.PathParameters["id"]
	conversationId := utils.GenerateUUID()
	startedAt := time.Now()

	err := storageClient.PutFriendship(ctx, entities.Friendship{
		UserId:         receiverId,
		FriendId:       senderId,
		ConversationId: conversationId,
		StartedAt:      startedAt,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to put friendship: %w", err)
	}

	err = storageClient.PutFriendship(ctx, entities.Friendship{
		UserId:         senderId,
		FriendId:       receiverId,
		ConversationId: conversationId,
		StartedAt:      startedAt,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to put friendship: %w", err)
	}

	err = storageClient.DeleteFriendRequest(ctx, senderId, receiverId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to delete friend request: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
