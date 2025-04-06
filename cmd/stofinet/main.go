package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/chess-vn/slchess/internal/app/stofinet"
	"github.com/chess-vn/slchess/pkg/logging"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan // Wait for a signal
		logging.Info("terminating")
		cancel() // Cancel the context
	}()

	client := stofinet.NewClient()

	err := client.Start(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			logging.Fatal("failed to start", zap.Error(client.Start(ctx)))
		}
	}
}
