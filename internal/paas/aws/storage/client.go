package storage

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	dynamodb *dynamodb.Client
	s3       *s3.Client
	cfg      config
}

type config struct {
	PlatformsTableName   *string
	GamesTableName       *string
	DeploymentsTableName *string
	MainBucketName       *string
}

func NewClient(dynamoClient *dynamodb.Client, s3Client *s3.Client) *Client {
	return &Client{
		dynamodb: dynamoClient,
		s3:       s3Client,
		cfg:      loadConfig(),
	}
}

func loadConfig() config {
	var cfg config
	if v, ok := os.LookupEnv("PLATFORMS_TABLE_NAME"); ok {
		cfg.PlatformsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("GAMES_TABLE_NAME"); ok {
		cfg.GamesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("DEPLOYMENTS_TABLE_NAME"); ok {
		cfg.DeploymentsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("MAIN_BUCKET_NAME"); ok {
		cfg.MainBucketName = aws.String(v)
	}

	return cfg
}
