package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
	"go.uber.org/zap"
)

const (
	PENDING  = "pending"
	DECLINED = "declined"
)

type Match struct {
	id      string
	players []*player
	game    *game
	moveCh  chan move
	timer   *time.Timer
	startAt time.Time
	cfg     MatchConfig

	endGameHandler   func(*Match)
	saveGameHandler  func(*Match)
	abortGameHandler func(*Match)

	ended bool
	mu    sync.Mutex
}

type MatchConfig struct {
	MatchDuration      time.Duration
	ClockIncrement     time.Duration
	CancelTimeout      time.Duration
	DisconnectTimeout  time.Duration
	MaxLagForgivenTime time.Duration
}

type matchResponse struct {
	Type      string            `json:"type"`
	GameState gameStateResponse `json:"game"`
}

type gameStateResponse struct {
	Outcome  string   `json:"outcome"`
	Method   string   `json:"method"`
	Fen      string   `json:"fen"`
	Clocks   []string `json:"clocks"`
	Statuses []string `json:"statuses,omitempty"`
}

type playerStatusResponse struct {
	Type     string `json:"type"`
	PlayerId string `json:"playerId"`
	Status   string `json:"status"`
}

type errorResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

type drawOfferResponse struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

func (match *Match) start() {
	for move := range match.moveCh {
		player, exist := match.getPlayerWithId(move.playerId)
		if !exist {
			player.Conn.WriteJSON(errorResponse{
				Type:  "error",
				Error: ErrStatusInvalidPlayerId,
			})
			continue
		}
		switch move.control {
		case ABORT:
			if match.currentPly() > 1 {
				player.Conn.WriteJSON(errorResponse{
					Type:  "error",
					Error: ErrStatusAbortInvalidPly,
				})
			}
			match.abort()
			continue
		case RESIGN:
			match.game.Resign(player.color())
		case OFFER_DRAW:
			draw := match.game.OfferDraw(player.color())
			if !draw {
				match.sendDrawOfferNotification(player, PENDING)
				continue
			}
		case DECLINE_DRAW:
			shouldNotify := match.game.DeclineDraw(player.color())
			if shouldNotify {
				match.sendDrawOfferNotification(player, DECLINED)
			}
			continue
		default:
			if expectedId := match.getCurrentTurnPlayer().Id; player.Id != expectedId {
				player.Conn.WriteJSON(errorResponse{
					Type: "error",
					Error: fmt.Sprintf(
						"%s: want %s - got %s",
						ErrStatusWrongTurn,
						expectedId,
						player.Id,
					),
				})
				continue
			}
			err := match.game.move(move)
			if err != nil {
				player.Conn.WriteJSON(errorResponse{
					Type:  "error",
					Error: ErrStatusInvalidMove,
				})
				continue
			}

			// If making move, update clock
			timeTaken := time.Since(player.TurnStartedAt)
			lagForgiven := match.calculateLagForgiven(move.createdAt)
			player.updateClock(timeTaken, lagForgiven, match.cfg.ClockIncrement)

			// If clock runs out, end the game
			if player.Clock <= 0 {
				match.game.outOfTime(player.Side)
				logging.Info("out of time", zap.String("player_id", player.Id))
			} else {
				// else next turn
				currentTurnPlayer := match.getCurrentTurnPlayer()
				currentTurnPlayer.TurnStartedAt = time.Now()
				match.setTimer(currentTurnPlayer.Clock)
				logging.Info(
					"new turn",
					zap.String("player_id", currentTurnPlayer.Id),
					zap.String("clock_w", match.players[0].Clock.String()),
					zap.String("clock_b", match.players[1].Clock.String()),
				)
			}
		}

		match.notifyPlayers(gameStateResponse{
			Outcome: match.game.outcome().String(),
			Method:  match.game.method(),
			Fen:     match.game.FEN(),
			Clocks: []string{
				match.players[0].Clock.String(),
				match.players[1].Clock.String(),
			},
		})

		// Save game state
		match.save()

		// Aborted because both player had disconnected
		if match.isEnded() {
			logging.Info("Game aborted", zap.String("matchId", match.id))
			return
		}

		// Check if game ended
		if match.game.Outcome() != chess.NoOutcome {
			logging.Info(
				"Game end by outcome",
				zap.String("outcome", match.game.Outcome().String()),
				zap.String("method", match.game.method()),
			)
			match.end()
		}
	}
}

