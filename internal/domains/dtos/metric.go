package dtos

import "time"

type ServiceMetrics struct {
	CPUAvg    float64   `json:"cpuAvg"`
	MemAvg    float64   `json:"memAvg"`
	Timestamp time.Time `json:"timestamp"`
}

type ServerMetricsResponse struct {
	ActiveMatches int32 `json:"activeMatches"`
	CanAccept     bool  `json:"canAccept"`
	MaxMatches    int32 `json:"maxMatches"`
}

type BackendMetricsResponse struct {
	ServiceMetrics    []ServiceMetrics        `json:"serviceMetricsList"`
	ServerMetricsList []ServerMetricsResponse `json:"serverMetricsList"`
}
