package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/notnil/chess"
	"github.com/yelaco/ludofy/pkg/logging"
	"github.com/yelaco/ludofy/pkg/server"
	"go.uber.org/zap"
)

type Match struct {
	server.Match
	cfg   MatchConfig
	game  *Game
	timer *time.Timer

	StartedAt time.Time
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

type matchResponse struct {
	Type      string            `json:"type"`
	GameState gameStateResponse `json:"game"`
}

type gameStateResponse struct {
	Outcome      string                `json:"outcome"`
	Method       string                `json:"method"`
	Fen          string                `json:"fen"`
	PlayerStates []playerStateResponse `json:"playerStates"`
}

type playerStateResponse struct {
	Id     string `json:"id"`
	Clock  string `json:"clocks"`
	Status string `json:"status"`
}

type drawOfferResponse struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type errorResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
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

// setTimer method    set the timer to the specified duration before trigger end game handler
func (m *Match) setTimer(d time.Duration) {
	if m.timer != nil {
		m.timer.Reset(d)
		logging.Info(
			"clock reset",
			zap.String("match_id", m.GetId()),
			zap.String("duration", d.String()),
		)
		return
	}
	m.timer = time.NewTimer(d)
	go func() {
		<-m.timer.C
		m.End()
	}()
	logging.Info(
		"clock set",
		zap.String("match_id", m.GetId()),
		zap.String("duration", d.String()),
	)
}

// skipTimer method    skips timer by set timer to 0 duration timeout
func (m *Match) skipTimer() {
	if m.timer == nil {
		logging.Info("clock nil", zap.String("match_id", m.GetId()))
		return
	}
	m.timer.Reset(0)
	logging.Info("clock skipped", zap.String("match_id", m.GetId()))
}

func (m *Match) currentPly() int {
	return len(m.game.moves)
}

func (m *Match) getCurrentTurnPlayer() *Player {
	side := BLACK_SIDE
	if m.game.Position().Turn() == chess.White {
		side = WHITE_SIDE
	}
	for _, player := range m.GetPlayers() {
		if player.(*Player).Side == side {
			return player.(*Player)
		}
	}

	return nil
}

func (m *Match) calculateLagForgiven(moveCreatedAt time.Time) time.Duration {
	lagTime := time.Since(moveCreatedAt)
	if lagTime > m.cfg.MaxLagForgivenTime {
		return m.cfg.MaxLagForgivenTime
	}
	return lagTime
}

func (m *Match) checkTimeout() {
	var players []*Player
	for _, player := range m.GetPlayers() {
		players = append(players, player.(*Player))
	}
	if players[0].GetStatus() == CONNECTED &&
		players[1].GetStatus() == CONNECTED {
		return
	}
	if players[0].GetStatus() == INIT ||
		players[1].GetStatus() == INIT {
		m.DisconnectPlayers("match cancelled", time.Now().Add(5*time.Second))
	}
	if players[0].GetStatus() == DISCONNECTED &&
		players[1].GetStatus() == CONNECTED {
		m.game.DisconnectTimeout(players[0].Side)
	} else if players[0].GetStatus() == CONNECTED &&
		players[1].GetStatus() == DISCONNECTED {
		m.game.DisconnectTimeout(players[1].Side)
	} else if players[0].GetStatus() == DISCONNECTED &&
		players[1].GetStatus() == DISCONNECTED {
		m.game.DrawByTimeout()
	}

	gameStateResp := gameStateResponse{
		Outcome:      m.game.outcome().String(),
		Method:       m.game.method(),
		Fen:          m.game.FEN(),
		PlayerStates: make([]playerStateResponse, len(m.GetPlayers())),
	}
	for _, player := range m.GetPlayers() {
		gameStateResp.PlayerStates = append(gameStateResp.PlayerStates, playerStateResponse{
			Id:     player.GetId(),
			Status: player.GetStatus(),
			Clock:  player.(*Player).Clock.String(),
		})
	}
	m.notifyPlayers(gameStateResp)
}

func (m *Match) notifyPlayers(resp gameStateResponse) {
	for _, player := range m.GetPlayers() {
		err := player.WriteJson(matchResponse{
			Type:      "gameState",
			GameState: resp,
		})
		if err != nil {
			logging.Error(
				"couldn't notify player",
				zap.String("player_id", player.GetId()),
				zap.Error(err),
			)
		}
	}
}

func (m *Match) sendDrawOfferNotification(sender *Player, status string) {
	for _, player := range m.GetPlayers() {
		if player.GetId() == sender.GetId() {
			continue
		}
		err := player.WriteJson(drawOfferResponse{
			Type:      "drawOffer",
			Status:    status,
			CreatedAt: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			logging.Error("couldn't send draw offer", zap.Error(err))
		}
	}
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
type MyMatchHandler struct {
	match *Match
}

func (h *MyMatchHandler) OnPlayerJoin(playerInterface server.Player) (bool, error) {
	match := h.GetMatch().(*Match)
	player := playerInterface.(*Player)
	if player.GetStatus() == INIT && player.Side == WHITE_SIDE {
		match.StartedAt = time.Now()
		player.TurnStartedAt = match.StartedAt
		match.setTimer(match.cfg.MatchDuration)
		return true, nil
	}
	return false, nil
}

func (h *MyMatchHandler) OnPlayerLeave(playerInterface server.Player) error {
	match := h.GetMatch().(*Match)
	currentClock := match.getCurrentTurnPlayer().Clock

	if !match.IsEnded() {
		allDisconnected := true
		for _, player := range match.GetPlayers() {
			if player.GetStatus() != DISCONNECTED {
				allDisconnected = false
			}
		}

		if allDisconnected {
			match.setTimer(currentClock)
		} else {
			if currentClock < match.cfg.DisconnectTimeout {
				match.setTimer(currentClock)
			} else {
				match.setTimer(match.cfg.DisconnectTimeout)
			}
		}
	}

	return nil
}

func (h *MyMatchHandler) OnPlayerSync(playerInterface server.Player) error {
	match := h.GetMatch().(*Match)
	player := playerInterface.(*Player)

	resp := matchResponse{
		Type: "gameState",
		GameState: gameStateResponse{
			Outcome:      match.game.outcome().String(),
			Method:       match.game.method(),
			Fen:          match.game.FEN(),
			PlayerStates: make([]playerStateResponse, 0, len(match.GetPlayers())),
		},
	}

	for _, player := range match.GetPlayers() {
		resp.GameState.PlayerStates = append(resp.GameState.PlayerStates, playerStateResponse{
			Id:     player.GetId(),
			Status: player.GetStatus(),
			Clock:  player.(*Player).Clock.String(),
		})
	}

	err := player.WriteJson(resp)
	if err != nil {
		return fmt.Errorf("couldn't sync player: %w", err)
	}
	return nil
}

func (h *MyMatchHandler) HandleMove(
	playerInterface server.Player,
	moveInterface server.Move,
) error {
	match := h.GetMatch().(*Match)
	player := playerInterface.(*Player)
	move := moveInterface.(*Move)
	switch move.Control {
	case ABORT:
		if match.currentPly() > 1 {
			player.WriteJson(errorResponse{
				Type:  "error",
				Error: "INVALID_PLY",
			})
		}
		match.Abort()
		return nil
	case RESIGN:
		match.game.Resign(player.Color())
	case OFFER_DRAW:
		draw := match.game.OfferDraw(player.Color())
		if !draw {
			match.sendDrawOfferNotification(player, DRAW_PENDING)
			return nil
		}
	case DECLINE_DRAW:
		shouldNotify := match.game.DeclineDraw(player.Color())
		if shouldNotify {
			match.sendDrawOfferNotification(player, DRAW_DECLINED)
		}
		return nil
	default:
		if expectedId := match.getCurrentTurnPlayer().GetId(); player.GetId() != expectedId {
			player.WriteJson(errorResponse{
				Type: "error",
				Error: fmt.Sprintf(
					"%s: want %s - got %s",
					"WRONG_TURN",
					expectedId,
					player.GetId(),
				),
			})
			return nil
		}
		err := match.game.move(move)
		if err != nil {
			player.WriteJson(errorResponse{
				Type:  "error",
				Error: "INVALID_MOVE",
			})
			return nil
		}

		// If making move, update clock
		timeTaken := time.Since(player.TurnStartedAt)
		lagForgiven := match.calculateLagForgiven(move.CreatedAt)
		player.UpdateClock(timeTaken, lagForgiven, match.cfg.ClockIncrement)

		// If clock runs out, end the game
		if player.Clock <= 0 {
			match.game.OutOfTime(player.Side)
			logging.Info("out of time", zap.String("player_id", player.GetId()))
		} else {
			// else next turn
			currentTurnPlayer := match.getCurrentTurnPlayer()
			currentTurnPlayer.TurnStartedAt = time.Now()
			match.setTimer(currentTurnPlayer.Clock)
			logging.Info(
				"new turn",
				zap.String("player_id", currentTurnPlayer.GetId()),
				zap.String("clock", currentTurnPlayer.Clock.String()),
			)
		}
	}

	gameStateResp := gameStateResponse{
		Outcome:      match.game.outcome().String(),
		Method:       match.game.method(),
		Fen:          match.game.FEN(),
		PlayerStates: make([]playerStateResponse, len(match.GetPlayers())),
	}
	for _, player := range match.GetPlayers() {
		gameStateResp.PlayerStates = append(gameStateResp.PlayerStates, playerStateResponse{
			Id:     player.GetId(),
			Status: player.GetStatus(),
			Clock:  player.(*Player).Clock.String(),
		})
	}
	match.notifyPlayers(gameStateResp)

	// Save game state
	match.Save()

	// Aborted because both player had disconnected
	if match.IsEnded() {
		return nil
	}

	// Check if game ended
	if match.game.Outcome() != chess.NoOutcome {
		logging.Info(
			"Game end by outcome",
			zap.String("outcome", match.game.Outcome().String()),
			zap.String("method", match.game.method()),
		)
		match.End()
	}
	return nil
}

func (h *MyMatchHandler) OnMatchAbort() error {
	return nil
}

func (h *MyMatchHandler) OnMatchSave() error {
	return nil
}

func (h *MyMatchHandler) OnMatchEnd() error {
	match := h.GetMatch().(*Match)
	for _, p := range match.GetPlayers() {
		player := p.(*Player)
		switch match.game.Outcome() {
		case chess.WhiteWon:
			if player.Side == WHITE_SIDE {
				player.SetResult(1)
			} else {
				player.SetResult(0)
			}
		case chess.BlackWon:
			if player.Side == BLACK_SIDE {
				player.SetResult(1)
			} else {
				player.SetResult(0)
			}
		case chess.Draw:
			player.SetResult(0.5)
		}
	}
	match.skipTimer()
	match.checkTimeout()
	return nil
}

func (h *MyMatchHandler) GetMatch() server.Match {
	return h.match
}

func NewMatchHandler(match *Match) server.MatchHandler {
	return &MyMatchHandler{
		match: match,
	}
}