func (m *Match) sendDrawOfferNotification(sender *player, status string) {
	for _, player := range m.players {
		if player == nil || player.Conn == nil || player.Id == sender.Id {
			continue
		}
		err := player.Conn.WriteJSON(drawOfferResponse{
			Type:      "drawOffer",
			Status:    status,
			CreatedAt: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			logging.Error(
				"couldn't send draw offer notification to player: ",
				zap.String("player_id", player.Id),
			)
		}
	}
}

func (m *Match) notifyPlayers(resp gameStateResponse) {
	for _, player := range m.players {
		if player == nil || player.Conn == nil {
			continue
		}
		err := player.Conn.WriteJSON(matchResponse{
			Type:      "gameState",
			GameState: resp,
		})
		if err != nil {
			logging.Error(
				"couldn't notify player: ",
				zap.String("player_id", player.Id),
			)
		}
	}
}

func (m *Match) notifyAboutPlayerStatus(resp playerStatusResponse) {
	for _, player := range m.players {
		if player == nil || player.Conn == nil || player.Id == resp.PlayerId {
			continue
		}
		err := player.Conn.WriteJSON(resp)
		if err != nil {
			logging.Error(
				"couldn't notify player: ",
				zap.String("player_id", player.Id),
			)
		}
	}
}

func (m *Match) notifyAboutDeclinedOffer(resp playerStatusResponse) {
	for _, player := range m.players {
		if player == nil || player.Conn == nil || player.Id == resp.PlayerId {
			continue
		}
		err := player.Conn.WriteJSON(resp)
		if err != nil {
			logging.Error(
				"couldn't notify player: ",
				zap.String("player_id", player.Id),
			)
		}
	}
}

func (m *Match) syncPlayer(player *player) {
	err := player.Conn.WriteJSON(matchResponse{
		Type: "gameState",
		GameState: gameStateResponse{
			Outcome: m.game.outcome().String(),
			Method:  m.game.method(),
			Fen:     m.game.FEN(),
			Clocks: []string{
				m.players[0].Clock.String(),
				m.players[1].Clock.String(),
			},
			Statuses: []string{
				m.players[0].Status.String(),
				m.players[1].Status.String(),
			},
		},
	})
	if err != nil {
		logging.Error(
			"couldn't sync player: ",
			zap.String("player_id", player.Id),
		)
	}
}

func (m *Match) syncPlayerWithId(id string) {
	player, exist := m.getPlayerWithId(id)
	if !exist {
		logging.Error(
			"couldn't sync player: ",
			zap.String("player_id", player.Id),
		)
		return
	}
	m.syncPlayer(player)
}

func (m *Match) getPlayerWithId(id string) (*player, bool) {
	for _, player := range m.players {
		if player.Id == id {
			return player, true
		}
	}
	return nil, false
}

func (m *Match) getCurrentTurnPlayer() *player {
	if m.game.Position().Turn() == chess.White {
		return m.players[0]
	}
	return m.players[1]
}

func (m *Match) getNextTurnPlayer() *player {
	if m.game.Position().Turn() == chess.White {
		return m.players[1]
	}
	return m.players[0]
}

func (m *Match) currentPly() int {
	return len(m.game.moves)
}

func (m *Match) processMove(playerId, moveUci string, createdAt time.Time) {
	m.moveCh <- move{
		playerId:  playerId,
		uci:       moveUci,
		control:   NONE,
		createdAt: createdAt,
	}
}

func (m *Match) processGameControl(playerId string, control GameControl) {
	m.moveCh <- move{
		playerId: playerId,
		control:  control,
	}
}

func (m *Match) abort() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ended {
		return
	}
	m.ended = true
	if !utils.IsClosed(m.moveCh) {
		close(m.moveCh)
	}
	// Fire off the timer to remove end game handling job
	m.skipTimer()
	for _, player := range m.players {
		if player.Conn != nil {
			player.Conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.CloseNormalClosure,
					"match aborted",
				),
				time.Now().Add(5*time.Second),
			)
		}
	}
	m.abortGameHandler(m)
}

