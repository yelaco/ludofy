package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chess-vn/slchess/internal/aws/storage"
	"github.com/disintegration/imaging"
)

type size int

const (
	SMALL  size = 50
	MEDIUM size = 200
	LARGE  size = 500
)

var (
	cfg           aws.Config
	s3Client      *s3.Client
	storageClient *storage.Client

	avatarBucketName = os.Getenv("AVATAR_BUCKET_NAME")
	sizes            = []size{SMALL, MEDIUM, LARGE}
)

func init() {
	cfg, _ = config.LoadDefaultConfig(context.TODO())
	s3Client = s3.NewFromConfig(cfg)
	storageClient = storage.NewClient(dynamodb.NewFromConfig(cfg))
}

func handle(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		userId, found := strings.CutPrefix(key, "avatars/")
		if !found {
			return fmt.Errorf("prefix not found")
		}

		// Download the image
		resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("failed to download image: %v", err)
		}
		defer resp.Body.Close()

		// Decode image
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to decode image: %v", err)
		}

		// Put image
		for _, size := range sizes {
			resizedImg := imaging.Resize(img, 0, int(size), imaging.Lanczos)
			err := PutImage(ctx, resizedImg, size, avatarBucketName, userId)
			if err != nil {
				return fmt.Errorf("failed to put image: %v", err)
			}
		}

		// Update user profile in DynamoDB
		avatarUrl := fmt.Sprintf(
			"https://%s.s3.%s.amazonaws.com/%s",
			avatarBucketName,
			cfg.Region,
			userId,
		)
		err = storageClient.UpdateUserProfile(
			ctx,
			userId,
			storage.UserProfileUpdateOptions{
				Avatar: aws.String(avatarUrl),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to update user profile: %v", err)
		}

		// Delete original image
		_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("failed to delete original image: %v", err)
		}
	}

	return nil
}

func PutImage(
	ctx context.Context,
	img *image.NRGBA,
	size size,
	bucket string,
	userId string,
) error {
	// Encode as JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	// Upload back to S3
	newKey := fmt.Sprintf("%s/%s", userId, size.String())
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(newKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload processed image: %v", err)
	}

	return nil
}

func (s size) String() string {
	switch s {
	case SMALL:
		return "small"
	case MEDIUM:
		return "medium"
	case LARGE:
		return "large"
	}
	return ""
}

func main() {
	lambda.Start(handle)
}
