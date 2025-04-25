package main

import "time"

type PlayerRecord struct {
	Id string `json:"id"`
}

func (pr PlayerRecord) GetPlayerId() string {
	return pr.Id
}

type PlayerState struct {
	Id     string        `json:"id"`
	Clock  time.Duration `json:"clocks"`
	Status string        `json:"status"`
}

func (ps PlayerState) GetPlayerId() string {
	return ps.Id
}

type MoveRequest struct {
	PlayerId  string    `json:"playerId"`
	Uci       string    `json:"uci"`
	Control   string    `json:"control"`
	CreatedAt time.Time `json:"createdAt"`
}

func (mr MoveRequest) GetPlayerId() string {
	return mr.PlayerId
}
