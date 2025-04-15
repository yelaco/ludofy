package server

import "errors"

var (
	ErrStatusInvalidMove     string = "INVALID_MOVE"
	ErrStatusInvalidPlayerId string = "INVALID_PLAYER_ID"
	ErrStatusWrongTurn       string = "WRONG_TURN"
	ErrStatusAbortInvalidPly string = "INVALID_PLY"
)

var (
	ErrFailedToLoadMatch = errors.New("failed to load match")
	ErrInvalidOutcome    = errors.New("invalid outcome")
)
