package compute

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/chess-vn/slchess/internal/domains/dtos"
)

func (client *Client) GetServiceMetrics(
	ctx context.Context,
	startTime,
	endTime time.Time,
	interval int32,
	clusterName,
	serviceName string,
) (
	[]dtos.ServiceMetrics,
	error,
) {
	getMetric := func(metricName string) ([]types.Datapoint, error) {
		resp, err := client.cloudwatch.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/ECS"),
			MetricName: aws.String(metricName),
			Dimensions: []types.Dimension{
				{Name: aws.String("ClusterName"), Value: aws.String(clusterName)},
				{Name: aws.String("ServiceName"), Value: aws.String(serviceName)},
			},
			Period:     aws.Int32(interval),
			StartTime:  aws.Time(startTime),
			EndTime:    aws.Time(endTime),
			Statistics: []types.Statistic{types.StatisticAverage},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", metricName, err)
		}
		return resp.Datapoints, nil
	}

	cpuPoints, err := getMetric("CPUUtilization")
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu metric: %w", err)
	}
	memPoints, err := getMetric("MemoryUtilization")
	if err != nil {
		return nil, fmt.Errorf("failed to get memory metric: %w", err)
	}

	metrics := make([]dtos.ServiceMetrics, 0, len(cpuPoints))

	for i := range len(cpuPoints) {
		metrics = append(metrics, dtos.ServiceMetrics{
			CPUAvg:    aws.ToFloat64(cpuPoints[i].Average),
			MemAvg:    aws.ToFloat64(memPoints[i].Average),
			Timestamp: aws.ToTime(cpuPoints[i].Timestamp),
		})
	}

	return metrics, nil
}
