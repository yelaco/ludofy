package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

func (client *Client) FetchMessages(
	ctx context.Context,
	conversationId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) ([]entities.Message, map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.MessagesTableName,
		IndexName:              aws.String("ConversationIndex"),
		KeyConditionExpression: aws.String("ConversationId = :conversationId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":conversationId": &types.AttributeValueMemberS{
				Value: conversationId,
			},
		},
		ExclusiveStartKey: lastKey,
		ScanIndexForward:  aws.Bool(false), // desc: most recent first
		Limit:             aws.Int32(limit),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var messages []entities.Message
	if err := attributevalue.UnmarshalListOfMaps(
		output.Items,
		&messages,
	); err != nil {
		return nil, nil, err
	}

	return messages, output.LastEvaluatedKey, nil
}
