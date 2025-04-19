package entities

type Backend struct {
	Id        string `dynamodbav:"Id"`
	UserId    string `dynamodbav:"UserId"`
	StackName string `dynamodbav:"StackName"`
}
