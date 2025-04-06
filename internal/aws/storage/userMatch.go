package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var (
	ErrUserMatchNotFound       = fmt.Errorf("user match not found")
	ErrUserMatchAlreadyExisted = fmt.Errorf("user already in a match")
)

func (client *Client) GetUserMatch(
	ctx context.Context,
	userId string,
) (
	entities.UserMatch,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.UserMatchesTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return entities.UserMatch{}, err
	}
	if output.Item == nil {
		return entities.UserMatch{}, ErrUserMatchNotFound
	}

	var userMatch entities.UserMatch
	err = attributevalue.UnmarshalMap(output.Item, &userMatch)
	if err != nil {
		return entities.UserMatch{}, fmt.Errorf("failed to unmarshal user match map")
	}

	return userMatch, nil
}

func (client *Client) PutUserMatch(
	ctx context.Context,
	userMatch entities.UserMatch,
) error {
	av, err := attributevalue.MarshalMap(userMatch)
	if err != nil {
		return fmt.Errorf("failed to marshal user match map")
	}

	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           client.cfg.UserMatchesTableName,
		ConditionExpression: aws.String("attribute_not_exists(UserId)"),
		Item:                av,
	})
	if err != nil {
		var condCheckFailed *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckFailed) {
			return fmt.Errorf(
				"%w [userId: %s][matchId: %s]",
				ErrUserMatchAlreadyExisted,
				userMatch.UserId,
				userMatch.MatchId,
			)
		}
		return err
	}

	return nil
}

func (client *Client) DeleteUserMatch(
	ctx context.Context,
	userId string,
) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.UserMatchesTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
