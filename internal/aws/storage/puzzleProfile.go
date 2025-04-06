package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrPuzzleProfileNotFound = fmt.Errorf("puzzle profile not found")

type PuzzleProfileUpdateOptions struct {
	Rating *float64
}

func (client *Client) GetPuzzleProfile(
	ctx context.Context,
	userId string,
) (
	entities.PuzzleProfile,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.PuzzleProfilesTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
	})
	if err != nil {
		return entities.PuzzleProfile{}, err
	}
	if output.Item == nil {
		return entities.PuzzleProfile{}, ErrPuzzleProfileNotFound
	}

	var puzzleProfile entities.PuzzleProfile
	err = attributevalue.UnmarshalMap(output.Item, &puzzleProfile)
	if err != nil {
		return entities.PuzzleProfile{}, err
	}

	return puzzleProfile, nil
}

func (client *Client) PutPuzzleProfile(
	ctx context.Context,
	puzzleProfile entities.PuzzleProfile,
) error {
	av, err := attributevalue.MarshalMap(puzzleProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal puzzle profile map")
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.PuzzleProfilesTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) UpdatePuzzleProfile(
	ctx context.Context,
	userId string,
	opts PuzzleProfileUpdateOptions,
) error {
	updateExpression := []string{}
	expressionAttributeValues := map[string]types.AttributeValue{}

	if opts.Rating != nil {
		updateExpression = append(updateExpression, "Rating = :rating")
		expressionAttributeValues[":rating"] = &types.AttributeValueMemberN{
			Value: strconv.FormatFloat(*opts.Rating, 'f', 1, 64),
		}
	}

	_, err := client.dynamodb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: client.cfg.PuzzleProfilesTableName,
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
