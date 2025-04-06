package entities

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
	for _, gm := range gameModes {
		if gameMode == gm {
			return nil
		}
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
