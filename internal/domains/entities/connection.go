package entities

type Connection struct {
	Id     string `dynamodbav:"Id"`
	UserId string `dynamodbav:"UserId"`
}
