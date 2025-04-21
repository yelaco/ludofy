package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

type DeployInput struct {
	StackName                     string                        `json:"stackName"`
	IncludeChatService            bool                          `json:"includeChatService"`
	IncludeFriendService          bool                          `json:"includeFriendService"`
	IncludeRankingService         bool                          `json:"includeRankingService"`
	IncludeMatchSpectatingService bool                          `json:"includeMatchSpectatingService"`
	MatchmakingConfiguration      MatchmakingConfigurationInput `json:"matchmakingConfiguration"`
	ServerConfiguration           ServerConfigurationInput      `json:"serverConfiguration"`
}

type MatchmakingConfigurationInput struct {
	MatchSize       int     `json:"matchSize"`
	RatingAlgorithm string  `json:"ratingAlgorithm"`
	InitialRating   float64 `json:"initialRating"`
}

type ServerConfigurationInput struct {
	ContainerImage ContainerImageInput `json:"containerImage"`
	MaxMatches     int                 `json:"maxMatches"`
	InitialCpu     float64             `json:"initialCpu"`
	InitialMemory  int                 `json:"initialMemory"`
}

type ContainerImageInput struct {
	Uri                 string                   `json:"uri"`
	IsPrivate           bool                     `json:"isPrivate"`
	RegistryCredentials RegistryCredentialsInput `json:"registryCredentials"`
}

type RegistryCredentialsInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DeploymentResponse struct {
	Id        string      `json:"id"`
	UserId    string      `json:"userId"`
	BackendId string      `json:"backendId"`
	Status    string      `json:"status"`
	Input     DeployInput `json:"input"`
	CreatedAt time.Time   `json:"createdAt"`
}

type DeploymentListResponse struct {
	Items         []DeploymentResponse     `json:"items"`
	NextPageToken *NextDeploymentPageToken `json:"nextPageToken"`
}

type NextDeploymentPageToken struct {
	Id string `json:"id"`
}

func DeploymentListResponseFromEntities(deployments []entities.Deployment) DeploymentListResponse {
	deploymentList := []DeploymentResponse{}
	for _, deployment := range deployments {
		deploymentList = append(deploymentList, DeploymentResponseFromEntity(deployment))
	}
	return DeploymentListResponse{
		Items: deploymentList,
	}
}

func DeploymentResponseFromEntity(deployment entities.Deployment) DeploymentResponse {
	return DeploymentResponse{
		Id:        deployment.Id,
		UserId:    deployment.UserId,
		BackendId: deployment.BackendId,
		Status:    deployment.Status,
		Input: DeployInput{
			StackName:                     deployment.Input.StackName,
			IncludeChatService:            deployment.Input.IncludeChatService,
			IncludeFriendService:          deployment.Input.IncludeFriendService,
			IncludeRankingService:         deployment.Input.IncludeRankingService,
			IncludeMatchSpectatingService: deployment.Input.IncludeMatchSpectatingService,
			MatchmakingConfiguration: MatchmakingConfigurationInput{
				MatchSize:       deployment.Input.MatchmakingConfiguration.MatchSize,
				RatingAlgorithm: deployment.Input.MatchmakingConfiguration.RatingAlgorithm,
				InitialRating:   deployment.Input.MatchmakingConfiguration.InitialRating,
			},
			ServerConfiguration: ServerConfigurationInput{
				ContainerImage: ContainerImageInput{
					Uri:       deployment.Input.ServerConfiguration.ContainerImage.Uri,
					IsPrivate: deployment.Input.ServerConfiguration.ContainerImage.IsPrivate,
					RegistryCredentials: RegistryCredentialsInput{
						Username: deployment.Input.ServerConfiguration.ContainerImage.RegistryCredentials.Username,
						Password: deployment.Input.ServerConfiguration.ContainerImage.RegistryCredentials.Password,
					},
				},
				MaxMatches:    deployment.Input.ServerConfiguration.MaxMatches,
				InitialCpu:    deployment.Input.ServerConfiguration.InitialCpu,
				InitialMemory: deployment.Input.ServerConfiguration.InitialMemory,
			},
		},
		CreatedAt: deployment.CreatedAt,
	}
}

func DeployInputRequestToEntity(input DeployInput) entities.DeployInput {
	return entities.DeployInput{
		StackName:                     input.StackName,
		IncludeChatService:            input.IncludeChatService,
		IncludeFriendService:          input.IncludeFriendService,
		IncludeRankingService:         input.IncludeRankingService,
		IncludeMatchSpectatingService: input.IncludeMatchSpectatingService,
		MatchmakingConfiguration: entities.MatchmakingConfigurationInput{
			MatchSize:       input.MatchmakingConfiguration.MatchSize,
			RatingAlgorithm: input.MatchmakingConfiguration.RatingAlgorithm,
			InitialRating:   input.MatchmakingConfiguration.InitialRating,
		},
		ServerConfiguration: entities.ServerConfigurationInput{
			ContainerImage: entities.ContainerImageInput{
				Uri:       input.ServerConfiguration.ContainerImage.Uri,
				IsPrivate: input.ServerConfiguration.ContainerImage.IsPrivate,
				RegistryCredentials: entities.RegistryCredentialsInput{
					Username: input.ServerConfiguration.ContainerImage.RegistryCredentials.Username,
					Password: input.ServerConfiguration.ContainerImage.RegistryCredentials.Password,
				},
			},
			MaxMatches:    input.ServerConfiguration.MaxMatches,
			InitialCpu:    input.ServerConfiguration.InitialCpu,
			InitialMemory: input.ServerConfiguration.InitialMemory,
		},
		ServerImageUri: "",
	}
}
