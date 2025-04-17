package dtos

import (
	"github.com/chess-vn/slchess/internal/paas/domains/entities"
)

type PlatformResponse struct {
	Id     string `json:"id"`
	UserId string `json:"userId"`
}

type PlatformListResponse struct {
	Items         []PlatformResponse     `json:"items"`
	NextPageToken *NextPlatformPageToken `json:"nextPageToken"`
}

type NextPlatformPageToken struct{}

func PlatformListResponseFromEntities(games []entities.Platform) PlatformListResponse {
	platformList := []PlatformResponse{}
	for _, game := range games {
		platformList = append(platformList, PlatformResponseFromEntity(game))
	}
	return PlatformListResponse{
		Items: platformList,
	}
}

func PlatformResponseFromEntity(platform entities.Platform) PlatformResponse {
	resp := PlatformResponse{
		Id:     platform.Id,
		UserId: platform.UserId,
	}
	return resp
}
