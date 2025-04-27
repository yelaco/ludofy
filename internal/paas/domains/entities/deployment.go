package entities

import "time"

type Deployment struct {
	Id        string      `dynamodbav:"Id"`
	UserId    string      `dynamodbav:"UserId"`
	BackendId string      `dynamodbav:"BackendId"`
	Status    string      `dynamodbav:"Status"`
	Input     DeployInput `dynamodbav:"Input"`
	CreatedAt time.Time   `dynamodbav:"CreatedAt"`
}

type DeployInput struct {
	StackName                     string                        `dynamodbav:"StackName"`
	IncludeChatService            bool                          `dynamodbav:"IncludeChatService"`
	IncludeFriendService          bool                          `dynamodbav:"IncludeFriendService"`
	IncludeRankingService         bool                          `dynamodbav:"IncludeRankingService"`
	IncludeMatchSpectatingService bool                          `dynamodbav:"IncludeMatchSpectatingService"`
	UseCustomization              bool                          `dynamodbav:"UseCustomization"`
	MatchmakingConfiguration      MatchmakingConfigurationInput `dynamodbav:"MatchmakingConfiguration"`
	ServerConfiguration           ServerConfigurationInput      `dynamodbav:"ServerConfiguration"`
}

type MatchmakingConfigurationInput struct {
	MatchSize       int     `dynamodbav:"MatchSize"`
	RatingAlgorithm string  `dynamodbav:"RatingAlgorithm"`
	InitialRating   float64 `dynamodbav:"InitialRating"`
}

type ServerConfigurationInput struct {
	ContainerImage ContainerImageInput `dynamodbav:"ContainerImage"`
	MaxMatches     int                 `dynamodbav:"MaxMatches"`
	InitialCpu     float64             `dynamodbav:"InitialCpu"`
	InitialMemory  int                 `dynamodbav:"InitialMemory"`
}

type ContainerImageInput struct {
	Uri                 string                   `dynamodbav:"Uri"`
	IsPrivate           bool                     `dynamodbav:"IsPrivate"`
	RegistryCredentials RegistryCredentialsInput `dynamodbav:"RegistryCredentials"`
}

type RegistryCredentialsInput struct {
	Username string `dynamodbav:"Username"`
	Password string `dynamodbav:"Password"`
}
