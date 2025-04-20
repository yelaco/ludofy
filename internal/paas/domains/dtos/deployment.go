package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

type DeployInput struct {
	StackName                     string                        `json:"stackName"`
	ServerImageUri                string                        `json:"serverImageUri"`
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
	InitialCpu    float64 `json:"initialCpu"`
	InitialMemory int     `json:"initialMemory"`
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
			ServerImageUri:                deployment.Input.ServerImageUri,
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
		ServerImageUri:                input.ServerImageUri,
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
			InitialCpu:    input.ServerConfiguration.InitialCpu,
			InitialMemory: input.ServerConfiguration.InitialMemory,
		},
	}
}
