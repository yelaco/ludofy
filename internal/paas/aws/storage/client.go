package storage

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Client struct {
	dynamodb *dynamodb.Client
	cfg      config
}

type config struct {
	PlatformsTableName   *string
	GamesTableName       *string
	DeploymentsTableName *string
}

func NewClient(dynamoClient *dynamodb.Client) *Client {
	return &Client{
		dynamodb: dynamoClient,
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

	return cfg
}
