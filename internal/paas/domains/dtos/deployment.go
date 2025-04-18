package dtos

type DeploymentRequest struct {
	UserId   string          `json:"userId"`
	Platform PlatformRequest `json:"platform"`
	Games    []GameRequest   `json:"games"`
}

type PlatformRequest struct {
	Name                 string `json:"name"`
	IncludeChatService   bool   `json:"includeChatService"`
	IncludeFriendService bool   `json:"includeFriendService"`
}

type GameRequest struct {
	Name                          string `json:"name"`
	ServerImageUri                string `json:"serverImageUri"`
	IncludeRankingService         bool   `json:"includeRankingService"`
	IncludeMatchSpectatingService bool   `json:"includeMatchSpectatingService"`
}
