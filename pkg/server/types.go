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

type Server struct {
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
	Handler         ServerHandler
}

type Player struct {
	Id       string
	Conn     *websocket.Conn
	MatchId  string
	SendChan chan any
	Status   Status

	mu *sync.Mutex
}

type Match struct {
	Id        string
	Players   map[string]*Player
	MoveCh    chan Move
	StartedAt time.Time

	endCallback   func(*Match)
	saveCallback  func(*Match)
	abortCallback func(*Match)

	ended bool
	mu    *sync.Mutex

	Handler MatchHandler
}

type Move struct {
	PlayerId string `json:"playerId"`
	Payload  any    `json:"payload"`
}
