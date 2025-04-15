package server

import (
	"time"

	"github.com/gorilla/websocket"
)

type Status uint8

const (
	INIT Status = iota
	CONNECTED
	DISCONNECTED
)

func (p *Player) setConn(conn *websocket.Conn) {
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

func (p *Player) WriteJson(msg interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p == nil || p.Conn == nil {
		return nil
	}
	return p.Conn.WriteJSON(msg)
}

func (p *Player) WriteControl(messageType int, data []byte, deadline time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p == nil || p.Conn == nil {
		return nil
	}
	return p.Conn.WriteControl(messageType, data, deadline)
}
