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
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrActiveMatchNotFound = fmt.Errorf("active match not found")

type ActiveMatchUpdateOptions struct {
	Server    *string
	StartedAt *time.Time
}

func (client *Client) GetActiveMatch(
	ctx context.Context,
	matchId string,
) (
	entities.ActiveMatch,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.ActiveMatchesTableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{Value: matchId},
		},
		ConsistentRead: aws.Bool(true),
	},
	)
	if err != nil {
		return entities.ActiveMatch{}, err
	}
	if output.Item == nil {
		return entities.ActiveMatch{}, ErrActiveMatchNotFound
	}

	var activeMatch entities.ActiveMatch
	err = attributevalue.UnmarshalMap(output.Item, &activeMatch)
	if err != nil {
		return entities.ActiveMatch{}, err
	}
	return activeMatch, nil
}

func (client *Client) FetchActiveMatches(
	ctx context.Context,
	gameMode string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.ActiveMatch,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.ActiveMatchesTableName,
		IndexName:              aws.String("AverageRatingIndex"),
		KeyConditionExpression: aws.String("#pk = :pk AND #rating >= :rating"),
		ExpressionAttributeNames: map[string]string{
			"#pk":     "PartitionKey",
			"#rating": "AverageRating",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "ActiveMatches"},
			":rating": &types.AttributeValueMemberN{Value: "1600.0"},
		},
		ExclusiveStartKey: lastKey,
		ScanIndexForward:  aws.Bool(false),
		Limit:             aws.Int32(limit),
	}
	if gameMode != "" {
		input.FilterExpression = aws.String("#gameMode = :gameMode")
		input.ExpressionAttributeNames["#gameMode"] = "GameMode"
		input.ExpressionAttributeValues[":gameMode"] = &types.AttributeValueMemberS{
			Value: gameMode,
		}
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var activeMatches []entities.ActiveMatch
	err = attributevalue.UnmarshalListOfMaps(
		output.Items,
		&activeMatches,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal active match maps: %w", err)
	}

	return activeMatches, output.LastEvaluatedKey, nil
}

func (client *Client) PutActiveMatch(
	ctx context.Context,
	activeMatch entities.ActiveMatch,
) error {
	av, err := attributevalue.MarshalMap(activeMatch)
	if err != nil {
		return fmt.Errorf("failed to marshal active match map: %w", err)
	}

	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.ActiveMatchesTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) UpdateActiveMatch(
	ctx context.Context,
	matchId string,
	opts ActiveMatchUpdateOptions,
) error {
	updateExpression := []string{}
	expressionAttributeValues := map[string]types.AttributeValue{}

	if opts.Server != nil {
		updateExpression = append(updateExpression, "Server = :server")
		expressionAttributeValues[":server"] = &types.AttributeValueMemberS{
			Value: *opts.Server,
		}
	}

	if opts.StartedAt != nil {
		updateExpression = append(updateExpression, "StartedAt = :startedAt")
		expressionAttributeValues[":startedAt"] = &types.AttributeValueMemberS{
			Value: opts.StartedAt.Format(time.RFC3339),
		}
	}

	_, err := client.dynamodb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: client.cfg.ActiveMatchesTableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
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

func (client *Client) DeleteActiveMatch(
	ctx context.Context,
	matchId string,
) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.ActiveMatchesTableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
