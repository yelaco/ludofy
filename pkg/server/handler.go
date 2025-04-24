package server

import (
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

type MatchHandler interface {
	OnPlayerJoin(match Match, player Player) (bool, error)
	OnPlayerLeave(match Match, player Player) error
	OnPlayerSync(match Match, player Player) error
	HandleMove(match Match, player Player, move Move) error
	OnMatchSave(match Match) error
	OnMatchEnd(match Match) error
	OnMatchAbort(match Match) error
}

type ServerHandler interface {
	OnMatchCreate(activeMatch entities.ActiveMatch) (Match, error)
	OnMatchResume(activeMatch entities.ActiveMatch, currentState entities.MatchState) (Match, error)
	OnHandleMessage(playerId string, match Match, message []byte) error
	OnHandleMatchEnd(record *dtos.MatchRecordRequest, match Match) error
	OnHandleMatchSave(matchState *dtos.MatchStateRequest, match Match) error
}
