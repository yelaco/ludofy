package server

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/chess-vn/slchess/internal/aws/auth"
	awsAuth "github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/compute"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/entities"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type server struct {
	address  string
	upgrader websocket.Upgrader

	config       Config
	matches      sync.Map
	totalMatches atomic.Int32
	mu           sync.Mutex

	cognitoPublicKeys map[string]*rsa.PublicKey
	storageClient     *storage.Client
	computeClient     *compute.Client
	lambdaClient      *lambda.Client

	protectionTimer *utils.Timer
}

type payload struct {
	Type      string            `json:"type"`
	Data      map[string]string `json:"data"`
	CreatedAt time.Time         `json:"createdAt"`
}

func NewServer() *server {
	cfg := NewConfig()
	tokenSigningKeyUrl := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		cfg.AwsRegion,
		cfg.CognitoUserPoolId,
	)
	cognitoPublicKeys, err := awsAuth.LoadCognitoPublicKeys(tokenSigningKeyUrl)
	if err != nil {
		panic(err)
	}
	awsCfg, _ := config.LoadDefaultConfig(context.TODO())
	srv := &server{
		address: "0.0.0.0:" + cfg.Port,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
		config:            cfg,
		cognitoPublicKeys: cognitoPublicKeys,
		storageClient: storage.NewClient(
			dynamodb.NewFromConfig(awsCfg),
		),
		computeClient: compute.NewClient(
			ecs.NewFromConfig(awsCfg),
			nil,
		),
		lambdaClient: lambda.NewFromConfig(awsCfg),
	}
	srv.resetProtectionTimer(cfg.IdleTimeout)
	return srv
}

// Start method    starts the game server
func (s *server) Start() error {
	http.HandleFunc("/game/{matchId}", func(w http.ResponseWriter, r *http.Request) {
		playerId, err := s.auth(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			logging.Error("failed to auth: %w", zap.Error(err))
			return
		}

		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Error(
				"failed to upgrade connection",
				zap.String("error", err.Error()),
			)
			return
		}
		defer conn.Close()

		matchId := r.PathValue("matchId")
		match, err := s.loadMatch(matchId)
		if err != nil {
			logging.Info("failed to load match", zap.String("error", err.Error()))
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.CloseNormalClosure,
					"match expired",
				),
				time.Now().Add(5*time.Second),
			)
			return
		}
		s.handlePlayerJoin(conn, match, playerId)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(
					err,
					websocket.CloseNormalClosure,
				) {
					logging.Info(
						"connection closed gracefully",
						zap.String("remote_address", conn.RemoteAddr().String()),
					)
				} else if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseAbnormalClosure,
				) {
					logging.Info(
						"unexpected connection close",
						zap.String("remote_address", conn.RemoteAddr().String()),
						zap.Error(err),
					)
				}
				s.handlePlayerDisconnect(match, playerId)
				break
			}

			var payload payload
			if err := json.Unmarshal(message, &payload); err != nil {
				conn.Close()
			}
			s.handleWebSocketMessage(playerId, match, payload)
		}
	})
	logging.Info("websocket server started", zap.String("port", s.config.Port))
	return http.ListenAndServe(s.address, nil)
}

// mustAuth method    authenticates and extract userId
func (s *server) auth(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no authorization")
	}
	validToken, err := auth.ValidateJwt(token, s.cognitoPublicKeys)
	if err != nil || !validToken.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	mapClaims, ok := validToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid map claims")
	}
	v, ok := mapClaims["sub"]
	if !ok {
		return "", fmt.Errorf("user id not found")
	}
	userId, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("invalid user id")
	}
	return userId, nil
}

