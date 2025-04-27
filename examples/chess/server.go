package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yelaco/ludofy/internal/domains/dtos"
	"github.com/yelaco/ludofy/internal/domains/entities"
	"github.com/yelaco/ludofy/pkg/server"
)

type Payload struct {
	Type      string            `json:"type"`
	Data      map[string]string `json:"data"`
	CreatedAt time.Time         `json:"createdAt"`
}

/*
 * Implement ServerHandler interface
 */
type MyServerHandler struct{}

func (h *MyServerHandler) OnMatchCreate(activeMatch entities.ActiveMatch) (server.Match, error) {
	cfg, err := ConfigForGameMode(activeMatch.GameMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	players := make(map[string]server.Player, len(activeMatch.Players))
	for i, player := range activeMatch.Players {
		players[player.Id] = &Player{
			Player: server.NewDefaultPlayer(player.Id, activeMatch.MatchId),
			Clock:  cfg.MatchDuration,
			Side:   i%2 == 0,
		}
	}
	match := Match{
		Match: server.NewDefaultMatch(activeMatch.MatchId, players),
		cfg:   cfg,
		game:  NewGame(),
	}
	match.setTimer(cfg.CancelTimeout)
	matchHandler := NewMatchHandler(&match)
	match.SetHandler(matchHandler)
	return match, nil
}

func (h *MyServerHandler) OnMatchResume(
	activeMatch entities.ActiveMatch,
	currentState entities.MatchState,
) (server.Match, error) {
	cfg, err := ConfigForGameMode(activeMatch.GameMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	players := make(map[string]server.Player, len(activeMatch.Players))
	for i, player := range activeMatch.Players {
		players[player.Id] = &Player{
			Player: server.NewDefaultPlayer(player.Id, activeMatch.MatchId),
			Clock:  cfg.MatchDuration,
			Side:   i%2 == 0,
		}
	}
	game, err := RestoreGame(currentState.GameState.(string))
	if err != nil {
		return nil, fmt.Errorf("failed to restore game: %w", err)
	}
	match := Match{
		Match: server.NewDefaultMatch(activeMatch.MatchId, players),
		cfg:   cfg,
		game:  game,
	}
	match.setTimer(cfg.CancelTimeout)
	matchHandler := NewMatchHandler(&match)
	match.SetHandler(matchHandler)
	return match, nil
}

func (h *MyServerHandler) OnHandleMessage(
	playerId string,
	matchHandler server.MatchHandler,
	message []byte,
) error {
	var payload Payload
	if err := json.Unmarshal(message, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	if payload.CreatedAt.Sub(time.Now()) > 2*time.Second {
		return fmt.Errorf("invalid timestamp")
	}
	switch payload.Type {
	case "gameData":
		move := NewMove(playerId)
		action := payload.Data["action"]
		switch action {
		case "abort":
			move.Control = ABORT
		case "resign":
			move.Control = RESIGN
		case "offerDraw":
			move.Control = OFFER_DRAW
		case "declineDraw":
			move.Control = DECLINE_DRAW
		case "move":
			move.Uci = payload.Data["move"]
			move.CreatedAt = payload.CreatedAt
		default:
			return fmt.Errorf("invalid game action: %s", payload.Type)
		}
		matchHandler.GetMatch().ProcessMove(move)
	default:
		return fmt.Errorf("invalid payload type: %s", payload.Type)
	}
	return nil
}

func (h *MyServerHandler) OnHandleMatchEnd(
	record *server.MatchRecordRequest,
	matchHandler server.MatchHandler,
) error {
	match := matchHandler.GetMatch().(*Match)
	record.Players = make([]server.PlayerRecord, 0, len(match.GetPlayers()))
	for _, player := range match.GetPlayers() {
		playerRecord := server.PlayerRecord{
			"PlayerId": player.GetId(),
		}
		record.Players = append(record.Players, playerRecord)
	}
	record.StartedAt = match.StartedAt
	record.EndedAt = time.Now()
	return nil
}

func (h *MyServerHandler) OnHandleMatchSave(
	matchState *dtos.MatchStateRequest,
	matchHandler server.MatchHandler,
) error {
	match := matchHandler.GetMatch().(*Match)
	lastMove := match.game.lastMove()
	matchState.PlayerStates = make([]dtos.PlayerStateRequest, 0, len(match.GetPlayers()))
	for _, player := range match.GetPlayers() {
		matchState.PlayerStates = append(matchState.PlayerStates, PlayerState{
			Id:     player.GetId(),
			Clock:  player.(*Player).Clock,
			Status: player.GetStatus(),
		})
	}
	matchState.Move = MoveRequest{
		PlayerId:  lastMove.GetPlayerId(),
		Uci:       lastMove.Uci,
		Control:   lastMove.Control,
		CreatedAt: lastMove.CreatedAt,
	}
	return nil
}

func NewServerHandler() server.ServerHandler {
	return &MyServerHandler{}
}
