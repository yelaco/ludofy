package entities

type Platform struct {
	Id     string `dynamodbav:"Id"`
	UserId string `dynamodbav:"UserId"`
}
