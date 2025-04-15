package main

import (
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/server"
	"go.uber.org/zap"
)

/*
 * Implement ServerHandler interface
 */
type MyServerHandler struct{}

func (h *MyServerHandler) OnMatchCreate(match entities.ActiveMatch) (*server.Match, error) {
	return nil, nil
}

func (h *MyServerHandler) OnMatchResume(
	match entities.ActiveMatch,
	currentState entities.MatchState,
) (*server.Match, error) {
	return nil, nil
}

func (h *MyServerHandler) OnHandleMessage(
	playerId string,
	match *server.Match,
	message []byte,
) error {
	return nil
}

func (h *MyServerHandler) OnHandleMatchEnd(
	record *dtos.MatchRecordRequest,
	match *server.Match,
) error {
	return nil
}

func (h *MyServerHandler) OnHandleMatchSave(
	matchState *dtos.MatchStateRequest,
	match *server.Match,
) error {
	return nil
}

func NewServerHandler() server.ServerHandler {
	return &MyServerHandler{}
}

/*
 * Implement MatchHandler interface
 */
type MyMatchHandler struct{}

func (h *MyMatchHandler) OnPlayerJoin(
	match *server.Match,
	player *server.Player,
) {
	// do something
}

func (h *MyMatchHandler) OnPlayerLeave(
	match *server.Match,
	player *server.Player,
) {
	// do something
}

func (h *MyMatchHandler) OnPlayerSync(
	match *server.Match,
	player *server.Player,
) {
	// do something
}

func (h *MyMatchHandler) HandleMove(
	match *server.Match,
	player *server.Player,
	move server.Move,
) {
	// do something
}

func (h *MyMatchHandler) OnMatchSave(match *server.Match) {
	// do something
}

func (h *MyMatchHandler) OnMatchEnd(match *server.Match) {
	// do something
}

func (h *MyMatchHandler) OnMatchAbort(match *server.Match) {
	// do something
}

func NewMatchHandler() server.MatchHandler {
	return &MyMatchHandler{}
}

// Run server
func main() {
	serverHandler := NewServerHandler()
	matchHanndler := NewMatchHandler()
	cfg := server.NewConfig("7202", serverHandler, matchHanndler)
	srv := server.NewFromConfig(cfg)
	logging.Fatal("server runtime error", zap.Error(srv.Start()))
}
