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
	"github.com/chess-vn/slchess/pkg/server"
)

type MatchRecordRequest struct {
	MatchId   string                `json:"matchId"`
	Players   []server.PlayerRecord `json:"players"`
	StartedAt time.Time             `json:"startedAt"`
	EndedAt   time.Time             `json:"endedAt"`
	Result    interface{}           `json:"results"`
}

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

	for _, player := range matchRecord.Players {
		err := storageClient.DeleteUserMatch(ctx, player.GetPlayerId())
		if err != nil {
			return fmt.Errorf(
				"failed to delete user match: [userId: %s] - %w",
				player.GetPlayerId(),
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

	for _, player := range matchRecordReq.Players {
		playerRating := entities.UserRating{
			UserId:       player.GetPlayerId(),
			PartitionKey: "UserRatings",
			Rating:       1200,
			RD:           200,
		}
		err = storageClient.PutUserRating(ctx, playerRating)
		if err != nil {
			return fmt.Errorf(
				"failed to put player rating: [userId: %s] - %w",
				player.GetPlayerId(),
				err,
			)
		}

		// playerMatchResult := entities.MatchResult{
		// 	UserId:         player.GetPlayerId(),
		// 	MatchId:        matchRecordReq.MatchId,
		// 	OpponentId:     matchRecordReq.Players[1-i].Id,
		// 	OpponentRating: matchRecordReq.Players[1-i].OldRating,
		// 	OpponentRD:     matchRecordReq.Players[1-i].OldRD,
		// 	Result:         matchRecordReq.Result[i],
		// 	Timestamp:      matchRecordReq.EndedAt.Format(time.RFC3339),
		// }
		// err = storageClient.PutMatchResult(ctx, playerMatchResult)
		// if err != nil {
		// 	return fmt.Errorf(
		// 		"failed to put user match result: [userId: %s] - %w",
		// 		player.Id,
		// 		err,
		// 	)
		// }
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
