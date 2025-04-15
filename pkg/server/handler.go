package server

import (
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

type MatchHandler interface {
	OnPlayerJoin(match *Match, player *Player)
	OnPlayerLeave(match *Match, player *Player)
	OnPlayerSync(match *Match, player *Player)
	HandleMove(match *Match, player *Player, move Move)
	OnMatchSave(match *Match)
	OnMatchEnd(match *Match)
	OnMatchAbort(match *Match)
}

type ServerHandler interface {
	OnMatchCreate(match entities.ActiveMatch) (*Match, error)
	OnMatchResume(match entities.ActiveMatch, currentState entities.MatchState) (*Match, error)
	OnHandleMessage(playerId string, match *Match, message []byte) error
	OnHandleMatchEnd(record *dtos.MatchRecordRequest, match *Match) error
	OnHandleMatchSave(matchState *dtos.MatchStateRequest, match *Match) error
}
