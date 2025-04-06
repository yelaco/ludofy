package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

var ErrEvaluationWorkNotFound = fmt.Errorf("evaluation work not found")

func (client *Client) SubmitEvaluationRequest(
	ctx context.Context,
	request dtos.EvaluationRequest,
) error {
	reqJson, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	_, err = client.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    client.cfg.EvaluationRequestQueueUrl,
		MessageBody: aws.String(string(reqJson)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (client *Client) AcquireEvaluationWork(ctx context.Context) (entities.EvaluationWork, error) {
	output, err := client.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            client.cfg.EvaluationRequestQueueUrl,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
		VisibilityTimeout:   60,
	})
	if err != nil {
		return entities.EvaluationWork{},
			fmt.Errorf("failed to receive message: %w", err)
	}
	if len(output.Messages) == 0 {
		return entities.EvaluationWork{}, ErrEvaluationWorkNotFound
	}

	var req dtos.EvaluationRequest
	err = json.Unmarshal([]byte(*output.Messages[0].Body), &req)
	if err != nil {
		return entities.EvaluationWork{},
			fmt.Errorf("failed to unmarshal message: %w", err)
	}

	evaluationWork := dtos.EvaluationWorkFromRequest(req)
	evaluationWork.ReceiptHandle = *output.Messages[0].ReceiptHandle

	return evaluationWork, nil
}

func (client *Client) RemoveEvaluationWork(ctx context.Context, receiptHandle string) error {
	_, err := client.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      client.cfg.EvaluationRequestQueueUrl,
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
