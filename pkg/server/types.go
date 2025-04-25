package server

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/gorilla/websocket"
)

// Interfaces
type Server interface {
	Start() error
	HandleMessage(playerId string, match Match, msg []byte) error
	HandleMatchEnd(match Match)
	HandleMatchSave(match Match)
	HandleMatchAbort(match Match)
}

type Match interface {
	start()
	SetHandler(MatchHandler)
	setSaveCallback(func(Match))
	setEndCallback(func(Match))
	setAbortCallback(func(Match))
	playerJoin(playerId string, conn *websocket.Conn)
	playerDisconnect(playerId string)
	GetId() string
	GetPlayers() map[string]Player
	Abort()
	Save()
	End()
	IsEnded() bool
	ProcessMove(move Move)
	GetPlayerWithId(id string) (Player, bool)
	DisconnectPlayers(msg string, deadline time.Time)
}

type Player interface {
	setConn(conn *websocket.Conn)
	GetId() string
	GetStatus() string
	WriteJson(msg interface{}) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
}

type Move interface {
	GetPlayerId() string
}

// Default implementations
type DefaultServer struct {
	address  string
	upgrader websocket.Upgrader

	cfg          Config
	matches      sync.Map
	totalMatches atomic.Int32
	mu           *sync.Mutex

	protectionTimer *utils.Timer
	handler         ServerHandler
}

type DefaultPlayer struct {
	Id      string
	Conn    *websocket.Conn
	MatchId string
	Status  Status

	mu *sync.Mutex
}

type DefaultMatch struct {
	Id      string
	Players map[string]Player
	moveCh  chan Move

	endCallback   func(Match)
	saveCallback  func(Match)
	abortCallback func(Match)

	ended bool
	mu    *sync.Mutex

	handler MatchHandler
}

type DefaultMove struct {
	PlayerId string `json:"playerId"`
}
