package dtos

type BackendDeployEvent struct {
	JobName  string `json:"jobName"`
	JobId    string `json:"jobId"`
	JobQueue string `json:"jobQueue"`
	Status   string `json:"status"`
}
