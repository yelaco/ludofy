package dtos

type MatchAbortRequest struct {
	MatchId   string   `json:"matchId"`
	PlayerIds []string `json:"playerIds"`
}
