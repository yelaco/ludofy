package entities

import "time"

type UserProfile struct {
	UserId     string    `dynamodbav:"UserId"`
	Username   string    `dynamodbav:"Username"`
	Avatar     string    `dynamodbav:"Avatar"`
	Phone      string    `dynamodbav:"Phone"`
	Locale     string    `dynamodbav:"Locale"`
	Membership string    `dynamodbav:"Membership"`
	CreatedAt  time.Time `dynamodbav:"CreatedAt"`
}
