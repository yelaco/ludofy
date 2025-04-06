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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/chess-vn/slchess/internal/app/lichess"
	"github.com/chess-vn/slchess/internal/aws/analysis"
	"github.com/chess-vn/slchess/internal/aws/compute"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/pkg/logging"
	"go.uber.org/zap"
)

var (
	storageClient    *storage.Client
	computeClient    *compute.Client
	analysisClient   *analysis.Client
	lichessClient    *lichess.Client
	apigatewayClient *apigatewaymanagementapi.Client

	clusterName       = os.Getenv("STOFINET_CLUSTER_NAME")
	serviceName       = os.Getenv("STOFINET_SERVICE_NAME")
	region            = os.Getenv("AWS_REGION")
	websocketApiId    = os.Getenv("WEBSOCKET_API_ID")
	websocketApiStage = os.Getenv("WEBSOCKET_API_STAGE")
)

type Body struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
	computeClient = compute.NewClient(
		ecs.NewFromConfig(cfg),
		ec2.NewFromConfig(cfg),
	)
	analysisClient = analysis.NewClient(
		nil,
		sqs.NewFromConfig(cfg),
	)
	lichessClient = lichess.NewClient()
	apiEndpoint := fmt.Sprintf(
		"https://%s.execute-api.%s.amazonaws.com/%s",
		websocketApiId,
		region,
		websocketApiStage,
	)
	apigatewayClient = apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
		BaseEndpoint: aws.String(apiEndpoint),
		Region:       region,
		Credentials:  cfg.Credentials,
	})
}

func handler(
	ctx context.Context,
	event events.APIGatewayWebsocketProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	connectionId := aws.String(event.RequestContext.ConnectionID)
	var body Body
	if err := json.Unmarshal([]byte(event.Body), &body); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to unmarshal body: %w", err)
	}
	fen := body.Message

	// Query from lichess
	if eval, err := lichessClient.CloudEvaluate(fen); err == nil {
		resp := dtos.EvaluationResponseFromEntity(eval)
		respJson, err := json.Marshal(resp)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to marshal response: %w", err)
		}
		_, err = apigatewayClient.PostToConnection(
			ctx,
			&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: connectionId,
				Data:         respJson,
			},
		)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to post to connection: %w", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}, nil
	} else {
		if !errors.Is(err, lichess.ErrEvaluationNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to query lichess api : %w", err)
		}
	}

	// If not found, check in dynamodb table
	if eval, err := storageClient.GetEvaluation(ctx, fen); err == nil {
		resp := dtos.EvaluationResponseFromEntity(eval)
		respJson, err := json.Marshal(resp)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to marshal response: %w", err)
		}
		_, err = apigatewayClient.PostToConnection(
			ctx,
			&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: connectionId,
				Data:         respJson,
			},
		)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to post to connection: %w", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}, nil
	} else {
		if !errors.Is(err, storage.ErrEvaluationNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to get evaluation: %w", err)
		}
	}

	// If no cached evaluation found, submit request for new evaluation
	if err := analysisClient.SubmitEvaluationRequest(
		ctx,
		dtos.EvaluationRequest{
			ConnectionId: *connectionId,
			Fen:          fen,
		},
	); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to submit evaluation request: %w", err)
	}

	// Check and start an analysis node
	err := computeClient.CheckAndStartTask(ctx, clusterName, serviceName)
	if err != nil {
		logging.Info("failed to check and start task", zap.Error(err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
