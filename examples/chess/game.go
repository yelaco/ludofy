package main

import (
	"time"

	"github.com/notnil/chess"
)

type (
	Side bool
)

const (
	WHITE_SIDE Side = true
	BLACK_SIDE Side = false

	BLACK_OUT_OF_TIME        = "BLACK_OUT_OF_TIME"
	WHITE_OUT_OF_TIME        = "WHITE_OUT_OF_TIME"
	BLACK_DISCONNECT_TIMEOUT = "BLACK_DISCONNECT_TIMEOUT"
	WHITE_DISCONNECT_TIMEOUT = "WHITE_DISCONNECT_TIMEOUT"
	DRAW_BY_TIMEOUT          = "DRAW_BY_TIMEOUT"

	DRAW_PENDING  = "pending"
	DRAW_DECLINED = "declined"
)

type drawOffer struct {
	Side      chess.Color
	Timestamp time.Time
}

type Game struct {
	chess.Game
	customOutcome chess.Outcome
	drawOffer     *drawOffer
	moves         []*Move
}

func NewGame() *Game {
	g := chess.NewGame(
		chess.UseNotation(chess.UCINotation{}),
	)
	return &Game{
		Game:      *g,
		drawOffer: nil,
		moves:     []*Move{},
	}
}

func RestoreGame(gameState string) (*Game, error) {
	withFen, err := chess.FEN(gameState)
	if err != nil {
		return nil, err
	}
	g := chess.NewGame(
		withFen,
		chess.UseNotation(chess.UCINotation{}),
	)
	return &Game{
		Game:      *g,
		drawOffer: nil,
		moves:     []*Move{},
	}, nil
}

func (g *Game) OfferDraw(side chess.Color) bool {
	if g.drawOffer != nil && g.drawOffer.Side != side &&
		time.Now().Before(g.drawOffer.Timestamp.Add(20*time.Second)) {
		g.Draw(chess.DrawOffer)
		return true
	}
	g.drawOffer = &drawOffer{
		Side:      side,
		Timestamp: time.Now(),
	}
	return false
}

func (g *Game) DeclineDraw(side chess.Color) bool {
	if g.drawOffer != nil && g.drawOffer.Side != side {
		g.drawOffer = nil
		return true
	}
	return false
}

func (g *Game) OutOfTime(side Side) {
	if side == WHITE_SIDE {
		g.customOutcome = WHITE_OUT_OF_TIME
	} else {
		g.customOutcome = BLACK_OUT_OF_TIME
	}
}

func (g *Game) DisconnectTimeout(side Side) {
	if side == WHITE_SIDE {
		g.customOutcome = WHITE_DISCONNECT_TIMEOUT
	} else {
		g.customOutcome = BLACK_DISCONNECT_TIMEOUT
	}
}

func (g *Game) DrawByTimeout() {
	g.customOutcome = DRAW_BY_TIMEOUT
}

func (g *Game) outcome() chess.Outcome {
	switch g.customOutcome {
	case BLACK_OUT_OF_TIME:
		return chess.WhiteWon
	case WHITE_OUT_OF_TIME:
		return chess.BlackWon
	case BLACK_DISCONNECT_TIMEOUT:
		return chess.WhiteWon
	case WHITE_DISCONNECT_TIMEOUT:
		return chess.BlackWon
	case DRAW_BY_TIMEOUT:
		return chess.Draw
	default:
		return g.Outcome()
	}
}

func (g *Game) method() string {
	switch g.customOutcome {
	case WHITE_OUT_OF_TIME, BLACK_OUT_OF_TIME:
		return "OUT_OF_TIME"
	case WHITE_DISCONNECT_TIMEOUT, BLACK_DISCONNECT_TIMEOUT:
		return "DISCONNECT_TIMEOUT"
	default:
		return g.Method().String()
	}
}

func (g *Game) lastMove() *Move {
	if length := len(g.moves); length > 0 {
		return g.moves[length-1]
	}
	return nil
}

func (g *Game) move(move *Move) error {
	if err := g.MoveStr(move.Uci); err != nil {
		return err
	}
	g.moves = append(g.moves, move)
	return nil
}
