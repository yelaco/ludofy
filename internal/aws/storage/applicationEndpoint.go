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

func (client *Client) FetchApplicationEndpoints(
	ctx context.Context,
	userId string,
) (
	[]entities.ApplicationEndpoint,
	error,
) {
	output, err := client.dynamodb.Query(ctx, &dynamodb.QueryInput{
		TableName:              client.cfg.ApplicationEndpointsTableName,
		KeyConditionExpression: aws.String("UserId = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query application endpoints: %w", err)
	}
	var endpoints []entities.ApplicationEndpoint
	err = attributevalue.UnmarshalListOfMaps(output.Items, &endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal list of maps: %w", err)
	}
	return endpoints, nil
}

func (client *Client) PutApplicationEndpoint(ctx context.Context, endpoint entities.ApplicationEndpoint) error {
	av, err := attributevalue.MarshalMap(endpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.ApplicationEndpointsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) DeleteApplicationEndpoint(ctx context.Context, userId, deviceToken string) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.ApplicationEndpointsTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
			"DeviceToken": &types.AttributeValueMemberS{
				Value: deviceToken,
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
