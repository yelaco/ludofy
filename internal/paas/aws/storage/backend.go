package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

var ErrBackendNotFound = fmt.Errorf("backend not found")

func (client *Client) GetBackend(
	ctx context.Context,
	id string,
) (
	entities.Backend,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.BackendsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	})
	if err != nil {
		return entities.Backend{}, err
	}
	if output.Item == nil {
		return entities.Backend{}, ErrBackendNotFound
	}
	var backend entities.Backend
	if err := attributevalue.UnmarshalMap(output.Item, &backend); err != nil {
		return entities.Backend{}, err
	}
	return backend, nil
}

func (client *Client) CheckExistedBackendStack(
	ctx context.Context,
	userId string,
	stackName string,
) (
	bool,
	error,
) {
	output, err := client.dynamodb.Query(ctx, &dynamodb.QueryInput{
		TableName:              client.cfg.BackendsTableName,
		IndexName:              aws.String("UserIndex"),
		KeyConditionExpression: aws.String("UserId = :userId"),
		FilterExpression:       aws.String("StackName = :stackName"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId":    &types.AttributeValueMemberS{Value: userId},
			":stackName": &types.AttributeValueMemberS{Value: stackName},
		},
		ProjectionExpression: aws.String("Id"),
		Limit:                aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("failed to query backend: %w", err)
	}
	return len(output.Items) > 0, nil
}

func (client *Client) FetchBackends(
	ctx context.Context,
	userId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.Backend,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.BackendsTableName,
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
	var backends []entities.Backend
	err = attributevalue.UnmarshalListOfMaps(output.Items, &backends)
	if err != nil {
		return nil, nil, err
	}
	return backends, output.LastEvaluatedKey, nil
}

type BackendUpdateOptions struct {
	Status *string
}

func (client *Client) UpdateBackend(
	ctx context.Context,
	backendId string,
	opts BackendUpdateOptions,
) error {
	updateExpression := []string{"UpdatedAt = :updatedAt"}
	expressionAttributeValues := map[string]types.AttributeValue{
		":updatedAt": &types.AttributeValueMemberS{
			Value: time.Now().Format(time.RFC3339Nano),
		},
	}
	expressionAttributeNames := map[string]string{}

	if opts.Status != nil {
		updateExpression = append(updateExpression, "#stat = :stat")
		expressionAttributeValues[":stat"] = &types.AttributeValueMemberS{
			Value: *opts.Status,
		}
		expressionAttributeNames["#stat"] = "Status"
	}

	_, err := client.dynamodb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: client.cfg.BackendsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: backendId,
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

func (client *Client) PutBackend(
	ctx context.Context,
	backend entities.Backend,
) error {
	av, err := attributevalue.MarshalMap(backend)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.BackendsTableName,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

func (client *Client) DeleteBackend(
	ctx context.Context,
	backendId string,
) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.BackendsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: backendId,
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
