package server

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/yelaco/ludofy/internal/aws/compute"
	"github.com/yelaco/ludofy/internal/aws/storage"
)

var (
	storageClient *storage.Client
	computeClient *compute.Client
	lambdaClient  *lambda.Client
)
