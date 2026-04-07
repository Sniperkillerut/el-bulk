package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Driver handles uploads to AWS S3.
type S3Driver struct {
	BucketName string
	Region     string
	Client     *s3.Client
}

// Upload sends a file to the configured S3 bucket.
// It sets public-read ACL to ensure the file is accessible via URL.
func (d *S3Driver) Upload(ctx context.Context, fileName string, contentType string, data io.Reader) (string, error) {
	_, err := d.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(d.BucketName),
		Key:         aws.String(fileName),
		Body:        data,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead, // For easy storefront display
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	// Canonical S3 public URL: https://[BUCKET].s3.[REGION].amazonaws.com/[FILENAME]
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", d.BucketName, d.Region, fileName), nil
}

// NewS3Driver initializes a new AWS S3 storage driver.
func NewS3Driver(ctx context.Context, bucketName string, region string) (*S3Driver, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Driver{
		BucketName: bucketName,
		Region:     region,
		Client:     client,
	}, nil
}
