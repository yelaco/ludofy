package dtos

import "github.com/chess-vn/slchess/internal/domains/entities"

type MatchResultListResponse struct {
	Items         []MatchResultResponse     `json:"items"`
	NextPageToken *NextMatchResultPageToken `json:"nextPageToken"`
}

type MatchResultResponse struct {
	UserId         string  `json:"userId"`
	MatchId        string  `json:"matchId"`
	OpponentId     string  `json:"opponentId"`
	OpponentRating float64 `json:"opponentRating"`
	OpponentRD     float64 `json:"opponentRD"`
	Result         float64 `json:"result"`
	Timestamp      string  `json:"timestamp"`
}

type NextMatchResultPageToken struct {
	Timestamp string `json:"timestamp"`
}

func MatchResultListResponseFromEntities(matchResults []entities.MatchResult) MatchResultListResponse {
	matchResultList := []MatchResultResponse{}
	for _, matchResult := range matchResults {
		matchResultList = append(matchResultList, MatchResultResponseFromEntity(matchResult))
	}
	return MatchResultListResponse{
		Items: matchResultList,
	}
}

func MatchResultResponseFromEntity(matchResult entities.MatchResult) MatchResultResponse {
	return MatchResultResponse{
		UserId:         matchResult.UserId,
		MatchId:        matchResult.MatchId,
		OpponentId:     matchResult.OpponentId,
		OpponentRating: matchResult.OpponentRating,
		OpponentRD:     matchResult.OpponentRD,
		Result:         matchResult.Result,
		Timestamp:      matchResult.Timestamp,
	}
}
