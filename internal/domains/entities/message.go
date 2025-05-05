package entities

import "time"

type Message struct {
	Id             string    `dynamodbav:"Id"`
	ConversationId string    `dynamodbav:"ConversationId"`
	SenderId       string    `dynamodbav:"SenderId"`
	Username       string    `dynamodbav:"Username"`
	Content        string    `dynamodbav:"Content"`
	CreatedAt      time.Time `dynamodbav:"CreatedAt"`
}
