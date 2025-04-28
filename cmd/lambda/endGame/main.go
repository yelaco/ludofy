package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ts "github.com/mafredri/go-trueskill"
	"github.com/yelaco/ludofy/internal/aws/storage"
	"github.com/yelaco/ludofy/internal/domains/entities"
	"github.com/yelaco/ludofy/internal/ranking"
	"github.com/yelaco/ludofy/pkg/server"
)

var (
	storageClient   *storage.Client
	ratingAlgorithm = os.Getenv("RATING_ALGORITHM")
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handler(ctx context.Context, event json.RawMessage) error {
	var matchRecordReq server.MatchRecordRequest
	if err := json.Unmarshal(event, &matchRecordReq); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}
	matchRecord := server.MatchRecordRequestToEntity(matchRecordReq)

	err := storageClient.DeleteActiveMatch(ctx, matchRecord.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete active match: %w", err)
	}
	err = storageClient.PutMatchRecord(ctx, matchRecord)
	if err != nil {
		return fmt.Errorf("failed to put match record: %w", err)
	}

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

	err = storageClient.DeleteSpectatorConversation(ctx, matchRecord.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete spectator conversation: %w", err)
	}

	switch ratingAlgorithm {
	case "glicko":
		userRatings := make([]entities.UserRating, 0, len(matchRecord.Players))
		for _, player := range matchRecord.Players {
			userRating, err := storageClient.GetUserRating(ctx, player.GetPlayerId())
			if err != nil {
				return fmt.Errorf(
					"failed to get user rating: [userId: %s] - %w",
					player.GetPlayerId(),
					err,
				)
			}
			userRatings = append(userRatings, userRating)
		}
		if len(userRatings) != 2 {
			return fmt.Errorf("expect 2 players for glicko ranking system")
		}

		newUserRatings := make([]entities.UserRating, 0, len(userRatings))
		for i, userRating := range userRatings {
			opponentRating := userRatings[1-i]
			matchResults, _, err := storageClient.FetchMatchResults(ctx, opponentRating.UserId, nil, 100)
			if err != nil {
				return fmt.Errorf("failed to fetch match results: %w", err)
			}

			opponentRatings := make([]entities.UserRating, 0, len(matchResults)+1)
			results := make([]float64, 0, len(matchResults)+1)
			for _, matchResult := range matchResults {
				opponentRatings = append(opponentRatings, entities.UserRating{
					UserId: matchResult.OpponentId,
					Rating: matchResult.OpponentRating,
					RD:     matchResult.OpponentRD,
				})

				results = append(results, matchResult.Result)
			}
			opponentRatings = append(opponentRatings, opponentRating)
			results = append(results, matchRecordReq.Players[i].GetResult())

			newRating, newRD := ranking.CalculateNewRating(userRating, opponentRatings, results)
			newUserRating := entities.UserRating{
				UserId:       userRating.UserId,
				PartitionKey: "UserRatings",
				Rating:       newRating,
				RD:           newRD,
			}
			err = storageClient.PutUserRating(ctx, newUserRating)
			if err != nil {
				return fmt.Errorf(
					"failed to put user rating: [userId: %s] - %w",
					userRating.UserId,
					err,
				)
			}
			newUserRatings = append(newUserRatings, newUserRating)
		}

		for i, userRating := range newUserRatings {
			playerMatchResult := entities.MatchResult{
				UserId:         userRating.UserId,
				MatchId:        matchRecordReq.MatchId,
				OpponentId:     newUserRatings[1-i].UserId,
				OpponentRating: newUserRatings[1-i].Rating,
				OpponentRD:     userRatings[1-i].Rating,
				Result:         matchRecord.Players[i].GetResult(),
				Timestamp:      matchRecordReq.EndedAt.Format(time.RFC3339Nano),
			}
			err = storageClient.PutMatchResult(ctx, playerMatchResult, 24*time.Hour)
			if err != nil {
				return fmt.Errorf(
					"failed to put user match result: [userId: %s] - %w",
					userRating.UserId,
					err,
				)
			}
		}
	case "trueskill":
		playerRecords := matchRecord.Players
		sort.Slice(playerRecords, func(i, j int) bool {
			return playerRecords[i].GetResult() > playerRecords[j].GetResult()
		})
		userRatings := make([]entities.UserRating, 0, len(matchRecord.Players))

		for _, player := range playerRecords {
			userRating, err := storageClient.GetUserRating(ctx, player.GetPlayerId())
			if err != nil {
				return fmt.Errorf(
					"failed to get user rating: [userId: %s] - %w",
					player.GetPlayerId(),
					err,
				)
			}
			userRatings = append(userRatings, userRating)
		}

		tsCfg := ts.New(ts.DrawProbabilityZero())
		players := make([]ts.Player, 0, len(userRatings))
		for _, userRating := range userRatings {
			players = append(players, ts.NewPlayer(userRating.Rating, userRating.Sigma))
		}
		draw := false
		newRatings, _ := tsCfg.AdjustSkills(players, draw)

		for i, newRating := range newRatings {
			err = storageClient.PutUserRating(ctx, entities.UserRating{
				UserId:       userRatings[i].UserId,
				PartitionKey: "UserRatings",
				Rating:       newRating.Mu(),
				Sigma:        newRating.Sigma(),
			})
			if err != nil {
				return fmt.Errorf(
					"failed to put user rating: %w",
					err,
				)
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
