package main

import (
	"time"

	"github.com/yelaco/ludofy/pkg/server"
)

const (
	ABORT        = "ABORT"
	RESIGN       = "RESIGN"
	OFFER_DRAW   = "OFFER_DRAW"
	DECLINE_DRAW = "DECLINE_DRAW"
	NONE         = "NONE"
)

type Move struct {
	server.Move
	Uci       string
	Control   string
	CreatedAt time.Time
}

func NewMove(playerId string) *Move {
	return &Move{
		Move: server.NewDefaultMove(playerId),
	}
}
