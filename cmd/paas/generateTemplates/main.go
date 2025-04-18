package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chess-vn/slchess/internal/paas/aws/storage"
	"github.com/chess-vn/slchess/internal/paas/domains/dtos"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var (
	storageClient *storage.Client
	stacks        = []string{"log, storage, auth, compute, appsync, httpApi, websocketApi, game"}
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(
		dynamodb.NewFromConfig(cfg),
		s3.NewFromConfig(cfg),
	)
}

func handler(ctx context.Context, event json.RawMessage) error {
	var req dtos.DeploymentRequest
	if err := json.Unmarshal(event, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	for _, stack := range stacks {
		// Open the external template file
		tmplContent, err := templatesFS.ReadFile(fmt.Sprintf("templates/%s.tmpl", stack))
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Parse the template
		t := template.Must(template.New(stack).Parse(string(tmplContent)))

		// Render template into memory (buffer)
		var buf bytes.Buffer
		err = t.Execute(&buf, req)
		if err != nil {
			panic(err)
		}

		// Upload to S3
		key := fmt.Sprintf("%s/templates/%s.yaml", req.UserId, stack)

		err = storageClient.UploadTemplate(ctx, key, bytes.NewReader(buf.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to upload template: %w", err)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
