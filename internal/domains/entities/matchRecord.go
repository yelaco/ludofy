package entities

import "time"

type MatchRecord struct {
	MatchId   string                  `dynamodbav:"MatchId"`
	Players   []PlayerRecordInterface `dynamodbav:"Players"`
	StartedAt time.Time               `dynamodbav:"StartedAt"`
	EndedAt   time.Time               `dynamodbav:"EndedAt"`
	Result    interface{}             `dynamodbav:"Result"`
}

type PlayerRecordInterface interface {
	GetPlayerId() string
}

type PlayerRecord map[string]interface{}

func (pr PlayerRecord) GetPlayerId() string {
	playerId, ok := pr["PlayerId"]
	if ok {
		return playerId.(string)
	}
	return ""
}
