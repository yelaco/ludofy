package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type FriendRequestListResponse struct {
	Items         []FriendRequestResponse     `json:"items"`
	NextPageToken *NextFriendRequestPageToken `json:"nextPageToken"`
}

type FriendRequestResponse struct {
	SenderId   string    `json:"senderId"`
	ReceiverId string    `json:"receiverId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type NextFriendRequestPageToken struct {
	SenderId   string `json:"senderId,omitempty"`
	ReceiverId string `json:"receiverId,omitempty"`
}

func FriendRequestListResponseFromEntities(friendships []entities.FriendRequest) FriendRequestListResponse {
	friendRequestList := []FriendRequestResponse{}
	for _, friendRequest := range friendships {
		friendRequestList = append(friendRequestList, FriendRequestResponseFromEntity(friendRequest))
	}
	return FriendRequestListResponse{
		Items: friendRequestList,
	}
}

func FriendRequestResponseFromEntity(friendRequest entities.FriendRequest) FriendRequestResponse {
	return FriendRequestResponse{
		SenderId:   friendRequest.SenderId,
		ReceiverId: friendRequest.ReceiverId,
		CreatedAt:  friendRequest.CreatedAt,
	}
}
