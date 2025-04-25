package server

import (
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
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
	OnHandleMessage(playerId string, match Match, message []byte) error
	OnHandleMatchEnd(record *dtos.MatchRecordRequest, match Match) error
	OnHandleMatchSave(matchState *dtos.MatchStateRequest, match Match) error
}
