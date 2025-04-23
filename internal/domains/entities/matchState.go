package entities

import "time"

type MatchState struct {
	Id           string        `dynamodbav:"Id"`
	MatchId      string        `dynamodbav:"MatchId"`
	PlayerStates []PlayerState `dynamodbav:"PlayerStates"`
	GameState    interface{}   `dynamodbav:"GameState"`
	Move         Move          `dynamodbav:"Move"`
	Timestamp    time.Time     `dynamodbav:"Timestamp"`
}

type PlayerState interface {
	GetPlayerId() string
}

type Move interface {
	GetPlayerId() string
}
