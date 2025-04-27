package dtos

import "github.com/yelaco/ludofy/internal/domains/entities"

type MatchmakingRequest struct {
	MinRating float64 `json:"minRating"`
	MaxRating float64 `json:"maxRating"`
	GameMode  string  `json:"gameMode"`
	IsRanked  bool    `json:"isRanked"`
}

func MatchmakingRequestToEntity(userId string, req MatchmakingRequest) entities.MatchmakingTicket {
	return entities.MatchmakingTicket{
		UserId:    userId,
		MinRating: req.MinRating,
		MaxRating: req.MaxRating,
		GameMode:  req.GameMode,
		IsRanked:  req.IsRanked,
	}
}
