package main

import (
	"time"

	"github.com/chess-vn/slchess/pkg/server"
)

type (
	Side        bool
	GameControl uint8
)

const (
	WHITE_SDIE Side = true
	BLACK_SIDE Side = false

	ABORT GameControl = iota
	RESIGN
	OFFER_DRAW
	DECLINE_DRAW
	NONE

	BLACK_OUT_OF_TIME        = "BLACK_OUT_OF_TIME"
	WHITE_OUT_OF_TIME        = "WHITE_OUT_OF_TIME"
	BLACK_DISCONNECT_TIMEOUT = "BLACK_DISCONNECT_TIMEOUT"
	WHITE_DISCONNECT_TIMEOUT = "WHITE_DISCONNECT_TIMEOUT"
	DRAW_BY_TIMEOUT          = "DRAW_BY_TIMEOUT"
)

type Player struct {
	server.Player

	// Custom fields
	Clock         time.Duration
	Side          Side
	TurnStartedAt time.Time
}
