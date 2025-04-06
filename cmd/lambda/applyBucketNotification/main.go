package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// CFNResponse is used to send responses to CloudFormation
type CFNResponse struct {
	Status             string                 `json:"Status"`
	Reason             string                 `json:"Reason,omitempty"`
	PhysicalResourceID string                 `json:"PhysicalResourceId,omitempty"`
	StackID            string                 `json:"StackId,omitempty"`
	RequestID          string                 `json:"RequestId,omitempty"`
	LogicalID          string                 `json:"LogicalResourceId,omitempty"`
	NoEcho             bool                   `json:"NoEcho,omitempty"`
	Data               map[string]interface{} `json:"Data,omitempty"`
}

// Request is the expected event format from AWS CloudFormation
type Request struct {
	RequestType        string            `json:"RequestType"`
	ResponseURL        string            `json:"ResponseURL"`
	StackID            string            `json:"StackId"`
	RequestID          string            `json:"RequestId"`
	LogicalResourceID  string            `json:"LogicalResourceId"`
	PhysicalResourceID string            `json:"PhysicalResourceId,omitempty"`
	ResourceProperties map[string]string `json:"ResourceProperties"`
}

// configureBucketNotification sets up the S3 event notification
func configureBucketNotification(ctx context.Context, bucketName, functionARN, prefixValue, notificationID string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	notificationConfig := &s3.PutBucketNotificationConfigurationInput{
		Bucket: aws.String(bucketName),
		NotificationConfiguration: &types.NotificationConfiguration{
			LambdaFunctionConfigurations: []types.LambdaFunctionConfiguration{
				{
					Id:                aws.String(notificationID),
					LambdaFunctionArn: aws.String(functionARN),
					Events:            []types.Event{"s3:ObjectCreated:*"},
					Filter: &types.NotificationConfigurationFilter{
						Key: &types.S3KeyFilter{
							FilterRules: []types.FilterRule{
								{
									Name:  types.FilterRuleNamePrefix,
									Value: aws.String(prefixValue),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = s3Client.PutBucketNotificationConfiguration(ctx, notificationConfig)
	if err != nil {
		return fmt.Errorf("failed to configure S3 bucket notification: %w", err)
	}
	return nil
}

// sendResponse sends a response back to CloudFormation
func sendResponse(ctx context.Context, event Request, status string, reason error) error {
	resp := CFNResponse{
		Status:             status,
		StackID:            event.StackID,
		RequestID:          event.RequestID,
		LogicalID:          event.LogicalResourceID,
		PhysicalResourceID: event.ResourceProperties["S3Bucket"],
		Data:               map[string]interface{}{"Bucket": event.ResourceProperties["S3Bucket"]},
	}

	if reason != nil {
		resp.Reason = reason.Error()
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, event.ResponseURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("Error sending response to CloudFormation: %v", err)
	}
	return err
}

// handler processes the CloudFormation custom resource event
func handler(ctx context.Context, event Request) error {
	log.Printf("Received event: %+v\n", event)

	bucketName := event.ResourceProperties["S3Bucket"]
	functionARN := event.ResourceProperties["FunctionARN"]
	prefixValue := event.ResourceProperties["PrefixValue"]
	notificationID := event.ResourceProperties["NotificationId"]

	var err error
	if event.RequestType == "Create" || event.RequestType == "Update" {
		err = configureBucketNotification(ctx, bucketName, functionARN, prefixValue, notificationID)
	}

	// Always send a response to CloudFormation
	status := "SUCCESS"
	if err != nil {
		log.Printf("Error configuring S3 bucket notification: %v", err)
		status = "FAILED"
	}

	return sendResponse(ctx, event, status, err)
}

func main() {
	lambda.Start(handler)
}
