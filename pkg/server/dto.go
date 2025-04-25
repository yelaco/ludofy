package server

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

type PlayerRecord map[string]string

func (pr PlayerRecord) GetPlayerId() string {
	return pr["PlayerId"]
}

func (pr PlayerRecord) ContainsPlayerId() bool {
	_, ok := pr["PlayerId"]
	return ok
}
