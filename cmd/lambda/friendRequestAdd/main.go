package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
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
	senderId := auth.MustAuth(event.RequestContext.Authorizer)
	receiverId := event.PathParameters["id"]

	friendship, err := storageClient.GetFriendship(ctx, senderId, receiverId)
	if err == nil {
		resp := dtos.FriendshipResponseFromEntity(friendship)
		respJson, err := json.Marshal(resp)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to marshal response: %w", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusConflict,
			Body:       string(respJson),
		}, nil
	} else if !errors.Is(err, storage.ErrFriendshipNotFound) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get friendship: %w", err)
	}

	friendRequest := entities.FriendRequest{
		SenderId:   senderId,
		ReceiverId: receiverId,
		CreatedAt:  time.Now(),
	}
	err = storageClient.PutFriendRequest(ctx, friendRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to put friend request: %w", err)
	}

	endpoints, err := storageClient.FetchApplicationEndpoints(ctx, receiverId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get application endpoint: %w", err)
	}

	msg, err := json.Marshal(dtos.FriendRequestResponseFromEntity(friendRequest))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, endpoint := range endpoints {
		err = notiClient.SendPushNotification(
			ctx,
			endpoint.EndpointArn,
			string(msg),
		)
		if err != nil {
			storageClient.DeleteApplicationEndpoint(
				ctx,
				endpoint.UserId,
				endpoint.DeviceToken,
			)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to send push notification: %w", err)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
