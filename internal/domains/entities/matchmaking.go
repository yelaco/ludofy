package entities

import (
	"fmt"
)

type MatchmakingTicket struct {
	UserId     string  `dynamodbav:"UserId"`
	IsRanked   bool    `dynamodbav:"IsRanked"`
	UserRating float64 `dynamodbav:"UserRating"`
	MinRating  float64 `dynamodbav:"MinRating"`
	MaxRating  float64 `dynamodbav:"MaxRating"`
	GameMode   string  `dynamodbav:"GameMode"`
}

func (t *MatchmakingTicket) Validate() error {
	if t.UserRating < t.MinRating || t.UserRating > t.MaxRating {
		return fmt.Errorf("invalid rating range: %v-%v-%v", t.MinRating, t.UserRating, t.MaxRating)
	}

	return nil
}
