package entities

type Deployment struct {
	Id        string `dynamodbav:"Id"`
	UserId    string `dynamodbav:"UserId"`
	StackName string `dynamodbav:"StackName"`
	Status    string `dynamodbav:"Status"`
}
