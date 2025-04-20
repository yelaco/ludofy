package dtos

import (
	"time"

	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

type BackendResponse struct {
	Id        string            `json:"id"`
	UserId    string            `json:"userId"`
	StackName string            `json:"stackName"`
	Outputs   map[string]string `json:"outputs"`
	CreatedAt time.Time         `json:"createdAt"`
}

type BackendListResponse struct {
	Items         []BackendResponse     `json:"items"`
	NextPageToken *NextBackendPageToken `json:"nextPageToken"`
}

type NextBackendPageToken struct {
	Id string `json:"id"`
}

func BackendListResponseFromEntities(backends []entities.Backend) BackendListResponse {
	matchResultList := []BackendResponse{}
	for _, backend := range backends {
		matchResultList = append(matchResultList, BackendResponseFromEntity(backend))
	}
	return BackendListResponse{
		Items: matchResultList,
	}
}

func BackendResponseFromEntity(backend entities.Backend) BackendResponse {
	return BackendResponse{
		Id:        backend.Id,
		UserId:    backend.UserId,
		StackName: backend.StackName,
		CreatedAt: backend.CreatedAt,
	}
}
