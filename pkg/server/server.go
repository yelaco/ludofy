package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/internal/aws/compute"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func NewFromConfig(cfg Config) Server {
	srv := &DefaultServer{
		address: "0.0.0.0:" + cfg.Port,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		mu:      new(sync.Mutex),
		cfg:     cfg,
		handler: cfg.ServerHandler,
	}
	storageClient = storage.NewClient(
		dynamodb.NewFromConfig(cfg.awsCfg),
	)
	computeClient = compute.NewClient(
		ecs.NewFromConfig(cfg.awsCfg),
		nil,
	)
	lambdaClient = lambda.NewFromConfig(cfg.awsCfg)

	srv.resetProtectionTimer(cfg.protectionTimeout)
	return srv
}

// Start method    starts the game server
func (s *DefaultServer) Start() error {
	// Server status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		count := s.totalMatches.Load()
		json.NewEncoder(w).Encode(map[string]any{
			"activeMatches": count,
			"maxMatches":    s.cfg.maxMatches,
			"canAccept":     count < s.cfg.maxMatches,
		})
	})

	// Websocket
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
					"match failed to load",
				),
				time.Now().Add(5*time.Second),
			)
			return
		}
		match.playerJoin(playerId, conn)

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
				match.playerDisconnect(playerId)
				break
			}

			err = s.HandleMessage(playerId, match, message)
			if err != nil {
				logging.Error("failed to handle message", zap.Error(err))
				conn.WriteControl(
					websocket.CloseNormalClosure,
					nil,
					time.Now().Add(5*time.Second),
				)
			}
		}
	})
	logging.Info("websocket server started", zap.String("port", s.cfg.Port))
	return http.ListenAndServe(s.address, nil)
}

func (s *DefaultServer) HandleMessage(playerId string, match Match, msg []byte) error {
	if match == nil {
		return fmt.Errorf("match not loaded")
	}
	err := s.handler.OnHandleMessage(playerId, match, msg)
	if err != nil {
		return fmt.Errorf("on handle message: %w", err)
	}
	return nil
}

func (s *DefaultServer) HandleMatchEnd(match Match) {
	if match == nil {
		return
	}
	matchRecordReq := dtos.MatchRecordRequest{
		MatchId: match.GetId(),
		EndedAt: time.Now(),
	}

	if err := s.handler.OnHandleMatchEnd(&matchRecordReq, match); err != nil {
		logging.Fatal("failed to hanlde match end", zap.Error(err))
	}

	payload, err := json.Marshal(matchRecordReq)
	if err != nil {
		logging.Fatal("failed to marshal match record request", zap.Error(err))
	}

	// Invoke Lambda function
	_, err = lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName:   aws.String(s.cfg.endGameFunctionArn),
		Payload:        payload,
		InvocationType: types.InvocationTypeRequestResponse,
	})
	if err != nil {
		logging.Error("failed to invoke end game", zap.Error(err))
	}

	s.removeMatch(match.GetId())
	logging.Info("match ended", zap.String("match_id", match.GetId()))
}

func (s *DefaultServer) HandleMatchSave(match Match) {
	ctx := context.Background()
	matchStateReq := dtos.MatchStateRequest{
		Id:        utils.GenerateUUID(),
		MatchId:   match.GetId(),
		Timestamp: time.Now(),
	}
	s.handler.OnHandleMatchSave(&matchStateReq, match)

	matchStateAppSyncReq := dtos.NewMatchStateAppSyncRequest(matchStateReq)
	payload, err := json.Marshal(matchStateAppSyncReq)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
		return
	}

	req, err := http.NewRequest(
		"POST",
		s.cfg.appSyncHttpUrl,
		bytes.NewReader(payload),
	)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Sign the request
	signer := v4.NewSigner()
	credentials, err := s.cfg.awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
		return
	}
	err = signer.SignHTTP(
		ctx,
		credentials,
		req,
		sha256Hash(payload),
		"appsync",
		s.cfg.awsCfg.Region,
		time.Now(),
	)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
		return
	}

	client := new(http.Client)
	response, err := client.Do(req)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
		return
	}

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			logging.Error("Failed to read response body", zap.Error(err))
			return
		}
		logging.Error("Failed to save game",
			zap.String("body", string(body)),
			zap.Error(fmt.Errorf("200 expected")),
		)
	}
}

