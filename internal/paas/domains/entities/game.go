package entities

type Game struct {
	Id         string `dynamodbav:"Id"`
	PlatformId string `dynamodbav:"UserId"`
}
