package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/compute"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

var (
	storageClient *storage.Client
	computeClient *compute.Client

	clusterName = os.Getenv("ECS_CLUSTER_NAME")
	serviceName = os.Getenv("ECS_SERVICE_NAME")

	ErrUserNotInMatch = fmt.Errorf("user not in match")
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
	computeClient = compute.NewClient(
		ecs.NewFromConfig(cfg),
		ec2.NewFromConfig(cfg),
	)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	userId := auth.MustAuth(event.RequestContext.Authorizer)
	matchId := event.PathParameters["id"]

	computeClient.CheckAndStartTask(ctx, clusterName, serviceName)

	activeMatch, err := storageClient.GetActiveMatch(ctx, matchId)
	if err != nil {
		if errors.Is(err, storage.ErrActiveMatchNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to get active match: %w", err)
	}
	if activeMatch.Player1.Id != userId && activeMatch.Player2.Id != userId {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to restore match: %w", ErrUserNotInMatch)
	}

	var serverIp string
	for range 5 {
		serverIp, err = computeClient.CheckAndGetNewServerIp(
			ctx,
			clusterName,
			serviceName,
			activeMatch.Server,
		)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get server ip: %w", err)
	}
	activeMatch.Server = serverIp

	if err := storageClient.UpdateActiveMatch(
		ctx,
		matchId,
		storage.ActiveMatchUpdateOptions{
			Server: aws.String(serverIp),
		},
	); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to update active match: %w", err)
	}

	resp := dtos.ActiveMatchResponseFromEntity(activeMatch)
	respJson, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal response: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(respJson),
	}, nil
}

func main() {
	lambda.Start(handler)
}
