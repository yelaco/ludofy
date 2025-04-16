package server

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Status uint8

const (
	INIT Status = iota
	CONNECTED
	DISCONNECTED
)

func NewDefaultPlayer(playerId, matchId string) Player {
	return &DefaultPlayer{
		Id:      playerId,
		Conn:    nil,
		MatchId: matchId,
		Status:  INIT,
		mu:      new(sync.Mutex),
	}
}

func (p *DefaultPlayer) setConn(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if conn == nil {
		p.Status = DISCONNECTED
	} else {
		p.Status = CONNECTED
	}
	p.Conn = conn
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

func (p *DefaultPlayer) GetId() string {
	return p.Id
}

func (p *DefaultPlayer) GetStatus() string {
	return p.Status.String()
}

func (p *DefaultPlayer) WriteJson(msg interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p == nil || p.Conn == nil {
		return nil
	}
	return p.Conn.WriteJSON(msg)
}

func (p *DefaultPlayer) WriteControl(messageType int, data []byte, deadline time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p == nil || p.Conn == nil {
		return nil
	}
	return p.Conn.WriteControl(messageType, data, deadline)
}
