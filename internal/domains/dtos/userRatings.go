package dtos

import "github.com/chess-vn/slchess/internal/domains/entities"

type UserRatingResponse struct {
	UserId string  `json:"userId"`
	Rating float64 `json:"rating"`
	RD     float64 `json:"rd,omitempty"`
}

type UserRatingListResponse struct {
	Items         []UserRatingResponse     `json:"items"`
	NextPageToken *NextUserRatingPageToken `json:"nextPageToken"`
}

type NextUserRatingPageToken struct {
	Rating string `json:"rating"`
}

func UserRatingListResponseFromEntities(userRatings []entities.UserRating) UserRatingListResponse {
	userRatingResponses := make([]UserRatingResponse, 0, len(userRatings))
	for _, userRating := range userRatings {
		userRatingResponses = append(userRatingResponses, UserRatingResponse{
			UserId: userRating.UserId,
			Rating: userRating.Rating,
		})
	}
	return UserRatingListResponse{
		Items: userRatingResponses,
	}
}
