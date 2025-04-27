package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/yelaco/ludofy/internal/domains/entities"
)

func (client *Client) FetchMatchResults(
	ctx context.Context,
	userId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) ([]entities.MatchResult, map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.MatchResultsTableName,
		KeyConditionExpression: aws.String("UserId = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId},
		},
		ExclusiveStartKey: lastKey,
		ScanIndexForward:  aws.Bool(false), // desc: most recent first
		Limit:             aws.Int32(limit),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var matchResults []entities.MatchResult
	if err := attributevalue.UnmarshalListOfMaps(
		output.Items,
		&matchResults,
	); err != nil {
		return nil, nil, err
	}

	return matchResults, output.LastEvaluatedKey, nil
}

func (client *Client) PutMatchResult(
	ctx context.Context,
	matchResult entities.MatchResult,
	ttl time.Duration,
) error {
	av, err := attributevalue.MarshalMap(matchResult)
	if err != nil {
		return fmt.Errorf("failed to marshal match result map: %w", err)
	}
	av["TTL"] = &types.AttributeValueMemberN{
		Value: strconv.FormatInt(time.Now().Add(ttl).Unix(), 10),
	}
	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.MatchResultsTableName,
		Item:      av,
	})
	if err != nil {
		return err
	}
	return nil
}