func (s *DefaultServer) HandleMatchAbort(match Match) {
	if match == nil {
		return
	}
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logging.Fatal("unable to load SDK config", zap.Error(err))
	}
	lambdaClient := lambda.NewFromConfig(cfg)

	matchAbortReq := dtos.MatchAbortRequest{
		MatchId:   match.GetId(),
		PlayerIds: make([]string, 0, len(match.GetPlayers())),
	}

	payload, err := json.Marshal(matchAbortReq)
	if err != nil {
		log.Fatal(err)
	}

	// Invoke Lambda function
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(s.cfg.abortGameFunctionArn),
		Payload:        payload,
		InvocationType: types.InvocationTypeEvent,
	}

	_, err = lambdaClient.Invoke(ctx, input)
	if err != nil {
		logging.Fatal("failed to invoke abort game", zap.Error(err))
	}

	s.removeMatch(match.GetId())
	logging.Info("match aborted", zap.String("match_id", match.GetId()))
}

// auth method    authenticates and extract userId
func (s *DefaultServer) auth(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no authorization")
	}
	validToken, err := auth.ValidateJwt(token, s.cfg.cognitoPublicKeys)
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

func (s *DefaultServer) loadMatch(matchId string) (Match, error) {
	ctx := context.Background()

	activeMatch, err := storageClient.GetActiveMatch(ctx, matchId)
	if err != nil {
		return nil, fmt.Errorf("failed to get active match: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	value, loaded := s.matches.Load(matchId)
	if loaded {
		match, ok := value.(*DefaultMatch)
		if ok {
			logging.Info("match loaded")
			return match, nil
		}
		return nil, ErrFailedToLoadMatch
	} else {
		matchStates, _, err := storageClient.FetchMatchStates(
			ctx,
			matchId,
			nil,
			1,
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch match states: %w", err)
		}

		var match Match
		if len(matchStates) > 0 {
			match, err = s.handler.OnMatchResume(activeMatch, matchStates[0])
			if err != nil {
				return nil, fmt.Errorf("failed to resume match: %w", err)
			}
		} else {
			match, err = s.handler.OnMatchCreate(activeMatch)
			if err != nil {
				return nil, fmt.Errorf("failed to create match: %w", err)
			}
		}

		match.setHandler(s.cfg.MatchHandler)
		match.setSaveCallback(s.HandleMatchSave)
		match.setEndCallback(s.HandleMatchEnd)
		match.setAbortCallback(s.HandleMatchAbort)
		s.matches.Store(matchId, match)
		s.totalMatches.Add(1)
		s.resetProtectionTimer(45 * time.Minute)

		go match.start()
		logging.Info("match loaded", zap.String("match_id", matchId))
		return match, nil
	}
}

func (s *DefaultServer) removeMatch(matchId string) {
	s.matches.Delete(matchId)
	total := s.totalMatches.Add(-1)
	if total <= 0 {
		s.skipProtectionTimer()
	}
	logging.Info("match removed", zap.Int32("total_matches", total))
}

func (s *DefaultServer) skipProtectionTimer() {
	if s.protectionTimer == nil {
		return
	}
	s.protectionTimer.Reset(0)
	logging.Info("server protection timer skipped")
}

func (s *DefaultServer) enableProtection() {
	err := computeClient.UpdateServerProtection(context.TODO(), true)
	if err != nil {
		logging.Info("failed to enable server protection", zap.Error(err))
		return
	}
	logging.Info("server protection enabled")
}

func (s *DefaultServer) disableProtection() {
	err := computeClient.UpdateServerProtection(context.TODO(), false)
	if err != nil {
		logging.Info("failed to disable server protection", zap.Error(err))
		return
	}
	logging.Info("server protection disabled")
}

func (s *DefaultServer) resetProtectionTimer(duration time.Duration) {
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
