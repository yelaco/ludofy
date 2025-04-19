package dtos

import "github.com/chess-vn/slchess/internal/paas/domains/entities"

type DeployInput struct {
	StackName                     string `json:"stackName"`
	ServerImageUri                string `json:"serverImageUri"`
	IncludeChatService            bool   `json:"includeChatService"`
	IncludeFriendService          bool   `json:"includeFriendService"`
	IncludeRankingService         bool   `json:"includeRankingService"`
	IncludeMatchSpectatingService bool   `json:"includeMatchSpectatingService"`
}

type DeploymentResponse struct {
	Id        string `json:"id"`
	UserId    string `json:"userId"`
	StackName string `json:"stackName"`
	Status    string `json:"status"`
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
		StackName: deployment.StackName,
		Status:    deployment.Status,
	}
}
