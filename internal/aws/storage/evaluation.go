package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrEvaluationNotFound = fmt.Errorf("evaluation not found")

func (client *Client) GetEvaluation(
	ctx context.Context,
	fen string,
) (
	entities.Evaluation,
	error,
) {
	output, err := client.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: client.cfg.EvaluationsTableName,
		Key: map[string]types.AttributeValue{
			"Fen": &types.AttributeValueMemberS{
				Value: fen,
			},
		},
	})
	if err != nil {
		return entities.Evaluation{}, err
	}
	if output.Item == nil {
		return entities.Evaluation{}, ErrEvaluationNotFound
	}
	var eval entities.Evaluation
	if err := attributevalue.UnmarshalMap(output.Item, &eval); err != nil {
		return entities.Evaluation{}, err
	}
	return eval, nil
}

func (client *Client) PutEvaluation(
	ctx context.Context,
	eval entities.Evaluation,
	ttl time.Duration,
) error {
	av, err := attributevalue.MarshalMap(eval)
	if err != nil {
		return fmt.Errorf("failed to marshal evaluation map: %w", err)
	}
	av["TTL"] = &types.AttributeValueMemberN{
		Value: strconv.FormatInt(time.Now().Add(ttl).Unix(), 10),
	}

	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.EvaluationsTableName,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put evaluation: %w", err)
	}
	return nil
}
