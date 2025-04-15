package entities

import "time"

type ActiveMatch struct {
	MatchId        string     `dynamodbav:"MatchId"`
	ConversationId string     `dynamodbav:"ConversationId"`
	PartitionKey   string     `dynamodbav:"PartitionKey"`
	Players        []Player   `dynamodbav:"Players"`
	GameMode       string     `dynamodbav:"GameMode"`
	Server         string     `dynamodbav:"Server"`
	StartedAt      *time.Time `dynamodbav:"StartedAt"`
	CreatedAt      time.Time  `dynamodbav:"CreatedAt"`
}

type Player struct {
	Id string `dynamodbav:"Id"`
}
