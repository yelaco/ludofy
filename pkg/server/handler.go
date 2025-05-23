package server

import (
	"github.com/yelaco/ludofy/internal/domains/dtos"
	"github.com/yelaco/ludofy/internal/domains/entities"
)

type MatchHandler interface {
	GetMatch() Match
	OnPlayerJoin(player Player) (bool, error)
	OnPlayerLeave(player Player) error
	OnPlayerSync(player Player) error
	HandleMove(player Player, move Move) error
	OnMatchSave() error
	OnMatchEnd() error
	OnMatchAbort() error
}

type ServerHandler interface {
	OnMatchCreate(activeMatch entities.ActiveMatch) (Match, error)
	OnMatchResume(activeMatch entities.ActiveMatch, currentState entities.MatchState) (Match, error)
	OnHandleMessage(playerId string, match MatchHandler, message []byte) error
	OnHandleMatchEnd(record *MatchRecordRequest, match MatchHandler) error
	OnHandleMatchSave(matchState *dtos.MatchStateRequest, match MatchHandler) error
}
