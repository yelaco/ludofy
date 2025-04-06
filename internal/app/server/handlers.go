package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
	"go.uber.org/zap"
)

func (s *server) handleAbortGame(match *Match) {
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
		MatchId: match.id,
		PlayerIds: []string{
			match.players[0].Id,
			match.players[1].Id,
		},
	}

	payload, err := json.Marshal(matchAbortReq)
	if err != nil {
		log.Fatal(err)
	}

	// Invoke Lambda function
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(s.config.AbortGameFunctionArn),
		Payload:        payload,
		InvocationType: types.InvocationTypeEvent,
	}

	_, err = lambdaClient.Invoke(ctx, input)
	if err != nil {
		logging.Fatal("failed to invoke abort game", zap.Error(err))
	}

	s.removeMatch(match.id)
	logging.Info("match aborted", zap.String("match_id", match.id))
}

// Handler for saving current game state.
func (s *server) handleSaveGame(match *Match) {
	ctx := context.Background()
	lastMove := match.game.lastMove()
	matchStateReq := dtos.MatchStateRequest{
		Id:      utils.GenerateUUID(),
		MatchId: match.id,
		PlayerStates: []dtos.PlayerStateRequest{
			{
				Clock:  match.players[0].Clock.String(),
				Status: match.players[0].Status.String(),
			},
			{
				Clock:  match.players[1].Clock.String(),
				Status: match.players[1].Status.String(),
			},
		},
		GameState: match.game.FEN(),
		Move: dtos.MoveRequest{
			PlayerId: lastMove.playerId,
			Uci:      lastMove.uci,
		},
		Ply:       match.currentPly(),
		Timestamp: time.Now(),
	}
	matchStateAppSyncReq := dtos.NewMatchStateAppSyncRequest(matchStateReq)
	payload, err := json.Marshal(matchStateAppSyncReq)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
		return
	}

	req, err := http.NewRequest(
		"POST",
		s.config.AppSyncHttpUrl,
		bytes.NewReader(payload),
	)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Sign the request
	signer := v4.NewSigner()
	credentials, err := s.config.AwsCfg.Credentials.Retrieve(ctx)
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
		s.config.AwsRegion,
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

// Handler for when a game match ends.
func (s *server) handleEndGame(match *Match) {
	if match == nil {
		return
	}
	ctx := context.TODO()

	newRatings, newRDs, err := match.getNewPlayerRatings()
	if err != nil {
		logging.Fatal("failed to invoke end game", zap.Error(err))
	}
	matchRecordReq := dtos.MatchRecordRequest{
		MatchId: match.id,
		Players: []dtos.PlayerRecordRequest{
			{
				Id:        match.players[0].Id,
				OldRating: match.players[0].Rating,
				NewRating: newRatings[0],
				OldRD:     match.players[0].RD,
				NewRD:     newRDs[0],
			},
			{
				Id:        match.players[1].Id,
				OldRating: match.players[1].Rating,
				NewRating: newRatings[1],
				NewRD:     newRDs[1],
				OldRD:     match.players[1].RD,
			},
		},
		Pgn:       match.game.String(),
		StartedAt: match.startAt,
		EndedAt:   time.Now(),
	}
	switch match.game.outcome() {
	case chess.WhiteWon:
		matchRecordReq.Results = []float64{1.0, 0.0}
	case chess.BlackWon:
		matchRecordReq.Results = []float64{0.0, 1.0}
	case chess.Draw, chess.NoOutcome:
		matchRecordReq.Results = []float64{0.5, 0.5}
	}
	payload, err := json.Marshal(matchRecordReq)
	if err != nil {
		log.Fatal(err)
	}

	// Invoke Lambda function
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(s.config.EndGameFunctionArn),
		Payload:        payload,
		InvocationType: types.InvocationTypeRequestResponse,
	}

	_, err = s.lambdaClient.Invoke(ctx, input)
	if err != nil {
		logging.Fatal("failed to invoke end game", zap.Error(err))
	}

	s.removeMatch(match.id)
	logging.Info("match ended", zap.String("match_id", match.id))
}

