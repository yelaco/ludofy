package server

import (
	"time"

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

func (m *Match) start() {
	for move := range m.MoveCh {
		player, exist := m.Players[move.PlayerId]
		if !exist {
			// Notify
			continue
		}
		m.Handler.HandleMove(m, player, move)
	}
}

func (m *Match) Abort() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ended {
		return
	}
	m.ended = true
	if !utils.IsClosed(m.MoveCh) {
		close(m.MoveCh)
	}
	m.disconnectPlayers("match aborted", time.Now().Add(5*time.Second))
	m.abortCallback(m)
}

func (m *Match) Save() {
	m.saveCallback(m)
}

func (m *Match) End() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ended {
		return
	}
	m.ended = true
	if !utils.IsClosed(m.MoveCh) {
		close(m.MoveCh)
	}
	m.Handler.OnMatchEnd(m)
	m.disconnectPlayers("match ended", time.Now().Add(5*time.Second))
	m.endCallback(m)
}

func (m *Match) ProcessMove(move Move) {
	m.MoveCh <- move
}

func (m *Match) GetPlayerWithId(id string) (*Player, bool) {
	player, exist := m.Players[id]
	return player, exist
}

func (m *Match) playerJoin(playerId string, conn *websocket.Conn) {
	if m == nil {
		return
	}

	player, exist := m.GetPlayerWithId(playerId)
	if !exist {
		logging.Fatal("invalid player id", zap.String("player_id", playerId))
		return
	}

	player.setConn(conn)
	m.Handler.OnPlayerSync(m, player)

	m.notifyAboutPlayerStatus(playerStatusResponse{
		Type:     "playerStatus",
		PlayerId: playerId,
		Status:   player.Status.String(),
	})
}

func (m *Match) playerDisconnect(playerId string) {
	if m == nil {
		return
	}

	player, exist := m.GetPlayerWithId(playerId)
	if !exist {
		logging.Fatal("invalid player id", zap.String("player_id", playerId))
		return
	}

	player.setConn(nil)
	m.Handler.OnPlayerLeave(m, player)

	m.notifyAboutPlayerStatus(playerStatusResponse{
		Type:     "playerStatus",
		PlayerId: playerId,
		Status:   player.Status.String(),
	})
}

func (m *Match) notifyAboutPlayerStatus(resp playerStatusResponse) {
	for _, player := range m.Players {
		if player.Id == resp.PlayerId {
			continue
		}
		err := player.WriteJson(resp)
		if err != nil {
			logging.Error(
				"couldn't notify player: ",
				zap.String("player_id", player.Id),
			)
		}
	}
}

func (m *Match) disconnectPlayers(msg string, deadline time.Time) {
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
