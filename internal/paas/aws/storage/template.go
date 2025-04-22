package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func (client *Client) RemoveTemplates(ctx context.Context, prefix string) error {
	paginator := s3.NewListObjectsV2Paginator(client.s3, &s3.ListObjectsV2Input{
		Bucket: client.cfg.MainBucketName,
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		if len(page.Contents) == 0 {
			break
		}

		var objects []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
		}

		_, err = client.s3.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: client.cfg.MainBucketName,
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}
	}

	_, err := client.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: client.cfg.MainBucketName,
		Key:    aws.String(prefix),
	})
	if err != nil {
		return fmt.Errorf("failed to delete folder marker: %w", err)
	}

	return nil
}
