package dtos

type BackendDeployEvent struct {
	JobName      string `json:"jobName"`
	JobId        string `json:"jobId"`
	JobQueue     string `json:"jobQueue"`
	DeploymentId string `json:"deploymentId"`
}
