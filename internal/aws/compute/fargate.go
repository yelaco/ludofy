package compute

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/chess-vn/slchess/pkg/logging"
)

type TaskMetadata struct {
	TaskArn     string `json:"TaskARN"`
	ClusterName string `json:"Cluster"`
}

func (client *Client) GetServerIp(
	ctx context.Context,
	clusterName,
	serviceName string,
) (string, error) {
	// List tasks in the cluster
	listTasksOutput, err := client.ecs.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:       &clusterName,
		ServiceName:   &serviceName,
		DesiredStatus: "RUNNING",
	})
	if err != nil || len(listTasksOutput.TaskArns) == 0 {
		return "", fmt.Errorf("no running tasks found or error occurred: %v", err)
	}

	describeTasksOutput, err := client.ecs.DescribeTasks(
		ctx,
		&ecs.DescribeTasksInput{
			Cluster: aws.String(clusterName),
			Tasks:   listTasksOutput.TaskArns,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to describe ECS tasks: %w", err)
	}

	sort.Slice(describeTasksOutput.Tasks, func(i, j int) bool {
		return describeTasksOutput.Tasks[i].StartedAt.Before(*describeTasksOutput.Tasks[j].StartedAt)
	})

	var eniId string
	for _, detail := range describeTasksOutput.Tasks[0].Attachments[0].Details {
		if *detail.Name == "networkInterfaceId" {
			eniId = *detail.Value
			break
		}
	}

	if eniId == "" {
		return "", fmt.Errorf("ENI ID not found in task details")
	}

	// Get the public IP from EC2
	eniResp, err := client.ec2.DescribeNetworkInterfaces(
		ctx,
		&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []string{eniId},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to describe network interface: %w", err)
	}

	if len(eniResp.NetworkInterfaces) == 0 ||
		eniResp.NetworkInterfaces[0].Association == nil {
		return "", fmt.Errorf("no public IP found for ENI")
	}

	return *eniResp.NetworkInterfaces[0].Association.PublicIp, nil
}

func (client *Client) CheckAndGetNewServerIp(
	ctx context.Context,
	clusterName,
	serviceName,
	targetPublicIp string,
) (string, error) {
	// List tasks in the cluster
	listTasksOutput, err := client.ecs.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:       &clusterName,
		ServiceName:   &serviceName,
		DesiredStatus: "RUNNING",
	})
	if err != nil || len(listTasksOutput.TaskArns) == 0 {
		return "", fmt.Errorf("no running tasks found or error occurred: %v", err)
	}

	describeTasksOutput, err := client.ecs.DescribeTasks(
		ctx,
		&ecs.DescribeTasksInput{
			Cluster: aws.String(clusterName),
			Tasks:   listTasksOutput.TaskArns,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to describe ECS tasks: %w", err)
	}

	sort.Slice(describeTasksOutput.Tasks, func(i, j int) bool {
		return describeTasksOutput.Tasks[i].StartedAt.Before(*describeTasksOutput.Tasks[j].StartedAt)
	})

	var newServerIp string
	for i, task := range describeTasksOutput.Tasks {
		for _, attachment := range task.Attachments {
			for _, detail := range attachment.Details {
				if *detail.Name == "networkInterfaceId" {
					eniID := *detail.Value

					eniOutput, err := client.ec2.DescribeNetworkInterfaces(
						ctx,
						&ec2.DescribeNetworkInterfacesInput{
							NetworkInterfaceIds: []string{eniID},
						},
					)
					if err != nil {
						return "", fmt.Errorf("failed to describe ENI: %w", err)
					}

					for _, eni := range eniOutput.NetworkInterfaces {
						if eni.Association != nil && eni.Association.PublicIp != nil {
							if *eni.Association.PublicIp == targetPublicIp {
								return targetPublicIp, nil
							}
							if i == 0 {
								newServerIp = *eni.Association.PublicIp
							}
						}
					}
				}
			}
		}
	}

	return newServerIp, nil
}

func (client *Client) CheckAndStartTask(
	ctx context.Context,
	clusterName,
	serviceName string,
) error {
	// Check running task count
	listTasksOutput, err := client.ecs.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:       aws.String(clusterName),
		ServiceName:   aws.String(serviceName),
		DesiredStatus: "RUNNING",
	})
	if err != nil {
		return fmt.Errorf("failed to list ECS tasks: %w", err)
	}

	// If no tasks are running, scale service to 1
	if len(listTasksOutput.TaskArns) == 0 {
		logging.Info("No running tasks found. Scaling up ECS service...")

		_, err := client.ecs.UpdateService(ctx, &ecs.UpdateServiceInput{
			Cluster:      aws.String(clusterName),
			Service:      aws.String(serviceName),
			DesiredCount: aws.Int32(1),
		})
		if err != nil {
			return fmt.Errorf("failed to update ECS desired count: %w", err)
		}
	}

	return nil
}

func (client *Client) UpdateServerProtection(
	ctx context.Context,
	enabled bool,
) error {
	if client.cfg.ClusterName == nil || client.cfg.TaskArn == nil {
		return fmt.Errorf("missing task metadata")
	}
	_, err := client.ecs.UpdateTaskProtection(ctx, &ecs.UpdateTaskProtectionInput{
		Cluster:           client.cfg.ClusterName,
		Tasks:             []string{*client.cfg.TaskArn},
		ProtectionEnabled: enabled,
	})
	if err != nil {
		return fmt.Errorf("failed to update task protection: %w", err)
	}
	return nil
}
