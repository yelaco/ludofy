package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type MatchRecordRequest struct {
	MatchId   string                `json:"matchId"`
	Players   []PlayerRecordRequest `json:"players"`
	StartedAt time.Time             `json:"startedAt"`
	EndedAt   time.Time             `json:"endedAt"`
	Result    interface{}           `json:"results"`
}

type PlayerRecordRequest interface {
	GetPlayerId() string
}

type MatchRecordGetResponse struct {
	MatchId   string                 `json:"matchId"`
	Players   []PlayerRecordResponse `json:"players"`
	StartedAt time.Time              `json:"startedAt"`
	EndedAt   time.Time              `json:"endedAt"`
	Result    interface{}            `json:"result"`
}

type PlayerRecordResponse interface {
	GetPlayerId() string
}

func MatchRecordRequestToEntity(req MatchRecordRequest) entities.MatchRecord {
	matchRecord := entities.MatchRecord{
		MatchId:   req.MatchId,
		Players:   make([]entities.PlayerRecord, 0, len(req.Players)),
		StartedAt: req.StartedAt,
		EndedAt:   req.EndedAt,
		Result:    req.Result,
	}
	for _, player := range req.Players {
		matchRecord.Players = append(matchRecord.Players, player)
	}
	return matchRecord
}

func MatchRecordGetResponseFromEntity(matchRecord entities.MatchRecord) MatchRecordGetResponse {
	resp := MatchRecordGetResponse{
		MatchId:   matchRecord.MatchId,
		Players:   make([]PlayerRecordResponse, 0, len(matchRecord.Players)),
		StartedAt: matchRecord.StartedAt,
		EndedAt:   matchRecord.EndedAt,
		Result:    matchRecord.Result,
	}
	for _, playerRecord := range matchRecord.Players {
		resp.Players = append(resp.Players, playerRecord)
	}
	return resp
}
