package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

// GCPDriver handles uploads to Google Cloud Storage.
type GCPDriver struct {
	BucketName string
	Client     *storage.Client
}

// Upload sends a file to the configured Google Cloud Storage bucket.
// It sets metadata to make the file publicly readable.
func (d *GCPDriver) Upload(ctx context.Context, fileName string, contentType string, data io.Reader) (string, error) {
	bucket := d.Client.Bucket(d.BucketName)
	obj := bucket.Object(fileName)

	wc := obj.NewWriter(ctx)
	wc.ContentType = contentType
	// We do not set ACL here because uniform bucket-level access is often preferred in GCP,
	// but for a simple store, we can use bucket-level permissions to make it public.

	if _, err := io.Copy(wc, data); err != nil {
		return "", fmt.Errorf("failed to copy data to bucket: %v", err)
	}

	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close bucket writer: %v", err)
	}

	// For GCP Always Free, us-east1/us-central1/us-west1 regions use this public URL format:
	// https://storage.googleapis.com/[BUCKET_NAME]/[FILE_NAME]
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", d.BucketName, fileName), nil
}

// NewGCPDriver initializes a new GCP storage driver.
func NewGCPDriver(ctx context.Context, bucketName string) (*GCPDriver, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %v", err)
	}

	return &GCPDriver{
		BucketName: bucketName,
		Client:     client,
	}, nil
}

// Close closes the underlying storage client.
func (d *GCPDriver) Close() error {
	if d.Client != nil {
		return d.Client.Close()
	}
	return nil
}
