package entities

import "time"

type FriendRequest struct {
	SenderId   string    `dynamodb:"SenderId"`
	ReceiverId string    `dynamodb:"ReceiverId"`
	CreatedAt  time.Time `dynamodb:"CreatedAt"`
}