// Handler for when a user connection closes
func (s *server) handlePlayerDisconnect(match *Match, playerId string) {
	if match == nil {
		return
	}

	player, exist := match.getPlayerWithId(playerId)
	if !exist {
		logging.Fatal("invalid player id", zap.String("player_id", playerId))
		return
	}
	player.Conn = nil
	player.Status = DISCONNECTED

	currentClock := match.getCurrentTurnPlayer().Clock

	// If both player disconnected, set the clock to current turn clock
	if match.players[0].Status == match.players[1].Status {
		logging.Info(
			"both player disconnected",
			zap.String("match_id", match.id),
		)
		if !match.isEnded() {
			match.setTimer(currentClock)
		}
	} else {
		// Else only set the timer for the disconnected player
		logging.Info(
			"player disconnected",
			zap.String("match_id", match.id),
			zap.String("player_id", player.Id),
		)
		if !match.isEnded() {
			if currentClock < match.cfg.DisconnectTimeout {
				match.setTimer(currentClock)
			} else {
				match.setTimer(match.cfg.DisconnectTimeout)
			}
		}
		match.notifyAboutPlayerStatus(playerStatusResponse{
			Type:     "playerStatus",
			PlayerId: playerId,
			Status:   player.Status.String(),
		})
	}
}

func (s *server) handlePlayerJoin(
	conn *websocket.Conn,
	match *Match,
	playerId string,
) {
	if match == nil {
		return
	}

	player, exist := match.getPlayerWithId(playerId)
	if !exist {
		logging.Fatal("invalid player id", zap.String("player_id", playerId))
		return
	}
	if player.Status == INIT && player.Side == WHITE_SIDE {
		match.startAt = time.Now()
		player.TurnStartedAt = match.startAt
		match.setTimer(match.cfg.MatchDuration)
		err := s.storageClient.UpdateActiveMatch(
			context.Background(),
			match.id,
			storage.ActiveMatchUpdateOptions{
				StartedAt: aws.Time(match.startAt),
			},
		)
		if err != nil {
			logging.Error(
				"failed to update match: %w",
				zap.Error(err),
			)
		}
	}
	player.Conn = conn
	player.Status = CONNECTED

	match.syncPlayer(player)

	logging.Info("player connected",
		zap.String("player_id", playerId),
		zap.String("match_id", match.id),
	)

	match.notifyAboutPlayerStatus(playerStatusResponse{
		Type:     "playerStatus",
		PlayerId: playerId,
		Status:   player.Status.String(),
	})
}

// Handler for when user sends a message
func (s *server) handleWebSocketMessage(
	playerId string,
	match *Match,
	payload payload,
) {
	if match == nil {
		logging.Error("match not loaded")
		return
	}
	if time.Since(payload.CreatedAt) < 0 {
		logging.Info("invalid timestamp",
			zap.String("created_at", payload.CreatedAt.String()),
			zap.String("validate_time", time.Now().String()),
		)
		return
	}
	switch payload.Type {
	case "gameData":
		action := payload.Data["action"]
		switch action {
		case "abort":
			match.processGameControl(playerId, ABORT)
		case "resign":
			match.processGameControl(playerId, RESIGN)
		case "offerDraw":
			match.processGameControl(playerId, OFFER_DRAW)
		case "declineDraw":
			match.processGameControl(playerId, DECLINE_DRAW)
		case "move":
			match.processMove(playerId, payload.Data["move"], payload.CreatedAt)
		default:
			logging.Info("invalid game action:", zap.String("action", payload.Type))
			return
		}
		logging.Info(
			"game data",
			zap.String("match_id", match.id),
			zap.String("action", action),
		)
	case "sync":
		match.syncPlayerWithId(playerId)
	default:
		logging.Info("invalid payload type:", zap.String("type", payload.Type))
	}
}
