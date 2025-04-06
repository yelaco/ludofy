package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type FriendshipListResponse struct {
	Items         []FriendshipResponse     `json:"items"`
	NextPageToken *NextFriendshipPageToken `json:"nextPageToken"`
}

type FriendshipResponse struct {
	UserId         string    `json:"userId"`
	FriendId       string    `json:"friendId"`
	ConversationId string    `json:"conversationId"`
	StartedAt      time.Time `json:"startedAt"`
}

type NextFriendshipPageToken struct {
	FriendId string `json:"friendId"`
}

func FriendshipListResponseFromEntities(friendships []entities.Friendship) FriendshipListResponse {
	friendshipList := []FriendshipResponse{}
	for _, friendship := range friendships {
		friendshipList = append(friendshipList, FriendshipResponseFromEntity(friendship))
	}
	return FriendshipListResponse{
		Items: friendshipList,
	}
}

func FriendshipResponseFromEntity(friendship entities.Friendship) FriendshipResponse {
	return FriendshipResponse{
		UserId:         friendship.UserId,
		FriendId:       friendship.FriendId,
		ConversationId: friendship.ConversationId,
		StartedAt:      friendship.StartedAt,
	}
}
