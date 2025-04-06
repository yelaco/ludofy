package dtos

import "github.com/chess-vn/slchess/internal/domains/entities"

type MatchmakingRequest struct {
	MinRating float64 `json:"minRating"`
	MaxRating float64 `json:"maxRating"`
	GameMode  string  `json:"gameMode"`
}

func MatchmakingRequestToEntity(userRating entities.UserRating, req MatchmakingRequest) entities.MatchmakingTicket {
	return entities.MatchmakingTicket{
		UserId:     userRating.UserId,
		UserRating: userRating.Rating,
		MinRating:  req.MinRating,
		MaxRating:  req.MaxRating,
		GameMode:   req.GameMode,
	}
}
