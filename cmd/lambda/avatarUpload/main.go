package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/presign"
)

type response struct {
	Url string `json:"url"`
}

var (
	s3Client  *s3.Client
	presigner *presign.Presigner

	bucketName = os.Getenv("IMAGE_BUCKET_NAME")
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	s3Client = s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)
	presigner = presign.NewPresigner(presignClient)
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	userId := auth.MustAuth(event.RequestContext.Authorizer)
	uploadKey := "avatars/" + userId
	presignedPutRequest, err := presigner.PutObject(ctx, bucketName, uploadKey, 60)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to presign put object request: %w", err)
	}
	resp := response{
		Url: presignedPutRequest.URL,
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
