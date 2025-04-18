package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (client *Client) UploadTemplate(ctx context.Context, key string, reader io.Reader) error {
	_, err := client.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: client.cfg.MainBucketName,
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}
