package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
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
	smClient      *secretsmanager.Client

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
	smClient = secretsmanager.NewFromConfig(cfg)

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
	deploymentId := utils.GenerateUUID()

	var input dtos.DeployInput
	err := json.Unmarshal([]byte(event.Body), &input)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	exist, err := storageClient.CheckExistedBackendStack(ctx, userId, input.StackName)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to check for existed backend stack: %w", err)
	}
	if exist {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusFound,
			Body:       "Stack with selected name already existed",
		}, nil
	}

	pending, err := storageClient.CheckPendingDeployment(ctx, userId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to check pending deployment: %w", err)
	}
	if pending {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusConflict,
			Body:       "Another deployment is ongoing",
		}, nil
	}

	funcMap := template.FuncMap{
		"mul": func(a float64, b int) int {
			return int(a * float64(b))
		},
	}

	for _, stack := range stacks {
		tmplContent, err := templatesFS.ReadFile(fmt.Sprintf("templates/%s.tmpl", stack))
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to read template content: %w", err)
		}

		// Parse the template
		t := template.Must(template.New(stack).Funcs(funcMap).Parse(string(tmplContent)))

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

	var secretArn string
	if input.ServerConfiguration.ContainerImage.IsPrivate {
		secretName := fmt.Sprintf("%s-registry-credentials", input.StackName)
		secretValue, _ := json.Marshal(input.ServerConfiguration.ContainerImage.RegistryCredentials)

		output, err := smClient.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
			Name:         aws.String(secretName),
			SecretString: aws.String(string(secretValue)),
		})
		if err != nil {
			if isResourceExistsError(err) {
				// Secret exists, update it
				output, err := smClient.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
					SecretId:     aws.String(secretName),
					SecretString: aws.String(string(secretValue)),
				})
				if err != nil {
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusInternalServerError,
					}, fmt.Errorf("failed to update existing secret: %w", err)
				}
				secretArn = *output.ARN
			} else {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
				}, fmt.Errorf("failed to create secret: %w", err)
			}
		} else {
			secretArn = *output.ARN
		}
	}

	jobInput := &batch.SubmitJobInput{
		JobName:       aws.String(batchJobName),
		JobQueue:      aws.String(batchJobQueue),
		JobDefinition: aws.String(batchJobDefinition),
		Tags: map[string]string{
			"deploymentId": deploymentId,
		},

		// Optional: Pass environment variables or overrides
		ContainerOverrides: &types.ContainerOverrides{
			Environment: []types.KeyValuePair{
				{
					Name:  aws.String("STACK_NAME"),
					Value: aws.String(input.StackName),
				},
				{
					Name:  aws.String("SERVER_IMAGE_URI"),
					Value: aws.String(input.ServerConfiguration.ContainerImage.Uri),
				},
				{
					Name:  aws.String("REGISTRY_CREDENTIALS_ARN"),
					Value: aws.String(secretArn),
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
		Id:        deploymentId,
		UserId:    userId,
		BackendId: backendId,
		Status:    "pending",
		Input:     dtos.DeployInputRequestToEntity(input),
		CreatedAt: time.Now(),
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

func isResourceExistsError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "ResourceExistsException"
	}
	return false
}

func main() {
	lambda.Start(handler)
}
