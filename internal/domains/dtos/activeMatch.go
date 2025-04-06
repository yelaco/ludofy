package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type ActiveMatchResponse struct {
	MatchId        string         `json:"matchId"`
	ConversationId string         `json:"conversationId,omitempty"`
	Player1        PlayerResponse `json:"player1"`
	Player2        PlayerResponse `json:"player2"`
	GameMode       string         `json:"gameMode"`
	Server         string         `json:"server,omitempty"`
	StartedAt      *time.Time     `json:"startedAt"`
	CreatedAt      time.Time      `json:"createdAt"`
}

type PlayerResponse struct {
	Id         string    `json:"id"`
	Rating     float64   `json:"rating"`
	NewRatings []float64 `json:"newRatings,omitempty"`
}

type ActiveMatchListResponse struct {
	Items         []ActiveMatchResponse     `json:"items"`
	NextPageToken *NextActiveMatchPageToken `json:"nextPageToken"`
}

type NextActiveMatchPageToken struct {
	CreatedAt string `json:"createdAt"`
}

func ActiveMatchResponseFromEntity(activeMatch entities.ActiveMatch) ActiveMatchResponse {
	return ActiveMatchResponse{
		MatchId:        activeMatch.MatchId,
		ConversationId: activeMatch.ConversationId,
		Player1: PlayerResponse{
			Id:         activeMatch.Player1.Id,
			Rating:     activeMatch.Player1.Rating,
			NewRatings: activeMatch.Player1.NewRatings,
		},
		Player2: PlayerResponse{
			Id:         activeMatch.Player2.Id,
			Rating:     activeMatch.Player2.Rating,
			NewRatings: activeMatch.Player2.NewRatings,
		},
		GameMode:  activeMatch.GameMode,
		Server:    activeMatch.Server,
		StartedAt: activeMatch.StartedAt,
		CreatedAt: activeMatch.CreatedAt,
	}
}

func ActiveMatchListResponseFromEntities(activeMatches []entities.ActiveMatch) ActiveMatchListResponse {
	activeMatchResponses := make([]ActiveMatchResponse, 0, len(activeMatches))
	for _, activeMatch := range activeMatches {
		activeMatchResponses = append(activeMatchResponses, ActiveMatchResponse{
			MatchId: activeMatch.MatchId,
			Player1: PlayerResponse{
				Id:     activeMatch.Player1.Id,
				Rating: activeMatch.Player1.Rating,
			},
			Player2: PlayerResponse{
				Id:     activeMatch.Player2.Id,
				Rating: activeMatch.Player2.Rating,
			},
			GameMode:  activeMatch.GameMode,
			StartedAt: activeMatch.StartedAt,
			CreatedAt: activeMatch.CreatedAt,
		})
	}
	return ActiveMatchListResponse{
		Items: activeMatchResponses,
	}
}
