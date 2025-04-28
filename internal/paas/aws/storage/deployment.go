package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/yelaco/ludofy/internal/paas/domains/entities"
)

var ErrDeploymentNotFound = fmt.Errorf("deployment not found")

func (client *Client) CheckInProgressDeployment(
	ctx context.Context,
	backendId string,
) (
	bool,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.DeploymentsTableName,
		IndexName:              aws.String("BackendIndex"),
		KeyConditionExpression: aws.String("BackendId = :backendId"),
		FilterExpression:       aws.String("#stat = :pending OR #stat = :deploying"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":backendId": &types.AttributeValueMemberS{Value: backendId},
			":pending":   &types.AttributeValueMemberS{Value: "pending"},
			":deploying": &types.AttributeValueMemberS{Value: "deploying"},
		},
		ExpressionAttributeNames: map[string]string{
			"#stat": "Status",
		},
		ProjectionExpression: aws.String("Id"),
		ScanIndexForward:     aws.Bool(false),
		Limit:                aws.Int32(3),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return false, fmt.Errorf("failed to query deployments: %w", err)
	}
	if len(output.Items) > 0 {
		return true, nil
	}
	return false, nil
}

func (client *Client) GetLatestSuccessfulDeployment(
	ctx context.Context,
	backendId string,
) (
	entities.Deployment,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.DeploymentsTableName,
		IndexName:              aws.String("BackendIndex"),
		KeyConditionExpression: aws.String("BackendId = :backendId"),
		FilterExpression:       aws.String("#stat = :stat"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":backendId": &types.AttributeValueMemberS{Value: backendId},
			":stat":      &types.AttributeValueMemberS{Value: "successful"},
		},
		ExpressionAttributeNames: map[string]string{
			"#stat": "Status",
		},
		ScanIndexForward: aws.Bool(false),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return entities.Deployment{}, fmt.Errorf("failed to query deployments: %w", err)
	}
	if len(output.Items) > 0 {
		var deployment entities.Deployment
		if err := attributevalue.UnmarshalMap(output.Items[0], &deployment); err != nil {
			return entities.Deployment{}, fmt.Errorf("failed to marshal map: %w", err)
		}
		return deployment, nil
	}
	return entities.Deployment{}, ErrDeploymentNotFound
}

func (client *Client) GetDeployment(
	ctx context.Context,
	id string,
) (
	entities.Deployment,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.DeploymentsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	})
	if err != nil {
		return entities.Deployment{}, err
	}
	if output.Item == nil {
		return entities.Deployment{}, ErrDeploymentNotFound
	}
	var deployment entities.Deployment
	if err := attributevalue.UnmarshalMap(output.Item, &deployment); err != nil {
		return entities.Deployment{}, err
	}
	return deployment, nil
}

func (client *Client) FetchDeployments(
	ctx context.Context,
	userId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.Deployment,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.DeploymentsTableName,
		IndexName:              aws.String("UserIndex"),
		KeyConditionExpression: aws.String("UserId = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId},
		},
		ExclusiveStartKey: lastKey,
		ScanIndexForward:  aws.Bool(false),
		Limit:             aws.Int32(limit),
	}

	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var deployments []entities.Deployment
	err = attributevalue.UnmarshalListOfMaps(output.Items, &deployments)
	if err != nil {
		return nil, nil, err
	}
	return deployments, output.LastEvaluatedKey, nil
}

type DeploymentUpdateOptions struct {
	Status *string
}

func (client *Client) UpdateDeployment(
	ctx context.Context,
	deploymentId string,
	opts DeploymentUpdateOptions,
) error {
	updateExpression := []string{}
	expressionAttributeValues := map[string]types.AttributeValue{}
	expressionAttributeNames := map[string]string{}

	if opts.Status != nil {
		updateExpression = append(updateExpression, "#stat = :stat")
		expressionAttributeValues[":stat"] = &types.AttributeValueMemberS{
			Value: *opts.Status,
		}
		expressionAttributeNames["#stat"] = "Status"
	}

	_, err := client.dynamodb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: client.cfg.DeploymentsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: deploymentId,
			},
		},
		UpdateExpression:          aws.String("SET " + strings.Join(updateExpression, ", ")),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	})
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) PutDeployment(
	ctx context.Context,
	deployment entities.Deployment,
) error {
	av, err := attributevalue.MarshalMap(deployment)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.DeploymentsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}
