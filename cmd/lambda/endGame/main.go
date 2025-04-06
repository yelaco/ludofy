package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var storageClient *storage.Client

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handler(ctx context.Context, event json.RawMessage) error {
	var matchRecordReq dtos.MatchRecordRequest
	if err := json.Unmarshal(event, &matchRecordReq); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}
	matchRecord := dtos.MatchRecordRequestToEntity(matchRecordReq)

	for _, player := range matchRecordReq.Players {
		err := storageClient.DeleteUserMatch(ctx, player.Id)
		if err != nil {
			return fmt.Errorf(
				"failed to delete user match: [userId: %s] - %w",
				player.Id,
				err,
			)
		}
	}

	err := storageClient.DeleteActiveMatch(ctx, matchRecord.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete active match: %w", err)
	}

	err = storageClient.PutMatchRecord(ctx, matchRecord)
	if err != nil {
		return fmt.Errorf("failed to put match record: %w", err)
	}

	for i, player := range matchRecordReq.Players {
		playerRating := entities.UserRating{
			UserId:       player.Id,
			PartitionKey: "UserRatings",
			Rating:       player.NewRating,
			RD:           player.NewRD,
		}
		err = storageClient.PutUserRating(ctx, playerRating)
		if err != nil {
			return fmt.Errorf(
				"failed to put player rating: [userId: %s] - %w",
				player.Id,
				err,
			)
		}

		playerMatchResult := entities.MatchResult{
			UserId:         player.Id,
			MatchId:        matchRecordReq.MatchId,
			OpponentId:     matchRecordReq.Players[1-i].Id,
			OpponentRating: matchRecordReq.Players[1-i].OldRating,
			OpponentRD:     matchRecordReq.Players[1-i].OldRD,
			Result:         matchRecordReq.Results[i],
			Timestamp:      matchRecordReq.EndedAt.Format(time.RFC3339),
		}
		err = storageClient.PutMatchResult(ctx, playerMatchResult)
		if err != nil {
			return fmt.Errorf(
				"failed to put user match result: [userId: %s] - %w",
				player.Id,
				err,
			)
		}
	}

	err = storageClient.DeleteSpectatorConversation(ctx, matchRecord.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete spectator conversation: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
