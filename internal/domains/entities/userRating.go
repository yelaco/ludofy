package entities

type UserRating struct {
	UserId       string  `dynamodbav:"UserId"`
	PartitionKey string  `dynamodbav:"PartitionKey"`
	Rating       float64 `dynamodbav:"Rating"`
	RD           float64 `dynamodbav:"RD,omitempty"`
	Sigma        float64 `dynamodbav:"Sigma,omitempty"`
}
