package storage

import (
	"context"
	"io"
)

// StorageDriver defines the interface for file uploads to cloud providers.
type StorageDriver interface {
	// Upload takes a file, filename, and content type and returns the public URL or error.
	Upload(ctx context.Context, fileName string, contentType string, data io.Reader) (string, error)
}
