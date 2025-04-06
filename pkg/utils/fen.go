package utils

import (
	"errors"
	"strconv"
	"strings"
)

// FEN represents the parsed data from a FEN string.
type FEN struct {
	Board           [8][8]string
	ActiveColor     string
	CastlingRights  string
	EnPassantTarget string
	HalfmoveClock   int
	FullmoveNumber  int
}

// ParseFEN parses a FEN string and returns a FEN struct.
func ParseFEN(fen string) (*FEN, error) {
	parts := strings.Split(fen, " ")
	if len(parts) != 6 {
		return nil, errors.New("invalid FEN: must have 6 fields")
	}

	board, err := parseBoard(parts[0])
	if err != nil {
		return nil, err
	}

	halfmoveClock, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, errors.New("invalid halfmove clock")
	}

	fullmoveNumber, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, errors.New("invalid fullmove number")
	}

	return &FEN{
		Board:           board,
		ActiveColor:     parts[1],
		CastlingRights:  parts[2],
		EnPassantTarget: parts[3],
		HalfmoveClock:   halfmoveClock,
		FullmoveNumber:  fullmoveNumber,
	}, nil
}

// parseBoard converts the FEN board string into a 2D array representation.
func parseBoard(boardStr string) ([8][8]string, error) {
	var board [8][8]string
	ranks := strings.Split(boardStr, "/")
	if len(ranks) != 8 {
		return board, errors.New("invalid board: must have 8 ranks")
	}

	for i, rank := range ranks {
		file := 0
		for _, char := range rank {
			if file >= 8 {
				return board, errors.New("invalid rank length")
			}
			if char >= '1' && char <= '8' {
				emptySquares := int(char - '0')
				for j := 0; j < emptySquares; j++ {
					board[i][file] = "."
					file++
				}
			} else {
				board[i][file] = string(char)
				file++
			}
		}
		if file != 8 {
			return board, errors.New("incomplete rank")
		}
	}
	return board, nil
}
