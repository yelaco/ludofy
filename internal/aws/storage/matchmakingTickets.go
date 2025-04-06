package storage

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

func (client *Client) ScanMatchmakingTickets(
	ctx context.Context,
	ticket entities.MatchmakingTicket,
) (
	[]entities.MatchmakingTicket,
	error,
) {
	filter := "MinRating >= :min AND MaxRating <= :max AND GameMode = :mode"
	output, err := client.dynamodb.Scan(ctx, &dynamodb.ScanInput{
		TableName:        client.cfg.MatchmakingTicketsTableName,
		FilterExpression: aws.String(filter),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":min": &types.AttributeValueMemberN{
				Value: strconv.Itoa(int(ticket.MinRating)),
			},
			":max": &types.AttributeValueMemberN{
				Value: strconv.Itoa(int(ticket.MaxRating)),
			},
			":mode": &types.AttributeValueMemberS{
				Value: ticket.GameMode,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var tickets []entities.MatchmakingTicket
	err = attributevalue.UnmarshalListOfMaps(output.Items, &tickets)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (client *Client) PutMatchmakingTickets(
	ctx context.Context,
	ticket entities.MatchmakingTicket,
) error {
	ticketAv, _ := attributevalue.MarshalMap(ticket)
	_, err := client.dynamodb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: client.cfg.MatchmakingTicketsTableName,
		Item:      ticketAv,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) DeleteMatchmakingTickets(
	ctx context.Context,
	userId string,
) error {
	_, err := client.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: client.cfg.MatchmakingTicketsTableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
