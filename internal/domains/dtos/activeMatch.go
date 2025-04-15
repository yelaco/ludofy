package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type ActiveMatchResponse struct {
	MatchId        string           `json:"matchId"`
	ConversationId string           `json:"conversationId,omitempty"`
	Players        []PlayerResponse `json:"players"`
	GameMode       string           `json:"gameMode"`
	Server         string           `json:"server,omitempty"`
	StartedAt      *time.Time       `json:"startedAt"`
	CreatedAt      time.Time        `json:"createdAt"`
}

type PlayerResponse struct {
	Id string
}

type ActiveMatchListResponse struct {
	Items         []ActiveMatchResponse     `json:"items"`
	NextPageToken *NextActiveMatchPageToken `json:"nextPageToken"`
}

type NextActiveMatchPageToken struct {
	CreatedAt string `json:"createdAt"`
}

func ActiveMatchResponseFromEntity(activeMatch entities.ActiveMatch) ActiveMatchResponse {
	resp := ActiveMatchResponse{
		MatchId:        activeMatch.MatchId,
		ConversationId: activeMatch.ConversationId,
		Players:        make([]PlayerResponse, 0, len(activeMatch.Players)),
		GameMode:       activeMatch.GameMode,
		Server:         activeMatch.Server,
		StartedAt:      activeMatch.StartedAt,
		CreatedAt:      activeMatch.CreatedAt,
	}
	for _, player := range activeMatch.Players {
		resp.Players = append(resp.Players, PlayerResponse{
			Id: player.Id,
		})
	}
	return resp
}

func ActiveMatchListResponseFromEntities(activeMatches []entities.ActiveMatch) ActiveMatchListResponse {
	activeMatchResponses := make([]ActiveMatchResponse, 0, len(activeMatches))
	for _, activeMatch := range activeMatches {
		activeMatchResponses = append(activeMatchResponses, ActiveMatchResponse{
			MatchId:   activeMatch.MatchId,
			Players:   make([]PlayerResponse, 0, len(activeMatch.Players)),
			GameMode:  activeMatch.GameMode,
			StartedAt: activeMatch.StartedAt,
			CreatedAt: activeMatch.CreatedAt,
		})
	}
	return ActiveMatchListResponse{
		Items: activeMatchResponses,
	}
}
