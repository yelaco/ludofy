package dtos

import (
	"github.com/yelaco/ludofy/internal/domains/entities"
)

type MatchSpectateResponse struct {
	MatchStates    MatchStateListResponse `json:"matchStates"`
	ConversationId string                 `json:"conversationId"`
}

func NewMatchSpectateResponse(matchStates []entities.MatchState, conversationId string) MatchSpectateResponse {
	return MatchSpectateResponse{
		MatchStates:    MatchStateListResponseFromEntities(matchStates),
		ConversationId: conversationId,
	}
}
