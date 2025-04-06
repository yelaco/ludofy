package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrConnectionNotFound = fmt.Errorf("connection not found")

func (client *Client) PutConnection(ctx context.Context, connection entities.Connection) error {
	av, err := attributevalue.MarshalMap(connection)
	if err != nil {
		return fmt.Errorf("failed to marshal connection map: %w", err)
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.ConnectionsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) GetConnectionByUserId(
	ctx context.Context,
	userId string,
) (
	entities.Connection,
	error,
) {
	output, err := client.dynamodb.Query(ctx, &dynamodb.QueryInput{
		TableName:              client.cfg.ConnectionsTableName,
		IndexName:              aws.String("UserIdIndex"),
		KeyConditionExpression: aws.String("UserId = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return entities.Connection{},
			fmt.Errorf("failed to query connections: %w", err)
	}
	if len(output.Items) == 0 {
		return entities.Connection{}, ErrConnectionNotFound
	}

	var connection entities.Connection
	err = attributevalue.UnmarshalMap(output.Items[0], &connection)
	if err != nil {
		return entities.Connection{},
			fmt.Errorf("failed to unmarshal connection map: %w", err)
	}

	return connection, nil
}

func (client *Client) GetConnection(
	ctx context.Context,
	connectionId string,
) (
	entities.Connection,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.ConnectionsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: connectionId,
			},
		},
	})
	if err != nil {
		return entities.Connection{}, err
	}
	if output.Item == nil {
		return entities.Connection{}, ErrConnectionNotFound
	}
	var connection entities.Connection
	err = attributevalue.UnmarshalMap(output.Item, &connection)
	if err != nil {
		return entities.Connection{}, err
	}
	return connection, nil
}

func (client *Client) DeleteConnection(ctx context.Context, connectionId string) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.ConnectionsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: connectionId},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
