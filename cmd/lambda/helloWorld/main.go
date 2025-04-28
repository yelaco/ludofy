package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Hello World!",
	}, nil
}

func main() {
	lambda.Start(handler)
}
