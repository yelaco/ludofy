package notification

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Client struct {
	sns *sns.Client
	cfg config
}

type config struct {
	PlatformApplicationArn *string
}

func NewClient(snsClient *sns.Client) *Client {
	return &Client{
		sns: snsClient,
		cfg: loadConfig(),
	}
}

func loadConfig() config {
	var cfg config
	if v, ok := os.LookupEnv("PLATFORM_APPLICATION_ARN"); ok {
		cfg.PlatformApplicationArn = aws.String(v)
	}
	return cfg
}
