package dtos

import (
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

type GameResponse struct {
	Id         string `json:"id"`
	PlatformId string `json:"platformId"`
}

type GameListResponse struct {
	Items         []GameResponse     `json:"items"`
	NextPageToken *NextGamePageToken `json:"nextPageToken"`
}

type NextGamePageToken struct{}

func GameListResponseFromEntities(games []entities.Game) GameListResponse {
	gameList := []GameResponse{}
	for _, game := range games {
		gameList = append(gameList, GameResponseFromEntity(game))
	}
	return GameListResponse{
		Items: gameList,
	}
}

func GameResponseFromEntity(game entities.Game) GameResponse {
	resp := GameResponse{
		Id:         game.Id,
		PlatformId: game.PlatformId,
	}
	return resp
}
