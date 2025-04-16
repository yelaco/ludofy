package main

import (
	"fmt"

	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
	"github.com/chess-vn/slchess/pkg/server"
)

/*
 * Implement ServerHandler interface
 */
type MyServerHandler struct{}

func (h *MyServerHandler) OnMatchCreate(match entities.ActiveMatch) (server.Match, error) {
	cfg, err := ConfigForGameMode(match.GameMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	players := make(map[string]server.Player, len(match.Players))
	for i, player := range match.Players {
		players[player.Id] = &Player{
			Player: server.NewDefaultPlayer(player.Id, match.MatchId),
			Clock:  cfg.MatchDuration,
			Side:   i%2 == 0,
		}
	}
	return Match{
		Match: server.NewDefaultMatch(match.MatchId, players),
		cfg:   cfg,
	}, nil
}

func (h *MyServerHandler) OnMatchResume(
	match entities.ActiveMatch,
	currentState entities.MatchState,
) (server.Match, error) {
	return nil, nil
}

func (h *MyServerHandler) OnHandleMessage(
	playerId string,
	match server.Match,
	message []byte,
) error {
	return nil
}

func (h *MyServerHandler) OnHandleMatchEnd(
	record *dtos.MatchRecordRequest,
	match server.Match,
) error {
	return nil
}

func (h *MyServerHandler) OnHandleMatchSave(
	matchState *dtos.MatchStateRequest,
	match server.Match,
) error {
	return nil
}

func NewServerHandler() server.ServerHandler {
	return &MyServerHandler{}
}
