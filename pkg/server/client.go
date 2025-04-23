package server

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/chess-vn/slchess/internal/aws/compute"
	"github.com/chess-vn/slchess/internal/aws/storage"
)

var (
	storageClient *storage.Client
	computeClient *compute.Client
	lambdaClient  *lambda.Client
)

func GetStorageClient() *storage.Client {
	return storageClient
}

func GetComputeClient() *compute.Client {
	return computeClient
}

func GetLambdaClient() *lambda.Client {
	return lambdaClient
}
