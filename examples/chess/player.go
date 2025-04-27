package main

import (
	"time"

	"github.com/notnil/chess"
	"github.com/yelaco/ludofy/pkg/server"
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
	if p.Side == WHITE_SIDE {
		return chess.White
	}
	return chess.Black
}
