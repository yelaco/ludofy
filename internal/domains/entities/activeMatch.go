package entities

import "time"

type ActiveMatch struct {
	MatchId        string     `dynamodbav:"MatchId"`
	ConversationId string     `dynamodbav:"ConversationId"`
	PartitionKey   string     `dynamodbav:"PartitionKey"`
	Player1        Player     `dynamodbav:"Player1"`
	Player2        Player     `dynamodbav:"Player2"`
	GameMode       string     `dynamodbav:"GameMode"`
	Server         string     `dynamodbav:"Server"`
	AverageRating  float64    `dynamodbav:"AverageRating"`
	StartedAt      *time.Time `dynamodbav:"StartedAt"`
	CreatedAt      time.Time  `dynamodbav:"CreatedAt"`
}

type Player struct {
	Id         string    `dynamodbav:"Id"`
	Rating     float64   `dynamodbav:"Rating"`
	RD         float64   `dynamodbav:"RD"`
	NewRatings []float64 `dynamodbav:"NewRatings"`
	NewRDs     []float64 `dynamodbav:"NewRDs"`
}
