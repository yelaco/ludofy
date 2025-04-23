package dtos

import (
	_ "embed"
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

//go:embed graphql/updateMatchState.graphql
var updateMatchStateMutation string

type MatchStateRequest struct {
	Id           string               `json:"id"`
	MatchId      string               `json:"matchId"`
	PlayerStates []PlayerStateRequest `json:"players"`
	GameState    interface{}          `json:"game"`
	Move         MoveRequest          `json:"move"`
	Timestamp    time.Time            `json:"timestamp"`
}

type PlayerStateRequest interface {
	GetPlayerId() string
}

type MoveRequest interface {
	GetPlayerId() string
}

type MatchStateAppSyncRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type PlayerStateResponse interface {
	GetPlayerId() string
}

type MoveResponse interface {
	GetPlayerId() string
}

type MatchStateResponse struct {
	Id           string                `json:"id"`
	MatchId      string                `json:"matchId"`
	PlayerStates []PlayerStateResponse `json:"players"`
	GameState    any                   `json:"game"`
	Move         MoveResponse          `json:"move"`
	Timestamp    time.Time             `json:"timestamp"`
}

type MatchStateListResponse struct {
	Items         []MatchStateResponse     `json:"items"`
	NextPageToken *NextMatchStatePageToken `json:"nextPageToken"`
}

type NextMatchStatePageToken struct {
	Id  string `json:"id"`
	Ply string `json:"ply"`
}

func NewMatchStateAppSyncRequest(req MatchStateRequest) MatchStateAppSyncRequest {
	return MatchStateAppSyncRequest{
		Query: updateMatchStateMutation,
		Variables: map[string]interface{}{
			"input": req,
		},
	}
}

func MatchStateRequestToEntity(req MatchStateRequest) entities.MatchState {
	matchState := entities.MatchState{
		Id:           req.Id,
		MatchId:      req.MatchId,
		PlayerStates: make([]entities.PlayerState, 0, len(req.PlayerStates)),
		GameState:    req.GameState,
		Move:         req.Move,
		Timestamp:    req.Timestamp,
	}
	for _, playerState := range req.PlayerStates {
		matchState.PlayerStates = append(matchState.PlayerStates, playerState)
	}
	return matchState
}

func MatchStateResponseFromEntitiy(matchState entities.MatchState) MatchStateResponse {
	resp := MatchStateResponse{
		Id:        matchState.Id,
		MatchId:   matchState.MatchId,
		GameState: matchState.GameState,
		Move:      matchState.Move,
		Timestamp: matchState.Timestamp,
	}
	for _, playerState := range matchState.PlayerStates {
		resp.PlayerStates = append(resp.PlayerStates, playerState)
	}
	return resp
}

func MatchStateListResponseFromEntities(matchStates []entities.MatchState) MatchStateListResponse {
	matchStateList := []MatchStateResponse{}
	for _, matchState := range matchStates {
		matchStateList = append(matchStateList, MatchStateResponseFromEntitiy(matchState))
	}
	return MatchStateListResponse{
		Items: matchStateList,
	}
}
