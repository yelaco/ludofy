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

var ErrMatchStateNotFound = fmt.Errorf("match state not found")

func (client *Client) FetchMatchStates(
	ctx context.Context,
	matchId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
	order bool,
) (
	[]entities.MatchState,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.MatchStatesTableName,
		IndexName:              aws.String("MatchIndex"),
		KeyConditionExpression: aws.String("MatchId = :matchId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":matchId": &types.AttributeValueMemberS{Value: matchId},
		},
		ExclusiveStartKey: lastKey,
		ScanIndexForward:  aws.Bool(order), // desc
		Limit:             aws.Int32(limit),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	matchStates := make([]entities.MatchState, 0, len(output.Items))
	for _, item := range output.Items {
		matchState := entities.MatchState{
			Move: make(entities.Move),
		}

		list := item["PlayerStates"].(*types.AttributeValueMemberL).Value
		playerStates := make([]entities.PlayeState, 0, len(list))
		if err := attributevalue.UnmarshalList(
			list,
			&playerStates,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal list: %w", err)
		}

		if err := attributevalue.UnmarshalMap(
			item["Move"].(*types.AttributeValueMemberM).Value,
			&matchState.Move,
		); err != nil {
			return nil, nil, err
		}

		item["PlayerStates"] = nil
		item["Move"] = nil

		err := attributevalue.UnmarshalMap(item, &matchState)
		if err != nil {
			return nil, nil, err
		}

		matchState.PlayerStates = make([]entities.PlayerStateInterface, 0, len(playerStates))
		for _, state := range playerStates {
			matchState.PlayerStates = append(matchState.PlayerStates, state)
		}

		matchStates = append(matchStates, matchState)
	}

	return matchStates, output.LastEvaluatedKey, nil
}

func (client *Client) PutMatchState(
	ctx context.Context,
	matchState entities.MatchState,
) error {
	av, err := attributevalue.MarshalMap(matchState)
	if err != nil {
		return fmt.Errorf("failed to marshal match state map: %w", err)
	}

	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.MatchStatesTableName,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put match state: %w", err)
	}
	return nil
}