/*
loadMatch method    loads match with corresponding matchId.
If no such match exists, create a new match.
This is used to start the match only when white side player send in the first valid move.
*/
func (s *server) loadMatch(matchId string) (*Match, error) {
	ctx := context.Background()

	activeMatch, err := s.storageClient.GetActiveMatch(ctx, matchId)
	if err != nil {
		return nil, fmt.Errorf("failed to get active match: %w", err)
	}

	config, err := configForGameMode(activeMatch.GameMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get match config: %w", err)
	}

	// Check if match is expired
	if (activeMatch.StartedAt == nil && time.Since(activeMatch.CreatedAt) > 2*time.Minute) ||
		(activeMatch.StartedAt != nil && time.Since(*activeMatch.StartedAt) > 2*config.MatchDuration+2*time.Minute) {
		err := s.removeExpiredMatch(activeMatch)
		if err != nil {
			return nil, fmt.Errorf("failed to remove expired match: %w", err)
		}
		return nil, fmt.Errorf("match expired")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	value, loaded := s.matches.Load(matchId)
	if loaded {
		match, ok := value.(*Match)
		if ok {
			logging.Info("match loaded")
			return match, nil
		}
		return nil, ErrFailedToLoadMatch
	} else {
		matchStates, _, err := s.storageClient.FetchMatchStates(
			ctx,
			matchId,
			nil,
			1,
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch match states: %w", err)
		}

		var clock1 time.Duration
		var clock2 time.Duration

		player1 := newPlayer(
			nil,
			activeMatch.Player1.Id,
			WHITE_SIDE,
			clock1,
			activeMatch.Player1.Rating,
			activeMatch.Player1.RD,
			activeMatch.Player1.NewRatings,
			activeMatch.Player1.NewRDs,
		)
		player2 := newPlayer(
			nil,
			activeMatch.Player2.Id,
			BLACK_SIDE,
			clock2,
			activeMatch.Player2.Rating,
			activeMatch.Player2.RD,
			activeMatch.Player2.NewRatings,
			activeMatch.Player1.NewRDs,
		)

		var match *Match
		if len(matchStates) > 0 {
			player1.Clock, _ = time.ParseDuration(matchStates[0].PlayerStates[0].Clock)
			player2.Clock, _ = time.ParseDuration(matchStates[0].PlayerStates[1].Clock)
			match, err = s.resumeMatch(
				matchId,
				player1,
				player2,
				config,
				matchStates[0].GameState,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to resume match: %w", err)
			}
		} else {
			player1.Clock = config.MatchDuration
			player2.Clock = config.MatchDuration
			match = s.newMatch(matchId, player1, player2, config)
		}
		logging.Info(
			"match loaded",
			zap.String("match_id", matchId),
			zap.String("player1_id", player1.Id),
			zap.String("player2_id", player2.Id),
		)

		s.matches.Store(matchId, match)
		s.totalMatches.Add(1)
		s.resetProtectionTimer(2*match.cfg.MatchDuration + 5*time.Minute)

		return match, nil
	}
}

func (s *server) newMatch(
	matchId string,
	player1,
	player2 player,
	config MatchConfig,
) *Match {
	match := &Match{
		id:               matchId,
		game:             newGame(),
		players:          []*player{&player1, &player2},
		moveCh:           make(chan move),
		cfg:              config,
		abortGameHandler: s.handleAbortGame,
		endGameHandler:   s.handleEndGame,
		saveGameHandler:  s.handleSaveGame,
	}
	// Timeout to cancel match if first move is not made
	match.setTimer(config.CancelTimeout)
	go match.start()
	return match
}

func (s *server) resumeMatch(
	matchId string,
	player1,
	player2 player,
	config MatchConfig,
	gameState string,
) (*Match, error) {
	game, err := restoreGame(gameState)
	if err != nil {
		return nil, fmt.Errorf("failed to restore game: %w", err)
	}
	match := &Match{
		id:               matchId,
		game:             game,
		players:          []*player{&player1, &player2},
		moveCh:           make(chan move),
		cfg:              config,
		abortGameHandler: s.handleAbortGame,
		endGameHandler:   s.handleEndGame,
		saveGameHandler:  s.handleSaveGame,
	}
	// Timeout to cancel match if first move is not made
	match.setTimer(config.CancelTimeout)
	go match.start()
	return match, nil
}

func (s *server) removeMatch(matchId string) {
	s.matches.Delete(matchId)
	total := s.totalMatches.Add(-1)
	if total <= 0 {
		s.skipProtectionTimer()
	}
	logging.Info("match removed", zap.Int32("total_matches", total))
}

func (s *server) resetProtectionTimer(duration time.Duration) {
	if s.protectionTimer != nil {
		if s.protectionTimer.TimeRemaining() < duration {
			s.protectionTimer.Reset(duration)
		}
		logging.Info("server protection timer reset",
			zap.String("duration", duration.String()),
		)
		return
	}
	s.protectionTimer = utils.NewTimer(duration)
	go func() {
		s.enableProtection()
		<-s.protectionTimer.C()
		s.disableProtection()
		s.protectionTimer = nil
	}()
	logging.Info("server protection timer set",
		zap.String("duration", duration.String()),
	)
}

func (s *server) skipProtectionTimer() {
	if s.protectionTimer == nil {
		return
	}
	s.protectionTimer.Reset(0)
	logging.Info("server protection timer skipped")
}

func (s *server) enableProtection() {
	err := s.computeClient.UpdateServerProtection(context.TODO(), true)
	if err != nil {
		logging.Info("failed to enable server protection", zap.Error(err))
		return
	}
	logging.Info("server protection enabled")
}

func (s *server) disableProtection() {
	err := s.computeClient.UpdateServerProtection(context.TODO(), false)
	if err != nil {
		logging.Info("failed to disable server protection", zap.Error(err))
		return
	}
	logging.Info("server protection disabled")
}

func (s *server) removeExpiredMatch(activeMatch entities.ActiveMatch) error {
	ctx := context.Background()
	err := s.storageClient.DeleteActiveMatch(ctx, activeMatch.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete match: %w", err)
	}
	err = s.storageClient.DeleteUserMatch(ctx, activeMatch.Player1.Id)
	if err != nil {
		return fmt.Errorf("failed to delete user match: %w", err)
	}
	err = s.storageClient.DeleteUserMatch(ctx, activeMatch.Player2.Id)
	if err != nil {
		return fmt.Errorf("failed to delete user match: %w", err)
	}
	err = s.storageClient.DeleteSpectatorConversation(ctx, activeMatch.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete user match: %w", err)
	}
	return nil
}
