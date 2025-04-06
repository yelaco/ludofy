package entities

type SpectatorConversation struct {
	MatchId        string `dynamodbav:"MatchId"`
	ConversationId string `dynamodbav:"ConversationId"`
}
