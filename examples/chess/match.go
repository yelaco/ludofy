package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/chess-vn/slchess/pkg/server"
)

type Match struct {
	server.Match
	cfg MatchConfig
}

type MatchConfig struct {
	MatchDuration      time.Duration
	ClockIncrement     time.Duration
	CancelTimeout      time.Duration
	DisconnectTimeout  time.Duration
	MaxLagForgivenTime time.Duration
}

type GameMode struct {
	Time      time.Duration
	Increment time.Duration
}

var gameModes = []string{
	"1+0", "1+1", "1+2", // Bullet
	"2+1", "2+2", // Bullet
	"3+0", "3+2", // Blitz
	"5+0", "5+3", "5+5", // Blitz
	"10+0", "10+5", // Rapid
	"15+10", "25+10", // Rapid
	"30+0", "45+15", "60+30", // Classical
}

func ValidateGameMode(gameMode string) error {
	if slices.Contains(gameModes, gameMode) {
		return nil
	}
	return fmt.Errorf("unknown game mode")
}

func ParseGameMode(gameMode string) (GameMode, error) {
	if err := ValidateGameMode(gameMode); err != nil {
		return GameMode{}, err
	}
	mode := strings.Split(gameMode, "+")
	initialTimePerPlayer, err := strconv.Atoi(mode[0])
	if err != nil {
		return GameMode{}, err
	}
	increment, err := strconv.Atoi(mode[1])
	if err != nil {
		return GameMode{}, err
	}
	return GameMode{
		Time:      time.Duration(initialTimePerPlayer) * time.Minute,
		Increment: time.Duration(increment) * time.Second,
	}, nil
}

func ConfigForGameMode(gameMode string) (MatchConfig, error) {
	gm, err := ParseGameMode(gameMode)
	if err != nil {
		return MatchConfig{}, err
	}
	return MatchConfig{
		MatchDuration:     gm.Time,
		ClockIncrement:    gm.Increment,
		CancelTimeout:     30 * time.Second,
		DisconnectTimeout: 120 * time.Second,
	}, nil
}

/*
 * Implement MatchHandler interface
 */
type MyMatchHandler struct{}

func (h *MyMatchHandler) OnPlayerJoin(
	match server.Match,
	player server.Player,
) {
	// do something
}

func (h *MyMatchHandler) OnPlayerLeave(
	match server.Match,
	player server.Player,
) {
	// do something
}

func (h *MyMatchHandler) OnPlayerSync(
	match server.Match,
	player server.Player,
) {
	// do something
}

func (h *MyMatchHandler) HandleMove(
	match server.Match,
	player server.Player,
	move server.Move,
) {
	// do something
}

func (h *MyMatchHandler) OnMatchSave(match server.Match) {
	// do something
}

func (h *MyMatchHandler) OnMatchEnd(match server.Match) {
	// do something
}

func (h *MyMatchHandler) OnMatchAbort(match server.Match) {
	// do something
}

func NewMatchHandler() server.MatchHandler {
	return &MyMatchHandler{}
}
