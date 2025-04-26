package entities

import "time"

type MatchState struct {
	Id           string                 `dynamodbav:"Id"`
	MatchId      string                 `dynamodbav:"MatchId"`
	PlayerStates []PlayerStateInterface `dynamodbav:"PlayerStates"`
	GameState    interface{}            `dynamodbav:"GameState"`
	Move         MoveInterface          `dynamodbav:"Move"`
	Timestamp    time.Time              `dynamodbav:"Timestamp"`
}

type (
	PlayeState map[string]interface{}
	Move       map[string]interface{}
)

func (ps PlayeState) GetPlayerId() string {
	playerId, ok := ps["PlayerId"]
	if ok {
		return playerId.(string)
	}
	return ""
}

func (m Move) GetPlayerId() string {
	playerId, ok := m["PlayerId"]
	if ok {
		return playerId.(string)
	}
	return ""
}

type PlayerStateInterface interface {
	GetPlayerId() string
}

type MoveInterface interface {
	GetPlayerId() string
}
