package dtos

import "github.com/chess-vn/slchess/internal/domains/entities"

type PuzzleProfileResponse struct {
	UserId string  `json:"userId"`
	Rating float64 `json:"rating"`
}

func PuzzleProfileResponseFromEntity(puzzleProfile entities.PuzzleProfile) PuzzleProfileResponse {
	return PuzzleProfileResponse{
		UserId: puzzleProfile.UserId,
		Rating: puzzleProfile.Rating,
	}
}
