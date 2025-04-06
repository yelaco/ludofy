package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrSpectatorConversationNotFound = fmt.Errorf("spectator conversation not found")

func (client *Client) GetSpectatorConversation(
	ctx context.Context,
	matchId string,
) (
	entities.SpectatorConversation,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.SpectatorConversationsTableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
			},
		},
	})
	if err != nil {
		return entities.SpectatorConversation{}, err
	}
	if output.Item == nil {
		return entities.SpectatorConversation{}, ErrSpectatorConversationNotFound
	}
	var spectatorConversation entities.SpectatorConversation
	err = attributevalue.UnmarshalMap(output.Item, &spectatorConversation)
	if err != nil {
		return entities.SpectatorConversation{}, err
	}
	return spectatorConversation, nil
}

func (client *Client) PutSpectatorConversation(
	ctx context.Context,
	spectatorConversation entities.SpectatorConversation,
) error {
	av, err := attributevalue.MarshalMap(spectatorConversation)
	if err != nil {
		return fmt.Errorf("failed to marshal spectator conversation map: %w", err)
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.SpectatorConversationsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) DeleteSpectatorConversation(
	ctx context.Context,
	matchId string,
) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.SpectatorConversationsTableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{Value: matchId},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
