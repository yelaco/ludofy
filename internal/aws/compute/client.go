package compute

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

type Client struct {
	ecs        *ecs.Client
	ec2        *ec2.Client
	cloudwatch *cloudwatch.Client
	http       *http.Client
	cfg        config
}

type config struct {
	ClusterName *string
	TaskArn     *string
}

func NewClient(ecsClient *ecs.Client, ec2Client *ec2.Client, cloudwatchClient *cloudwatch.Client) *Client {
	return &Client{
		ecs:        ecsClient,
		ec2:        ec2Client,
		cloudwatch: cloudwatchClient,
		http:       new(http.Client),
		cfg:        loadConfig(),
	}
}

func loadConfig() config {
	var cfg config
	taskMetadata, err := getTaskMetadata()
	if err == nil {
		cfg.ClusterName = aws.String(taskMetadata.ClusterName)
		cfg.TaskArn = aws.String(taskMetadata.TaskArn)
	}

	return cfg
}
