package notification

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func (client *Client) SendPushNotification(
	ctx context.Context,
	endpointArn,
	message string,
) error {
	_, err := client.sns.Publish(ctx, &sns.PublishInput{
		Message:          aws.String(message),
		MessageStructure: aws.String("json"),
		TargetArn:        aws.String(endpointArn),
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (client *Client) CreateApplicationEndpoint(
	ctx context.Context,
	fcmToken string,
) (
	string,
	error,
) {
	input := &sns.CreatePlatformEndpointInput{
		PlatformApplicationArn: client.cfg.PlatformApplicationArn,
		Token:                  aws.String(fcmToken),
	}
	result, err := client.sns.CreatePlatformEndpoint(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create platform endpoint: %w", err)
	}
	if result.EndpointArn == nil {
		return "", fmt.Errorf("endpoint arn is nil")
	}

	return *result.EndpointArn, nil
}
