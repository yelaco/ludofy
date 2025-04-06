package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type MatchRecordRequest struct {
	MatchId   string                `json:"matchId"`
	Players   []PlayerRecordRequest `json:"players"`
	Pgn       string                `json:"pgn"`
	StartedAt time.Time             `json:"startedAt"`
	EndedAt   time.Time             `json:"endedAt"`
	Results   []float64             `json:"results"`
}

type PlayerRecordRequest struct {
	Id        string  `json:"id"`
	OldRating float64 `json:"rating"`
	NewRating float64 `json:"newRating"`
	OldRD     float64 `json:"oldRD"`
	NewRD     float64 `json:"newRD"`
}

type PlayerRecordGetResponse struct {
	Id        string  `json:"id"`
	OldRating float64 `json:"oldRating"`
	NewRating float64 `json:"newRating"`
}

type MatchRecordGetResponse struct {
	MatchId   string                    `json:"matchId"`
	Players   []PlayerRecordGetResponse `json:"players"`
	Pgn       string                    `json:"pgn"`
	StartedAt time.Time                 `json:"startedAt"`
	EndedAt   time.Time                 `json:"endedAt"`
}

func MatchRecordRequestToEntity(req MatchRecordRequest) entities.MatchRecord {
	return entities.MatchRecord{
		MatchId: req.MatchId,
		Players: []entities.PlayerRecord{
			{
				Id:        req.Players[0].Id,
				OldRating: req.Players[0].OldRating,
				NewRating: req.Players[0].NewRating,
			},
			{
				Id:        req.Players[1].Id,
				OldRating: req.Players[1].OldRating,
				NewRating: req.Players[1].NewRating,
			},
		},
		Pgn:       req.Pgn,
		StartedAt: req.StartedAt,
		EndedAt:   req.EndedAt,
	}
}

func MatchRecordGetResponseFromEntity(matchRecord entities.MatchRecord) MatchRecordGetResponse {
	return MatchRecordGetResponse{
		MatchId: matchRecord.MatchId,
		Players: []PlayerRecordGetResponse{
			{
				Id:        matchRecord.Players[0].Id,
				OldRating: matchRecord.Players[0].OldRating,
				NewRating: matchRecord.Players[0].NewRating,
			},
			{
				Id:        matchRecord.Players[1].Id,
				OldRating: matchRecord.Players[1].OldRating,
				NewRating: matchRecord.Players[1].NewRating,
			},
		},
		Pgn:       matchRecord.Pgn,
		StartedAt: matchRecord.StartedAt,
		EndedAt:   matchRecord.EndedAt,
	}
}
