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

func (client *Client) GetServiceMetrics(ctx context.Context, clusterName, serviceName string) (dtos.ServiceMetrics, error) {
	getMetric := func(metricName string) (float64, time.Time, error) {
		resp, err := client.cloudwatch.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/ECS"),
			MetricName: aws.String(metricName),
			Dimensions: []types.Dimension{
				{Name: aws.String("ClusterName"), Value: aws.String(clusterName)},
				{Name: aws.String("ServiceName"), Value: aws.String(serviceName)},
			},
			Period:     aws.Int32(60),
			StartTime:  aws.Time(time.Now().Add(-5 * time.Minute)),
			EndTime:    aws.Time(time.Now()),
			Statistics: []types.Statistic{types.StatisticAverage},
		})
		if err != nil {
			return 0, time.Time{}, fmt.Errorf("failed to get %s: %w", metricName, err)
		}
		if len(resp.Datapoints) == 0 {
			return 0, time.Time{}, fmt.Errorf("no datapoints for %s", metricName)
		}
		latest := resp.Datapoints[0]
		for _, dp := range resp.Datapoints {
			if dp.Timestamp.After(aws.ToTime(latest.Timestamp)) {
				latest = dp
			}
		}
		return *latest.Average, *latest.Timestamp, nil
	}

	cpu, ts, err := getMetric("CPUUtilization")
	if err != nil {
		return dtos.ServiceMetrics{}, err
	}
	mem, _, err := getMetric("MemoryUtilization")
	if err != nil {
		return dtos.ServiceMetrics{}, err
	}

	return dtos.ServiceMetrics{
		Timestamp: ts,
		CPUAvg:    cpu,
		MemAvg:    mem,
	}, nil
}
