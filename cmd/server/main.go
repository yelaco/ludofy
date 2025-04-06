package main

import (
	"github.com/chess-vn/slchess/internal/app/server"
	"github.com/chess-vn/slchess/pkg/logging"
	"go.uber.org/zap"
)

func main() {
	logging.Fatal("Game server exited: ", zap.Error(
		server.NewServer().Start(),
	))
}
