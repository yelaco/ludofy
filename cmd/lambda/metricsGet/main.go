package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
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

	clusterName = os.Getenv("SERVER_CLUSTER_NAME")
	serviceName = os.Getenv("SERVER_SERVICE_NAME")
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
	computeClient = compute.NewClient(
		ecs.NewFromConfig(cfg),
		ec2.NewFromConfig(cfg),
		cloudwatch.NewFromConfig(cfg),
	)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	auth.MustAuth(event.RequestContext.Authorizer)

	serverIps, err := computeClient.GetServerIps(ctx, clusterName, serviceName)
	if err != nil {
		if !errors.Is(err, compute.ErrNoServerRunning) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get server ips: %w", err)
		}
	}

	serverStatuses := make([]dtos.ServerStatusResponse, 0, len(serverIps))
	for _, serverIp := range serverIps {
		status, err := computeClient.GetServerStatus(serverIp, 7202)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get service metrics: %w", err)
		}
		serverStatuses = append(serverStatuses, status)
	}

	serviceMetrics, err := computeClient.GetServiceMetrics(ctx, clusterName, serviceName)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get service metrics: %w", err)
	}

	resp := dtos.BackendMetricsResponse{
		ServiceMetrics: serviceMetrics,
		ServerStatuses: serverStatuses,
	}

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
