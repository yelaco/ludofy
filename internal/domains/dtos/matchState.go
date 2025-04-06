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
	PlayerStates []PlayerStateRequest `json:"playerStates"`
	GameState    string               `json:"gameState"`
	Move         MoveRequest          `json:"move"`
	Ply          int                  `json:"ply"`
	Timestamp    time.Time            `json:"timestamp"`
}

type PlayerStateRequest struct {
	Clock  string `json:"clock"`
	Status string `json:"status"`
}

type MoveRequest struct {
	PlayerId string `json:"playerId"`
	Uci      string `json:"uci"`
}

type MatchStateAppSyncRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type PlayerStateResponse struct {
	Clock  string `json:"clock"`
	Status string `json:"status"`
}

type MoveResponse struct {
	PlayerId string `json:"playerId"`
	Uci      string `json:"uci"`
}

type MatchStateResponse struct {
	Id           string                `json:"id"`
	MatchId      string                `json:"matchId"`
	PlayerStates []PlayerStateResponse `json:"playerStates"`
	GameState    string                `json:"gameState"`
	Move         MoveResponse          `json:"move"`
	Ply          int                   `json:"ply"`
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
	return entities.MatchState{
		Id:      req.Id,
		MatchId: req.MatchId,
		PlayerStates: []entities.PlayerState{
			{
				Clock:  req.PlayerStates[0].Clock,
				Status: req.PlayerStates[0].Status,
			},
			{
				Clock:  req.PlayerStates[1].Clock,
				Status: req.PlayerStates[1].Status,
			},
		},
		GameState: req.GameState,
		Move: entities.Move{
			PlayerId: req.Move.PlayerId,
			Uci:      req.Move.Uci,
		},
		Ply:       req.Ply,
		Timestamp: req.Timestamp,
	}
}

func MatchStateResponseFromEntitiy(matchState entities.MatchState) MatchStateResponse {
	return MatchStateResponse{
		Id:      matchState.Id,
		MatchId: matchState.MatchId,
		PlayerStates: []PlayerStateResponse{
			{
				Clock:  matchState.PlayerStates[0].Clock,
				Status: matchState.PlayerStates[0].Status,
			},
			{
				Clock:  matchState.PlayerStates[1].Clock,
				Status: matchState.PlayerStates[1].Status,
			},
		},
		GameState: matchState.GameState,
		Move: MoveResponse{
			PlayerId: matchState.Move.PlayerId,
			Uci:      matchState.Move.Uci,
		},
		Ply:       matchState.Ply,
		Timestamp: matchState.Timestamp,
	}
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
