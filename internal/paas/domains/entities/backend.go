package entities

import "time"

type Backend struct {
	Id        string    `dynamodbav:"Id"`
	UserId    string    `dynamodbav:"UserId"`
	StackName string    `dynamodbav:"StackName"`
	Status    string    `dynamodbav:"Status"`
	CreatedAt time.Time `dynamodbav:"CreatedAt"`
}