func (m *Match) save() {
	m.saveGameHandler(m)
}

func (m *Match) end() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ended {
		return
	}
	m.ended = true
	if !utils.IsClosed(m.moveCh) {
		close(m.moveCh)
	}
	// Fire off the timer to remove end game handling job
	m.skipTimer()
	m.checkTimeout()
	m.disconnectPlayers("match ended", time.Now().Add(5*time.Second))
	m.endGameHandler(m)
}

func (m *Match) isEnded() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ended
}

// setTimer method    set the timer to the specified duration before trigger end game handler
func (m *Match) setTimer(d time.Duration) {
	if m.timer != nil {
		m.timer.Reset(d)
		logging.Info(
			"clock reset",
			zap.String("match_id", m.id),
			zap.String("duration", d.String()),
		)
		return
	}
	m.timer = time.NewTimer(d)
	go func() {
		<-m.timer.C
		m.end()
	}()
	logging.Info(
		"clock set",
		zap.String("match_id", m.id),
		zap.String("duration", d.String()),
	)
}

// skipTimer method    skips timer by set timer to 0 duration timeout
func (m *Match) skipTimer() {
	if m.timer == nil {
		logging.Info("clock nil", zap.String("match_id", m.id))
		return
	}
	m.timer.Reset(0)
	logging.Info("clock skipped", zap.String("match_id", m.id))
}

func configForGameMode(gameMode string) (MatchConfig, error) {
	gm, err := entities.ParseGameMode(gameMode)
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

func (m *Match) getNewPlayerRatings() ([]float64, []float64, error) {
	switch m.game.outcome() {
	case chess.WhiteWon:
		return []float64{m.players[0].NewRatings[0], m.players[1].NewRatings[2]},
			[]float64{m.players[0].NewRDs[0], m.players[1].NewRDs[2]},
			nil
	case chess.BlackWon:
		return []float64{m.players[0].NewRatings[2], m.players[1].NewRatings[0]},
			[]float64{m.players[0].NewRDs[2], m.players[1].NewRDs[0]},
			nil
	case chess.Draw:
		return []float64{m.players[0].NewRatings[1], m.players[1].NewRatings[1]},
			[]float64{m.players[0].NewRDs[1], m.players[1].NewRDs[1]},
			nil
	case chess.NoOutcome:
		return []float64{m.players[0].Rating, m.players[1].Rating},
			[]float64{m.players[0].RD, m.players[1].RD},
			nil
	}
	return nil, nil, ErrInvalidOutcome
}

func (m *Match) calculateLagForgiven(moveCreatedAt time.Time) time.Duration {
	lagTime := time.Since(moveCreatedAt)
	if lagTime > m.cfg.MaxLagForgivenTime {
		return m.cfg.MaxLagForgivenTime
	}
	return lagTime
}

func (m *Match) checkTimeout() {
	if m.players[0].Status == CONNECTED &&
		m.players[1].Status == CONNECTED {
		return
	}
	if m.players[0].Status == INIT ||
		m.players[1].Status == INIT {
		m.disconnectPlayers("match cancelled", time.Now().Add(5*time.Second))
	}
	if m.players[0].Status == DISCONNECTED &&
		m.players[1].Status == CONNECTED {
		m.game.disconnectTimeout(m.players[0].Side)
	} else if m.players[0].Status == CONNECTED &&
		m.players[1].Status == DISCONNECTED {
		m.game.disconnectTimeout(m.players[1].Side)
	} else if m.players[0].Status == DISCONNECTED &&
		m.players[1].Status == DISCONNECTED {
		m.game.drawByTimeout()
	}
	m.notifyPlayers(gameStateResponse{
		Outcome: m.game.outcome().String(),
		Method:  m.game.method(),
		Fen:     m.game.FEN(),
		Clocks: []string{
			m.players[0].Clock.String(),
			m.players[1].Clock.String(),
		},
	})
}

func (m *Match) disconnectPlayers(msg string, deadline time.Time) {
	for _, player := range m.players {
		if player.Conn != nil {
			player.Conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.CloseNormalClosure,
					msg,
				),
				deadline,
			)
		}
	}
}
