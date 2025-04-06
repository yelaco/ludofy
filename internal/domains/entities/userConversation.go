package entities

type UserConversation struct {
	UserId         string `dynamodbav:"UserId"`
	ConversationId string `dynamodbav:"ConversationId"`
	UnreadCount    int    `dynamodbav:"UnreadCount"`
	UpdatedAt      string `dynamodbav:"UpdatedAt"`
}
