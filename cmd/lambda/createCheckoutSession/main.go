package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type response struct {
	Url string `json:"url"`
}

func handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse,
	error,
) {
	userId := auth.MustAuth(event.RequestContext.Authorizer)
	membership, err := extractParameters(event.QueryStringParameters)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("failed to extract parameters: %w", err)
	}
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	var priceId string
	switch membership {
	case "plus":
		priceId = os.Getenv("PLUS_PRICE_ID")
	case "premium":
		priceId = os.Getenv("PREMIUM_PRICE_ID")
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("unknown membership type: %s", membership)
	}
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String("subscription"),
		SuccessURL:         stripe.String("https://slchess.vn"),
		CancelURL:          stripe.String("https://slchess.vn"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceId),
				Quantity: stripe.Int64(1),
			},
		},
		ClientReferenceID: stripe.String(userId),
		Metadata: map[string]string{
			"membership": membership,
		},
	}

	s, err := session.New(params)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("failed to create checkout session: %w", err)
	}

	resp := response{
		Url: s.URL,
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

func extractParameters(params map[string]string) (
	string,
	error,
) {
	membership, ok := params["membership"]
	if !ok {
		return "", fmt.Errorf("membership type not provided")
	}

	return membership, nil
}

func main() {
	lambda.Start(handler)
}
