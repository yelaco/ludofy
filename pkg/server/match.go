package server

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type playerStatusResponse struct {
	Type     string `json:"type"`
	PlayerId string `json:"playerId"`
	Status   string `json:"status"`
}

type errorResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

func NewDefaultMatch(id string, players map[string]Player) Match {
	return &DefaultMatch{
		Id:      id,
		Players: players,
		moveCh:  make(chan Move),
		mu:      new(sync.Mutex),
	}
}

func (m *DefaultMatch) start() {
	for move := range m.moveCh {
		player, exist := m.Players[move.GetPlayerId()]
		if !exist {
			player.WriteJson(errorResponse{
				Type:  "error",
				Error: ErrStatusInvalidPlayerId,
			})
			continue
		}
		m.handler.HandleMove(player, move)
	}
}

func (m *DefaultMatch) Abort() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ended {
		return
	}
	m.ended = true
	if !utils.IsClosed(m.moveCh) {
		close(m.moveCh)
	}
	m.DisconnectPlayers("match aborted", time.Now().Add(5*time.Second))
	m.handler.OnMatchAbort()
	m.abortCallback(m)
}

func (m *DefaultMatch) Save() {
	m.handler.OnMatchSave()
	m.saveCallback(m)
}

func (m *DefaultMatch) End() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ended {
		return
	}
	m.ended = true
	if !utils.IsClosed(m.moveCh) {
		close(m.moveCh)
	}
	m.handler.OnMatchEnd()
	m.DisconnectPlayers("match ended", time.Now().Add(5*time.Second))
	m.endCallback(m)
}

func (m *DefaultMatch) ProcessMove(move Move) {
	m.moveCh <- move
}

func (m *DefaultMatch) GetPlayerWithId(id string) (Player, bool) {
	player, exist := m.Players[id]
	return player, exist
}

func (m *DefaultMatch) IsEnded() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ended
}

func (m *DefaultMatch) playerJoin(playerId string, conn *websocket.Conn) {
	if m == nil {
		return
	}

	player, exist := m.GetPlayerWithId(playerId)
	if !exist {
		logging.Fatal("invalid player id", zap.String("player_id", playerId))
		return
	}

	init, err := m.handler.OnPlayerJoin(player)
	if err != nil {
		logging.Error("on player join", zap.Error(err))
	}
	if init {
		err := storageClient.UpdateActiveMatch(
			context.Background(),
			m.GetId(),
			storage.ActiveMatchUpdateOptions{
				StartedAt: aws.Time(time.Now()),
			},
		)
		if err != nil {
			logging.Error("failed to update match", zap.Error(err))
		}
	}

	player.setConn(conn)
	m.handler.OnPlayerSync(player)

	m.notifyAboutPlayerStatus(playerStatusResponse{
		Type:     "playerStatus",
		PlayerId: playerId,
		Status:   player.GetStatus(),
	})
}

func (m *DefaultMatch) playerDisconnect(playerId string) {
	if m == nil {
		return
	}

	player, exist := m.GetPlayerWithId(playerId)
	if !exist {
		logging.Fatal("invalid player id", zap.String("player_id", playerId))
		return
	}
	player.setConn(nil)

	m.handler.OnPlayerLeave(player)

	m.notifyAboutPlayerStatus(playerStatusResponse{
		Type:     "playerStatus",
		PlayerId: playerId,
		Status:   player.GetStatus(),
	})
}

func (m *DefaultMatch) notifyAboutPlayerStatus(resp playerStatusResponse) {
	for _, player := range m.Players {
		if player.GetId() == resp.PlayerId {
			continue
		}
		err := player.WriteJson(resp)
		if err != nil {
			logging.Error(
				"couldn't notify player: ",
				zap.String("player_id", player.GetId()),
			)
		}
	}
}

func (m *DefaultMatch) DisconnectPlayers(msg string, deadline time.Time) {
	for _, player := range m.Players {
		player.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseNormalClosure,
				msg,
			),
			deadline,
		)
	}
}

func (m *DefaultMatch) SetHandler(handler MatchHandler) {
	m.handler = handler
}

func (m *DefaultMatch) setSaveCallback(callback func(Match)) {
	m.saveCallback = callback
}

func (m *DefaultMatch) setEndCallback(callback func(Match)) {
	m.endCallback = callback
}

func (m *DefaultMatch) setAbortCallback(callback func(Match)) {
	m.abortCallback = callback
}

func (m *DefaultMatch) GetId() string {
	return m.Id
}

func (m *DefaultMatch) GetPlayers() map[string]Player {
	return m.Players
}
