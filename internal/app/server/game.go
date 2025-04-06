package server

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
)

type (
	Status      uint8
	Side        bool
	GameControl uint8
)

const (
	INIT Status = iota
	CONNECTED
	DISCONNECTED

	WHITE_SIDE Side = true
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

type drawOffer struct {
	Side      chess.Color
	Timestamp time.Time
}

type game struct {
	chess.Game
	customOutcome chess.Outcome
	drawOffer     *drawOffer
	moves         []move
}

func newGame() *game {
	g := chess.NewGame(
		chess.UseNotation(chess.UCINotation{}),
	)
	return &game{
		Game:      *g,
		drawOffer: nil,
		moves:     []move{},
	}
}

func restoreGame(gameState string) (*game, error) {
	withFen, err := chess.FEN(gameState)
	if err != nil {
		return nil, err
	}
	g := chess.NewGame(
		withFen,
		chess.UseNotation(chess.UCINotation{}),
	)
	return &game{
		Game:      *g,
		drawOffer: nil,
		moves:     []move{},
	}, nil
}

func (g *game) OfferDraw(side chess.Color) bool {
	fmt.Println(side)
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

func (g *game) DeclineDraw(side chess.Color) bool {
	if g.drawOffer != nil && g.drawOffer.Side != side {
		g.drawOffer = nil
		return true
	}
	return false
}

func (g *game) outOfTime(side Side) {
	if side == WHITE_SIDE {
		g.customOutcome = WHITE_OUT_OF_TIME
	} else {
		g.customOutcome = BLACK_OUT_OF_TIME
	}
}

func (g *game) disconnectTimeout(side Side) {
	if side == WHITE_SIDE {
		g.customOutcome = WHITE_DISCONNECT_TIMEOUT
	} else {
		g.customOutcome = BLACK_DISCONNECT_TIMEOUT
	}
}

func (g *game) drawByTimeout() {
	g.customOutcome = DRAW_BY_TIMEOUT
}

func (g *game) outcome() chess.Outcome {
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

func (g *game) method() string {
	switch g.customOutcome {
	case WHITE_OUT_OF_TIME, BLACK_OUT_OF_TIME:
		return "OUT_OF_TIME"
	case WHITE_DISCONNECT_TIMEOUT, BLACK_DISCONNECT_TIMEOUT:
		return "DISCONNECT_TIMEOUT"
	default:
		return g.Method().String()
	}
}

func (g *game) lastMove() move {
	if length := len(g.moves); length > 0 {
		return g.moves[length-1]
	}
	return move{}
}

func (g *game) move(move move) error {
	if err := g.MoveStr(move.uci); err != nil {
		return err
	}
	g.moves = append(g.moves, move)
	return nil
}

type move struct {
	playerId  string
	uci       string
	control   GameControl
	createdAt time.Time
}

type player struct {
	Id            string
	Rating        float64
	RD            float64
	NewRatings    []float64
	NewRDs        []float64
	Conn          *websocket.Conn
	Side          Side
	Status        Status
	Clock         time.Duration
	TurnStartedAt time.Time
}

func newPlayer(
	conn *websocket.Conn,
	playerId string,
	side Side,
	clock time.Duration,
	rating float64,
	rd float64,
	newRatings []float64,
	newRDs []float64,
) player {
	player := player{
		Id:         playerId,
		Rating:     rating,
		RD:         rd,
		NewRatings: newRatings,
		NewRDs:     newRDs,
		Conn:       conn,
		Side:       side,
		Status:     INIT,
		Clock:      clock,
	}
	return player
}

func (p *player) color() chess.Color {
	if p.Side == WHITE_SIDE {
		return chess.White
	}
	return chess.Black
}

func (p *player) updateClock(
	timeTaken time.Duration,
	lagForgiven time.Duration,
	increment time.Duration,
) {
	p.Clock = p.Clock - timeTaken + lagForgiven + increment
}

func (s Status) String() string {
	switch s {
	case INIT:
		return "INIT"
	case CONNECTED:
		return "CONNECTED"
	case DISCONNECTED:
		return "DISCONNECTED"
	default:
		return "UNKNOWN"
	}
}
