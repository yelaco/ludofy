package entities

import "time"

type PlayerState struct {
	Clock  string `dynamodbav:"Clock"`
	Status string `dynamodbav:"Status"`
}

type Move struct {
	PlayerId string `dynamodbav:"PlayerId"`
	Uci      string `dynamodbav:"Uci"`
}

type MatchState struct {
	Id           string        `dynamodbav:"Id"`
	MatchId      string        `dynamodbav:"MatchId"`
	PlayerStates []PlayerState `dynamodbav:"PlayerStates"`
	GameState    string        `dynamodbav:"GameState"`
	Move         Move          `dynamodbav:"Move"`
	Ply          int           `dynamodbav:"Ply"`
	Timestamp    time.Time     `dynamodbav:"Timestamp"`
}
