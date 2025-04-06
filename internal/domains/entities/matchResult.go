package entities

type MatchResult struct {
	UserId         string  `dynamodbav:"UserId"`
	MatchId        string  `dynamodbav:"MatchId"`
	OpponentId     string  `dynamodbav:"OpponentId"`
	OpponentRating float64 `dynamodbav:"OpponentRating"`
	OpponentRD     float64 `dynamodbav:"OpponentRD"`
	Result         float64 `dynamodbav:"Result"`
	Timestamp      string  `dynamodbav:"Timestamp"`
}
