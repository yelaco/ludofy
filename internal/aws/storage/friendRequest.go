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

func (client *Client) FetchReceivedFriendRequests(
	ctx context.Context,
	userId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.FriendRequest,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.FriendRequestsTableName,
		IndexName:              aws.String("ReceiverIndex"),
		KeyConditionExpression: aws.String("ReceiverId = :receiverId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":receiverId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
		ExclusiveStartKey: lastKey,
		Limit:             aws.Int32(limit),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var friendRequests []entities.FriendRequest
	err = attributevalue.UnmarshalListOfMaps(output.Items, &friendRequests)
	if err != nil {
		return nil, nil, err
	}

	return friendRequests, output.LastEvaluatedKey, nil
}

func (client *Client) FetchSentFriendRequests(
	ctx context.Context,
	userId string,
	lastKey map[string]types.AttributeValue,
	limit int32,
) (
	[]entities.FriendRequest,
	map[string]types.AttributeValue,
	error,
) {
	input := &dynamodb.QueryInput{
		TableName:              client.cfg.FriendRequestsTableName,
		KeyConditionExpression: aws.String("SenderId = :senderId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":senderId": &types.AttributeValueMemberS{Value: userId},
		},
		ExclusiveStartKey: lastKey,
		Limit:             aws.Int32(limit),
	}
	output, err := client.dynamodb.Query(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	var friendRequests []entities.FriendRequest
	err = attributevalue.UnmarshalListOfMaps(output.Items, &friendRequests)
	if err != nil {
		return nil, nil, err
	}
	return friendRequests, output.LastEvaluatedKey, nil
}

func (client *Client) PutFriendRequest(ctx context.Context, friendRequest entities.FriendRequest) error {
	av, err := attributevalue.MarshalMap(friendRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	_, err = client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.FriendRequestsTableName,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put friend request: %w", err)
	}
	return nil
}

func (client *Client) DeleteFriendRequest(ctx context.Context, userId, receiverId string) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.FriendRequestsTableName,
		Key: map[string]types.AttributeValue{
			"SenderId":   &types.AttributeValueMemberS{Value: userId},
			"ReceiverId": &types.AttributeValueMemberS{Value: receiverId},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
