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

var ErrPlatformNotFound = fmt.Errorf("platform not found")

func (client *Client) GetPlatform(
	ctx context.Context,
	platformId string,
) (
	entities.Platform,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.PlatformsTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: platformId,
			},
		},
	})
	if err != nil {
		return entities.Platform{}, err
	}
	if output.Item == nil {
		return entities.Platform{}, ErrPlatformNotFound
	}
	var platform entities.Platform
	if err := attributevalue.UnmarshalMap(output.Item, &platform); err != nil {
		return entities.Platform{}, err
	}
	return platform, nil
}

func (client *Client) FetchPlatforms(
	ctx context.Context,
	userId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.Platform,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.PlatformsTableName,
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
	var platforms []entities.Platform
	err = attributevalue.UnmarshalListOfMaps(output.Items, &platforms)
	if err != nil {
		return nil, nil, err
	}
	return platforms, output.LastEvaluatedKey, nil
}

func (client *Client) PutPlatform(
	ctx context.Context,
	platform entities.Platform,
) error {
	av, err := attributevalue.MarshalMap(platform)
	if err != nil {
		return fmt.Errorf("failed to marshal match record map: %w", err)
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.PlatformsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}
