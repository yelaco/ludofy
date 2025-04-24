package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsAuth "github.com/chess-vn/slchess/internal/aws/auth"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port          string
	MatchHandler  MatchHandler
	ServerHandler ServerHandler

	cognitoUserPoolId    string
	appSyncHttpUrl       string
	appSyncAccessRoleArn string
	abortGameFunctionArn string
	endGameFunctionArn   string
	maxMatches           int32
	protectionTimeout    time.Duration

	awsCfg            aws.Config
	cognitoPublicKeys map[string]*rsa.PublicKey
}

func NewConfig(port string, serverHandler ServerHandler, matchHandler MatchHandler) Config {
	viper.AutomaticEnv()
	protectionTimeout, err := time.ParseDuration(viper.GetString("SERVER_PROTECTION_TIMEOUT"))
	if err != nil {
		logging.Fatal("fatal error config file", zap.Error(err))
	}
	cfg := Config{
		Port:                 port,
		MatchHandler:         matchHandler,
		ServerHandler:        serverHandler,
		cognitoUserPoolId:    viper.GetString("COGNITO_USER_POOL_ID"),
		appSyncHttpUrl:       viper.GetString("APPSYNC_HTTP_URL"),
		appSyncAccessRoleArn: viper.GetString("APPSYNC_ACCESS_ROLE_ARN"),
		abortGameFunctionArn: viper.GetString("ABORT_GAME_FUNCTION_ARN"),
		endGameFunctionArn:   viper.GetString("END_GAME_FUNCTION_ARN"),
		maxMatches:           viper.GetInt32("MAX_MATCHES"),
		protectionTimeout:    protectionTimeout,
	}
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	tokenSigningKeyUrl := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		awsCfg.Region,
		cfg.cognitoUserPoolId,
	)
	cfg.cognitoPublicKeys, err = awsAuth.LoadCognitoPublicKeys(tokenSigningKeyUrl)
	if err != nil {
		panic(err)
	}

	err = cfg.loadAwsConfig(awsCfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *Config) loadAwsConfig(cfg aws.Config) error {
	ctx := context.Background()

	assumedCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			stscreds.NewAssumeRoleProvider(
				sts.NewFromConfig(cfg),
				c.appSyncAccessRoleArn,
			),
		),
	)
	if err != nil {
		return fmt.Errorf("unable to assume config: %w", err)
	}
	c.awsCfg = assumedCfg
	return nil
}
