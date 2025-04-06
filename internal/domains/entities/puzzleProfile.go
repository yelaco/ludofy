package entities

type PuzzleProfile struct {
	UserId string  `dynamodbav:"UserId"`
	Rating float64 `dynamodbav:"Rating"`
}
