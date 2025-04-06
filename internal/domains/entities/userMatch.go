package entities

type UserMatch struct {
	UserId  string `dynamodbav:"UserId"`
	MatchId string `dynamodbav:"MatchId"`
}
