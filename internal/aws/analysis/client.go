package analysis

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Client struct {
	athena *athena.Client
	sqs    *sqs.Client
	cfg    config
}

type config struct {
	AthenaDatabaseName         *string
	PuzzlesTableName           *string
	PuzzlesResultLocation      *string
	EvaluationRequestQueueName *string
	EvaluationRequestQueueUrl  *string
}

func NewClient(athenaClient *athena.Client, sqsClient *sqs.Client) *Client {
	return &Client{
		athena: athenaClient,
		sqs:    sqsClient,
		cfg:    loadConfig(),
	}
}

func loadConfig() config {
	var cfg config
	if v, ok := os.LookupEnv("ATHENA_DATABASE_NAME"); ok {
		cfg.AthenaDatabaseName = aws.String(v)
	}
	if v, ok := os.LookupEnv("PUZZLES_TABLE_NAME"); ok {
		cfg.PuzzlesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("PUZZLES_RESULT_LOCATION"); ok {
		cfg.PuzzlesResultLocation = aws.String(v)
	}
	if v, ok := os.LookupEnv("EVALUATION_REQUEST_QUEUE_NAME"); ok {
		cfg.EvaluationRequestQueueName = aws.String(v)
	}
	if v, ok := os.LookupEnv("EVALUATION_REQUEST_QUEUE_URL"); ok {
		cfg.EvaluationRequestQueueUrl = aws.String(v)
	}
	return cfg
}
