package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
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
	startTime, endTime, interval, err := extractParameters(event.QueryStringParameters)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to extract parameters: %w", err)
	}

	serverIps, err := computeClient.GetServerIps(ctx, clusterName, serviceName)
	if err != nil {
		if !errors.Is(err, compute.ErrNoServerRunning) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get server ips: %w", err)
		}
	}

	serverMetricsList := make([]dtos.ServerMetricsResponse, 0, len(serverIps))
	for _, serverIp := range serverIps {
		status, err := computeClient.GetServerStatus(serverIp, 7202)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get server status: %w", err)
		}
		serverMetricsList = append(serverMetricsList, status)
	}

	serviceMetricsList, err := computeClient.GetServiceMetrics(
		ctx,
		startTime,
		endTime,
		interval,
		clusterName,
		serviceName,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get service metrics: %w", err)
	}

	resp := dtos.BackendMetricsResponse{
		ServiceMetrics:    serviceMetricsList,
		ServerMetricsList: serverMetricsList,
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

func extractParameters(
	params map[string]string,
) (
	time.Time,
	time.Time,
	int32,
	error,
) {
	startStr, ok := params["start"]
	if !ok {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("missing start time")
	}
	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("failed to parse start time: %w", err)
	}

	endStr, ok := params["end"]
	if !ok {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("missing start time")
	}
	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("failed to parse end time: %w", err)
	}

	intervalStr, ok := params["interval"]
	if !ok {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("missing interval")
	}
	intervalInt64, err := strconv.ParseInt(intervalStr, 10, 32)
	if err != nil {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("invalid limit: %v", err)
	}
	interval := int32(intervalInt64)

	return startTime, endTime, interval, nil
}

func main() {
	lambda.Start(handler)
}
