package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/entities"
	"github.com/golang-jwt/jwt/v5"
)

var (
	storageClient     *storage.Client
	cognitoPublicKeys map[string]*rsa.PublicKey
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("couldn't load config")
	}
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
	tokenSigningKeyUrl := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		os.Getenv("AWS_REGION"),
		os.Getenv("COGNITO_USER_POOL_ID"),
	)
	cognitoPublicKeys, err = auth.LoadCognitoPublicKeys(tokenSigningKeyUrl)
	if err != nil {
		panic("coulnd't load cognito public keys")
	}
}

// Handle WebSocket connection with authentication
func handler(
	ctx context.Context,
	event events.APIGatewayWebsocketProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	token := event.Headers["Authorization"]
	validToken, err := auth.ValidateJwt(token, cognitoPublicKeys)
	if err != nil || !validToken.Valid {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
		}, fmt.Errorf("failed to validate token: %w", err)
	}

	// Store connection in DynamoDB
	connection := entities.Connection{
		Id:     event.RequestContext.ConnectionID,
		UserId: validToken.Claims.(jwt.MapClaims)["sub"].(string),
	}
	if err = storageClient.PutConnection(ctx, connection); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to put conntection: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
