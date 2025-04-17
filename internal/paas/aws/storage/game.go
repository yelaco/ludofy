package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

var ErrGameNotFound = fmt.Errorf("game not found")

func (client *Client) GetGame(
	ctx context.Context,
	gameId string,
) (
	entities.Game,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.GamesTableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: gameId,
			},
		},
	})
	if err != nil {
		return entities.Game{}, err
	}
	if output.Item == nil {
		return entities.Game{}, ErrPlatformNotFound
	}
	var game entities.Game
	if err := attributevalue.UnmarshalMap(output.Item, &game); err != nil {
		return entities.Game{}, err
	}
	return game, nil
}

func (client *Client) FetchGames(
	ctx context.Context,
	platformId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.Game,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.GamesTableName,
		IndexName:              aws.String("PlatformIndex"),
		KeyConditionExpression: aws.String("PlatformId = :platformId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":platformId": &types.AttributeValueMemberS{Value: platformId},
		},
		ExclusiveStartKey: lastKey,
		Limit:             aws.Int32(limit),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var games []entities.Game
	err = attributevalue.UnmarshalListOfMaps(output.Items, &games)
	if err != nil {
		return nil, nil, err
	}
	return games, output.LastEvaluatedKey, nil
}

func (client *Client) PutGames(
	ctx context.Context,
	games []entities.Game,
) error {
	var writeRequests []types.WriteRequest
	for _, game := range games {
		avMap, err := attributevalue.MarshalMap(game)
		if err != nil {
			log.Fatalf("failed to marshal game: %v", err)
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: avMap,
			},
		})
	}
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			*client.cfg.GamesTableName: writeRequests,
		},
	}

	_, err := client.dynamodb.BatchWriteItem(ctx, input)
	if err != nil {
		log.Fatalf("failed to batch write items: %v", err)
	}

	return nil
}
