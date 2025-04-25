package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrMatchRecordNotFound = fmt.Errorf("match record not found")

func (client *Client) GetMatchRecord(
	ctx context.Context,
	matchId string,
) (
	entities.MatchRecord,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.MatchRecordsTableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
			},
		},
	})
	if err != nil {
		return entities.MatchRecord{}, err
	}
	if output.Item == nil {
		return entities.MatchRecord{}, ErrMatchRecordNotFound
	}
	players := output.Item["Players"].(*types.AttributeValueMemberL).Value
	playerRecords := []entities.PlayerRecord{}
	if err := attributevalue.UnmarshalList(players, &playerRecords); err != nil {
		return entities.MatchRecord{}, err
	}
	output.Item["Players"] = nil

	var matchRecord entities.MatchRecord
	if err := attributevalue.UnmarshalMap(output.Item, &matchRecord); err != nil {
		return entities.MatchRecord{}, err
	}

	matchRecord.Players = make([]entities.PlayerRecordInterface, 0, len(playerRecords))
	for _, record := range playerRecords {
		matchRecord.Players = append(matchRecord.Players, record)
	}

	return matchRecord, nil
}

func (client *Client) PutMatchRecord(
	ctx context.Context,
	matchRecord entities.MatchRecord,
) error {
	av, err := attributevalue.MarshalMap(matchRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal match record map: %w", err)
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.MatchRecordsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}
