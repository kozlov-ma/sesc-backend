package filesvc

import (
	"context"
	"io"
	"time"
)

// ObjectStorage defines the interface for storage operations.
// This allows the filesvc package to be independent of specific storage implementations.
type ObjectStorage interface {
	// PutObject uploads an object to storage
	PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64) error

	// RemoveObject deletes an object from storage
	RemoveObject(ctx context.Context, objectKey string) error

	// GetObjectURL returns a URL that can be used to download the object
	GetObjectURL(ctx context.Context, objectKey string, downloadName string, expires time.Duration) (string, error)
}
