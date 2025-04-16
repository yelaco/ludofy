package server

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/chess-vn/slchess/internal/aws/compute"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/gorilla/websocket"
)

// Interfaces
type Server interface {
	Start() error
	HandleMessage(playerId string, match Match, msg []byte)
	HandleMatchEnd(match Match)
	HandleMatchSave(match Match)
	HandleMatchAbort(match Match)
}

type Match interface {
	start()
	setHandler(MatchHandler)
	setSaveCallback(func(Match))
	setEndCallback(func(Match))
	setAbortCallback(func(Match))
	playerJoin(playerId string, conn *websocket.Conn)
	playerDisconnect(playerId string)
	GetId() string
	GetPlayers() map[string]Player
	GetStartedAt() time.Time
	Abort()
	Save()
	End()
	ProcessMove(move Move)
	GetPlayerWithId(id string) (Player, bool)
}

type Player interface {
	setConn(conn *websocket.Conn)
	GetId() string
	GetStatus() string
	WriteJson(msg interface{}) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
}

// Default implementations
type DefaultServer struct {
	address  string
	upgrader websocket.Upgrader

	cfg          Config
	matches      sync.Map
	totalMatches atomic.Int32
	mu           *sync.Mutex

	storageClient *storage.Client
	computeClient *compute.Client
	lambdaClient  *lambda.Client

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
	Id        string
	Players   map[string]Player
	moveCh    chan Move
	StartedAt time.Time

	endCallback   func(Match)
	saveCallback  func(Match)
	abortCallback func(Match)

	ended bool
	mu    *sync.Mutex

	handler MatchHandler
}

type Move struct {
	PlayerId string `json:"playerId"`
	Payload  any    `json:"payload"`
}
