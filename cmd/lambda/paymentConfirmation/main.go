package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var storageClient *storage.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

// Handle Stripe Webhook
func handler(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	// Verify the webhook signature
	event, err := webhook.ConstructEvent(
		[]byte(request.Body),
		request.Headers["stripe-signature"],
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to verfiy webhook signature: %w", err)
	}

	// Process event
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
			}, fmt.Errorf("failed to parse session data: %w", err)
		}

		// productName, err := getSubscriptionProductName(session.Subscription.ID)
		// if err != nil {
		// 	return events.APIGatewayProxyResponse{
		// 		StatusCode: http.StatusBadRequest,
		// 	}, fmt.Errorf("failed to get product name: %w", err)
		// }
		// var membership string
		// switch productName {
		// case "Slchess Plus":
		// 	membership = "plus"
		// case "Slchess Premium":
		// 	membership = "premium"
		// }
		membership := session.Metadata["membership"]
		userId := session.ClientReferenceID
		err = storageClient.UpdateUserProfile(
			ctx,
			userId,
			storage.UserProfileUpdateOptions{
				Membership: aws.String(membership),
			},
		)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("failed to update user profile: %w", err)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

// func getSubscriptionProductName(subscriptionId string) (string, error) {
// 	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
//
// 	// Retrieve sub details
// 	sub, err := subscription.Get(subscriptionId, nil)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get subscription: %w", err)
// 	}
//
// 	if len(sub.Items.Data) == 0 {
// 		return "", fmt.Errorf("no subscription item")
// 	}
//
// 	return sub.Items.Data[0].Price.Product.Name, nil
// }

func main() {
	lambda.Start(handler)
}
