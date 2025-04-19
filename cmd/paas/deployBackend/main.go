package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chess-vn/slchess/internal/paas/aws/auth"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
	"github.com/chess-vn/slchess/internal/paas/domains/dtos"
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
	"github.com/chess-vn/slchess/pkg/utils"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var (
	stacks = []string{
		"log",
		"storage",
		"auth",
		"compute",
		"appsync",
		"httpApi",
		"websocketApi",
		"template",
	}
	storageClient *storage.Client
	batchClient   *batch.Client

	batchJobName       string
	batchJobQueue      string
	batchJobDefinition string
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(
		dynamodb.NewFromConfig(cfg),
		s3.NewFromConfig(cfg),
	)
	batchClient = batch.NewFromConfig(cfg)

	batchJobName = os.Getenv("BATCH_JOB_NAME")
	batchJobQueue = os.Getenv("BATCH_JOB_QUEUE")
	batchJobDefinition = os.Getenv("BATCH_JOB_DEFINITION")
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	userId := auth.MustAuth(event.RequestContext.Authorizer)
	backendId := utils.GenerateUUID()

	var input dtos.DeployInput
	err := json.Unmarshal([]byte(event.Body), &input)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	for _, stack := range stacks {
		tmplContent, err := templatesFS.ReadFile(fmt.Sprintf("templates/%s.tmpl", stack))
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to read template content: %w", err)
		}

		// Parse the template
		t := template.Must(template.New(stack).Parse(string(tmplContent)))

		buf := new(bytes.Buffer)
		err = t.Execute(buf, input)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to execute template: %w", err)
		}

		key := fmt.Sprintf("%s/%s/templates/%s.yaml", userId, backendId, stack)
		err = storageClient.UploadTemplate(ctx, key, buf)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to upload template: %w", err)
		}
	}

	jobInput := &batch.SubmitJobInput{
		JobName:       aws.String(batchJobName + "-" + userId),
		JobQueue:      aws.String(batchJobQueue),
		JobDefinition: aws.String(batchJobDefinition),

		// Optional: Pass environment variables or overrides
		ContainerOverrides: &types.ContainerOverrides{
			Environment: []types.KeyValuePair{
				{
					Name:  aws.String("STACK_NAME"),
					Value: aws.String(input.StackName),
				},
				{
					Name:  aws.String("SERVER_IMAGE_URI"),
					Value: aws.String(input.ServerImageUri),
				},
				{
					Name:  aws.String("USER_ID"),
					Value: aws.String(userId),
				},
				{
					Name:  aws.String("BACKEND_ID"),
					Value: aws.String(backendId),
				},
			},
		},
	}

	_, err = batchClient.SubmitJob(ctx, jobInput)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to submit deploy job: %w", err)
	}

	deployment := entities.Deployment{
		Id:        utils.GenerateUUID(),
		UserId:    userId,
		StackName: input.StackName,
		Status:    "pending",
	}
	if err := storageClient.PutDeployment(ctx, deployment); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to put deployment: %w", err)
	}

	resp := dtos.DeploymentResponseFromEntity(deployment)
	respJson, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to marshal response: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusAccepted,
		Body:       string(respJson),
	}, nil
}

func main() {
	lambda.Start(handler)
}
