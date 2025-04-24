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

func (client *Client) CheckForActiveMatch(
	ctx context.Context,
	userId string,
) (
	entities.ActiveMatch,
	error,
) {
	userMatch, err := client.GetUserMatch(ctx, userId)
	if err != nil {
		return entities.ActiveMatch{}, err
	}
	activeMatch, err := client.GetActiveMatch(ctx, userMatch.MatchId)
	if err != nil {
		return entities.ActiveMatch{}, err
	}
	return activeMatch, nil
}

func (client *Client) TransactCreateMatch(ctx context.Context, match entities.ActiveMatch) error {
	transactItems := make([]types.TransactWriteItem, 0, len(match.Players)*2+1)
	for _, player := range match.Players {
		transactItems = append(transactItems, types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: client.cfg.MatchmakingTicketsTableName,
				Key: map[string]types.AttributeValue{
					"UserId": &types.AttributeValueMemberS{Value: player.Id},
				},
				ConditionExpression: aws.String("attribute_exists(UserId)"),
			},
		})
	}
	av, err := attributevalue.MarshalMap(match)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	transactItems = append(transactItems, types.TransactWriteItem{
		Put: &types.Put{
			TableName: client.cfg.ActiveMatchesTableName,
			Item:      av,
		},
	})
	for _, player := range match.Players {
		userMatch := entities.UserMatch{
			UserId:  player.Id,
			MatchId: match.MatchId,
		}
		av, err := attributevalue.MarshalMap(userMatch)
		if err != nil {
			return fmt.Errorf("failed to marshal map: %w", err)
		}
		transactItems = append(transactItems, types.TransactWriteItem{
			Put: &types.Put{
				TableName: client.cfg.UserMatchesTableName,
				Item:      av,
			},
		})
	}

	_, err = client.dynamodb.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	})
	if err != nil {
		return fmt.Errorf("failed to transact write items: %w", err)
	}

	return nil
}
