package s3

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps a MinIO client for S3 operations.
type Client struct {
	minio    *minio.Client
	bucket   string
	logger   *slog.Logger
	endpoint string
}

// ListObjects returns a list of object keys in the bucket with the given prefix.
func (c *Client) ListObjects(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	}
	var keys []string
	for obj := range c.minio.ListObjects(ctx, c.bucket, opts) {
		if obj.Err != nil {
			c.logger.Error("error listing object", "error", obj.Err)
			return nil, fmt.Errorf("list error: %w", obj.Err)
		}
		keys = append(keys, obj.Key)
	}
	return keys, nil
}

// New initializes a new S3 client.
// It reads configuration from environment variables:
// MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET, MINIO_USE_SSL.
func New(logger *slog.Logger) (*Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT not set")
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY not set")
	}
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("MINIO_SECRET_KEY not set")
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("MINIO_BUCKET not set")
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	// Create bucket if it does not exist
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("could not check bucket: %w", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("could not create bucket: %w", err)
		}
		logger.Info("created bucket", "bucket", bucket)
	}

	return &Client{
		minio:    minioClient,
		bucket:   bucket,
		logger:   logger,
		endpoint: endpoint,
	}, nil
}

// Store uploads the given reader as an object with the provided key and content type.
// Returns a presigned URL for accessing the object.
func (c *Client) Store(ctx context.Context, key string, reader interface{ Read(p []byte) (n int, err error) }, size int64, contentType string, expiry time.Duration) (*url.URL, error) {
	// Upload object
	_, err := c.minio.PutObject(ctx, c.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.logger.Error("failed to upload object", "bucket", c.bucket, "key", key, "error", err)
		return nil, fmt.Errorf("upload error: %w", err)
	}

	// Generate presigned URL
	presignedURL, err := c.minio.PresignedGetObject(ctx, c.bucket, key, expiry, nil)
	if err != nil {
		c.logger.Error("failed to presign object", "bucket", c.bucket, "key", key, "error", err)
		return nil, fmt.Errorf("presign error: %w", err)
	}
	return presignedURL, nil
}

// PresignGet returns a presigned GET URL for the specified key and expiry.
func (c *Client) PresignGet(ctx context.Context, key string, expiry time.Duration) (*url.URL, error) {
	url, err := c.minio.PresignedGetObject(ctx, c.bucket, key, expiry, nil)
	if err != nil {
		c.logger.Error("failed to presign GET object", "bucket", c.bucket, "key", key, "error", err)
		return nil, fmt.Errorf("presign GET error: %w", err)
	}
	return url, nil
}

// DeleteObject deletes the object with the given key from the bucket.
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	err := c.minio.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		c.logger.Error("failed to delete object", "bucket", c.bucket, "key", key, "error", err)
		return fmt.Errorf("delete error: %w", err)
	}
	return nil
}