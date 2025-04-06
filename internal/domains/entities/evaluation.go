package entities

type Pv struct {
	Cp    int    `dynamodbav:"Cp"`
	Moves string `dynamodbav:"Moves"`
}

type Evaluation struct {
	Fen    string `dynamodbav:"Fen"`
	Depth  int    `dynamodbav:"Depth"`
	Knodes int    `dynamodbav:"Knodes"`
	Pvs    []Pv   `dynamodbav:"Pvs"`
}

type EvaluationWork struct {
	ConnectionId  string
	Fen           string
	ReceiptHandle string
}
