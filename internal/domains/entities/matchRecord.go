package entities

import "time"

type PlayerRecord struct {
	Id        string  `dynamodbav:"Id"`
	OldRating float64 `dynamodbav:"Rating"`
	NewRating float64 `dynamodbav:"NewRating"`
}

type MatchRecord struct {
	MatchId   string         `dynamodbav:"MatchId"`
	Players   []PlayerRecord `dynamodbav:"Players"`
	Pgn       string         `dynamodbav:"Pgn"`
	StartedAt time.Time      `dynamodbav:"StartedAt"`
	EndedAt   time.Time      `dynamodbav:"EndedAt"`
}
