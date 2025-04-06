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

var ErrUserRatingNotFound = fmt.Errorf("user rating not found")

func (client *Client) GetUserRating(
	ctx context.Context,
	userId string,
) (
	entities.UserRating,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.UserRatingsTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
	})
	if err != nil {
		return entities.UserRating{}, err
	}
	var userRating entities.UserRating
	if err := attributevalue.UnmarshalMap(output.Item, &userRating); err != nil {
		return entities.UserRating{}, err
	}
	return userRating, nil
}

func (client *Client) FetchUserRatings(
	ctx context.Context,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.UserRating,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.UserRatingsTableName,
		IndexName:              aws.String("RatingIndex"),
		KeyConditionExpression: aws.String("#pk = :pk"),
		ExpressionAttributeNames: map[string]string{
			"#pk": "PartitionKey",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{
				Value: "UserRatings",
			},
		},
		ExclusiveStartKey: lastKey,
		ScanIndexForward:  aws.Bool(false),
		Limit:             aws.Int32(limit),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var userRatings []entities.UserRating
	err = attributevalue.UnmarshalListOfMaps(output.Items, &userRatings)
	if err != nil {
		return nil, nil, err
	}

	return userRatings, output.LastEvaluatedKey, nil
}

func (client *Client) PutUserRating(
	ctx context.Context,
	userRating entities.UserRating,
) error {
	av, err := attributevalue.MarshalMap(userRating)
	if err != nil {
		return fmt.Errorf("failed to marshal user rating map: %w", err)
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.UserRatingsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}
