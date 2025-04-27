package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/paas/aws/auth"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
	"github.com/chess-vn/slchess/internal/paas/domains/dtos"
)

var (
	storageClient *storage.Client
	cfClient      *cloudformation.Client
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg), nil)
	cfClient = cloudformation.NewFromConfig(cfg)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	auth.MustAuth(event.RequestContext.Authorizer)
	id := event.PathParameters["id"]

	backend, err := storageClient.GetBackend(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrBackendNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to get backend: %w", err)
	}

	// Describe main stack
	resp := dtos.BackendResponseFromEntity(backend)
	description, err := cfClient.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(backend.StackName),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to describe stack: %w", err)
	}
	if len(description.Stacks) > 0 {
		outputs := make(map[string]string, len(description.Stacks[0].Outputs))
		for _, output := range description.Stacks[0].Outputs {
			outputs[aws.ToString(output.OutputKey)] = aws.ToString(output.OutputValue)
		}
		resp.Outputs = outputs
	}

	// Describe customization stack
	listStacksOutput, err := cfClient.ListStacks(ctx, &cloudformation.ListStacksInput{
		StackStatusFilter: []types.StackStatus{
			types.StackStatusCreateComplete,
			types.StackStatusUpdateComplete,
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to list stacks: %w", err)
	}

	prefix := fmt.Sprintf("%s-CustomizationStack", backend.StackName)

	var matchedStackNames []string
	for _, summary := range listStacksOutput.StackSummaries {
		if strings.HasPrefix(aws.ToString(summary.StackName), prefix) {
			matchedStackNames = append(matchedStackNames, aws.ToString(summary.StackName))
		}
	}

	for _, stackName := range matchedStackNames {
		describeOutput, err := cfClient.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			continue
		}

		if len(description.Stacks) > 0 {
			for _, output := range describeOutput.Stacks[0].Outputs {
				resp.Outputs[aws.ToString(output.OutputKey)] = aws.ToString(output.OutputValue)
			}
		}
	}

	// Return
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
