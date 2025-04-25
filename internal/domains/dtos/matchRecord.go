package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

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
