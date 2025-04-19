package entities

type Deployment struct {
	Id        string `dynamodbav:"Id"`
	UserId    string `dynamodbav:"UserId"`
	BackendId string `dynamodbav:"BackendId"`
	StackName string `dynamodbav:"StackName"`
	Status    string `dynamodbav:"Status"`
}
