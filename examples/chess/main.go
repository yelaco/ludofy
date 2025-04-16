package main

import (
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/server"
	"go.uber.org/zap"
)

// Run server
func main() {
	serverHandler := NewServerHandler()
	matchHanndler := NewMatchHandler()
	cfg := server.NewConfig("7202", serverHandler, matchHanndler)
	srv := server.NewFromConfig(cfg)
	logging.Fatal("server runtime error", zap.Error(srv.Start()))
}
