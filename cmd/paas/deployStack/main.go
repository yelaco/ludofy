package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type DeployInput struct {
	StackName      string `json:"stackName"`
	TemplateURL    string `json:"templateUrl"`
	ServerImageUri string `json:"serverImageUri"`
}

func handler(ctx context.Context, input DeployInput) (string, error) {
	// Load default AWS config (from Lambda environment)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	cfClient := cloudformation.NewFromConfig(cfg)

	// Check if the stack already exists
	exists, err := stackExists(ctx, cfClient, input.StackName)
	if err != nil {
		return "", err
	}

	if exists {
		// Update the stack
		_, err = cfClient.UpdateStack(ctx, &cloudformation.UpdateStackInput{
			StackName:   aws.String(input.StackName),
			TemplateURL: aws.String(input.TemplateURL),
			Capabilities: []cftypes.Capability{
				cftypes.CapabilityCapabilityIam,
				cftypes.CapabilityCapabilityNamedIam,
			},
			Parameters: []cftypes.Parameter{
				{
					ParameterKey:   aws.String("ServerImageUri"),
					ParameterValue: aws.String(input.ServerImageUri),
				},
			},
			RetainExceptOnCreate: aws.Bool(true),
			DisableRollback:      aws.Bool(false),
		})
		if err != nil {
			// If the error is "No updates are to be performed", it's OK
			if isNoUpdateError(err) {
				log.Println("No updates needed for stack.")
				return "No updates needed", nil
			}
			return "", fmt.Errorf("failed to update stack: %w", err)
		}
		log.Println("Stack update initiated successfully.")
		return "Stack update initiated", nil
	}

	// Create the stack
	_, err = cfClient.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:   aws.String(input.StackName),
		TemplateURL: aws.String(input.TemplateURL),
		Capabilities: []cftypes.Capability{
			cftypes.CapabilityCapabilityIam,
			cftypes.CapabilityCapabilityNamedIam,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create stack: %w", err)
	}
	log.Println("Stack creation initiated successfully.")
	return "Stack creation initiated", nil
}

func stackExists(ctx context.Context, cfClient *cloudformation.Client, stackName string) (bool, error) {
	_, err := cfClient.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		if isStackNotExistError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to describe stack: %w", err)
	}
	return true, nil
}

func isStackNotExistError(err error) bool {
	if err == nil {
		return false
	}
	// Check if the error message contains "does not exist"
	return strings.Contains(err.Error(), "does not exist")
}

func isNoUpdateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "No updates are to be performed")
}

func main() {
	lambda.Start(handler)
}
