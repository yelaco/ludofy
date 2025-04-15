package entities

import "time"

type MatchRecord struct {
	MatchId   string         `dynamodbav:"MatchId"`
	Players   []PlayerRecord `dynamodbav:"Players"`
	StartedAt time.Time      `dynamodbav:"StartedAt"`
	EndedAt   time.Time      `dynamodbav:"EndedAt"`
	Result    interface{}    `dynamodbav:"Result"`
}

type PlayerRecord interface {
	GetPlayerId() string
}
