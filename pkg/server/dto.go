package server

import (
	"time"

	"github.com/yelaco/ludofy/internal/domains/entities"
)

type MatchRecordRequest struct {
	MatchId   string         `json:"matchId"`
	Players   []PlayerRecord `json:"players"`
	StartedAt time.Time      `json:"startedAt"`
	EndedAt   time.Time      `json:"endedAt"`
	Result    interface{}    `json:"results"`
}

func MatchRecordRequestToEntity(req MatchRecordRequest) entities.MatchRecord {
	matchRecord := entities.MatchRecord{
		MatchId:   req.MatchId,
		Players:   make([]entities.PlayerRecordInterface, 0, len(req.Players)),
		StartedAt: req.StartedAt,
		EndedAt:   req.EndedAt,
		Result:    req.Result,
	}
	for _, player := range req.Players {
		matchRecord.Players = append(matchRecord.Players, player)
	}
	return matchRecord
}

type PlayerRecord map[string]interface{}

func (pr PlayerRecord) GetPlayerId() string {
	id, ok := pr["PlayerId"]
	if ok {
		playerId, ok := id.(string)
		if ok {
			return playerId
		}
	}
	return ""
}

func (pr PlayerRecord) ContainsPlayerId() bool {
	_, ok := pr["PlayerId"]
	return ok
}

func (pr PlayerRecord) GetResult() float64 {
	result, ok := pr["Result"]
	if ok {
		return result.(float64)
	}
	return 0
}
