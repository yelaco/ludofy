package main

import (
	"time"

	"github.com/chess-vn/slchess/pkg/server"
	"github.com/notnil/chess"
)

const (
	INIT         = "INIT"
	CONNECTED    = "CONNECTED"
	DISCONNECTED = "DISCONNECTED"
)

type Player struct {
	server.Player

	// Custom fields
	Clock         time.Duration
	Side          Side
	TurnStartedAt time.Time
}

func (p *Player) UpdateClock(
	timeTaken time.Duration,
	lagForgiven time.Duration,
	increment time.Duration,
) {
	p.Clock = p.Clock - timeTaken + lagForgiven + increment
}

func (p *Player) Color() chess.Color {
	if p.Side == WHITE_SDIE {
		return chess.White
	}
	return chess.Black
}
