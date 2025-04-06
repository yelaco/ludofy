package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type MessageListResponse struct {
	Items         []MessageResponse     `json:"items"`
	NextPageToken *NextMessagePageToken `json:"nextPageToken"`
}

type MessageResponse struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversationId"`
	SenderId       string    `json:"senderId"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}

type NextMessagePageToken struct {
	CreatedAt string `json:"createdAt"`
}

func MessageListResponseFromEntities(message []entities.Message) MessageListResponse {
	messageList := []MessageResponse{}
	for _, matchResult := range message {
		messageList = append(messageList, MessageResponseFromEntity(matchResult))
	}
	return MessageListResponse{
		Items: messageList,
	}
}

func MessageResponseFromEntity(message entities.Message) MessageResponse {
	return MessageResponse{
		Id:             message.Id,
		ConversationId: message.ConversationId,
		SenderId:       message.SenderId,
		Content:        message.Content,
		CreatedAt:      message.CreatedAt,
	}
}
