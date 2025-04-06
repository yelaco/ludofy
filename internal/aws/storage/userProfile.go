package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrUserProfileNotFound = fmt.Errorf("user profile not found")

type UserProfileUpdateOptions struct {
	Avatar     *string
	Membership *string
}

func (client *Client) GetUserProfile(
	ctx context.Context,
	userId string,
) (
	entities.UserProfile,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.UserProfilesTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
	})
	if err != nil {
		return entities.UserProfile{}, err
	}
	if output.Item == nil {
		return entities.UserProfile{}, ErrUserProfileNotFound
	}

	var userProfile entities.UserProfile
	err = attributevalue.UnmarshalMap(output.Item, &userProfile)
	if err != nil {
		return entities.UserProfile{}, err
	}

	return userProfile, nil
}

func (client *Client) PutUserProfile(
	ctx context.Context,
	userProfile entities.UserProfile,
) error {
	av, err := attributevalue.MarshalMap(userProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile map")
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.UserProfilesTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) UpdateUserProfile(
	ctx context.Context,
	userId string,
	opts UserProfileUpdateOptions,
) error {
	updateExpression := []string{}
	expressionAttributeValues := map[string]types.AttributeValue{}

	if opts.Avatar != nil {
		updateExpression = append(updateExpression, "Avatar = :avatar")
		expressionAttributeValues[":avatar"] = &types.AttributeValueMemberS{
			Value: *opts.Avatar,
		}
	}

	if opts.Membership != nil {
		updateExpression = append(updateExpression, "Membership = :membership")
		expressionAttributeValues[":membership"] = &types.AttributeValueMemberS{
			Value: *opts.Membership,
		}
	}

	_, err := client.dynamodb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: client.cfg.UserProfilesTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
		UpdateExpression:          aws.String("SET " + strings.Join(updateExpression, ", ")),
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		return err
	}

	return nil
}
