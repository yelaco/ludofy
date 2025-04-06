package entities

import "time"

type Friendship struct {
	UserId         string    `dynamodb:"UserId"`
	FriendId       string    `dynamodb:"FriendId"`
	ConversationId string    `dynamodb:"ConversationId"`
	StartedAt      time.Time `dynamodb:"StartedAt"`
}
