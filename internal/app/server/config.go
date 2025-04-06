package server

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port        string
	IdleTimeout time.Duration

	AwsRegion            string
	CognitoUserPoolId    string
	AppSyncHttpUrl       string
	AppSyncAccessRoleArn string
	AbortGameFunctionArn string
	EndGameFunctionArn   string

	AwsCfg aws.Config
}

func NewConfig() Config {
	var cfg Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs/server")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	// List of env files to load
	envFiles := []string{
		"./configs/aws/base.env",
		"./configs/aws/cognito.env",
		"./configs/aws/lambda.env",
		"./configs/aws/appsync.env",
		"./configs/aws/dynamodb.env",
	}

	// Load into environment
	godotenv.Load(envFiles...)

	// Load into config struct
	err = loadEnvFiles(envFiles)
	if err != nil {
		logging.Fatal("fatal error config file", zap.Error(err))
	}

	cfg.Port = viper.GetString("Server.Port")
	idleTimeout, err := time.ParseDuration(viper.GetString("Server.IdleTimeout"))
	if err != nil {
		logging.Fatal("fatal error config file", zap.Error(err))
	}
	cfg.IdleTimeout = idleTimeout
	cfg.AwsRegion = viper.GetString("AWS_REGION")
	cfg.CognitoUserPoolId = viper.GetString("COGNITO_USER_POOL_ID")
	cfg.AppSyncHttpUrl = viper.GetString("APPSYNC_HTTP_URL")
	cfg.AppSyncAccessRoleArn = viper.GetString("APPSYNC_ACCESS_ROLE_ARN")
	cfg.AbortGameFunctionArn = viper.GetString("ABORT_GAME_FUNCTION_ARN")
	cfg.EndGameFunctionArn = viper.GetString("END_GAME_FUNCTION_ARN")

	if err := cfg.loadAwsConfig(); err != nil {
		logging.Fatal("failed to load aws config: %w", zap.Error(err))
	}

	return cfg
}

func loadEnvFiles(filenames []string) error {
	for _, file := range filenames {
		viper.SetConfigFile(file)
		viper.SetConfigType("env")
		viper.AutomaticEnv()

		err := viper.MergeInConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) loadAwsConfig() error {
	ctx := context.Background()
	defaultAwsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load default config: %w", err)
	}
	assumedCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			stscreds.NewAssumeRoleProvider(
				sts.NewFromConfig(defaultAwsCfg),
				c.AppSyncAccessRoleArn,
			),
		),
	)
	if err != nil {
		return fmt.Errorf("unable to assume config: %w", err)
	}
	c.AwsCfg = assumedCfg
	return nil
}
