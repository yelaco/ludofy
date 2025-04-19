package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

var ErrDeploymentNotFound = fmt.Errorf("deployment not found")

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
