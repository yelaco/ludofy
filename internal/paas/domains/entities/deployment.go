package entities

type Deployment struct {
	Id     string `dynamodbav:"Id"`
	UserId string `dynamodbav:"UserId"`
}
