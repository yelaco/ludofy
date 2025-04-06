package dtos

import "github.com/chess-vn/slchess/internal/domains/entities"

type ApplicationEndpointRequest struct {
	DeviceToken string `json:"deviceToken"`
}

type ApplicationEndpointResponse struct {
	UserId      string `json:"userId"`
	DeviceToken string `json:"deviceToken"`
	EndpointArn string `json:"endpointArn"`
}

func ApplicationEndpointResponseFromEntity(endpoint entities.ApplicationEndpoint) ApplicationEndpointResponse {
	return ApplicationEndpointResponse{
		UserId:      endpoint.UserId,
		DeviceToken: endpoint.DeviceToken,
		EndpointArn: endpoint.EndpointArn,
	}
}
