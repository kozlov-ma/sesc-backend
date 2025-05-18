// Package s3svc provides S3-compatible object storage functionality.
package s3svc

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/kozlov-ma/sesc-backend/internal/filesvc"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage implements the filesvc.ObjectStorage interface using MinIO.
type Storage struct {
	client     *minio.Client
	bucketName string
}

// NewStorage creates a new S3 storage instance.
func NewStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*Storage, error) {
	ctx := context.Background()
	ctx, rec := event.NewRecord(ctx, "s3svc/new_storage")
	defer rec.Finish()

	rec.Set(
		"endpoint", endpoint,
		"bucket_name", bucketName,
		"use_ssl", useSSL,
	)

	// Initialize MinIO client
	var minioClient *minio.Client
	err := rec.Operation("init_client", func(*event.Record) error {
		var err error
		minioClient, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})
		return err
	})

	if err != nil {
		return nil, err
	}

	// Check if bucket exists
	var exists bool
	err = rec.Operation("check_bucket", func(rec *event.Record) error {
		rec.Set("bucket_name", bucketName)
		var err error
		exists, err = minioClient.BucketExists(ctx, bucketName)
		if err == nil {
			rec.Set("bucket_exists", exists)
		}
		return err
	})

	if err != nil {
		return nil, err
	}

	// Create bucket if it doesn't exist
	if !exists {
		err = rec.Operation("create_bucket", func(rec *event.Record) error {
			rec.Set("bucket_name", bucketName)
			return minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		})

		if err != nil {
			return nil, err
		}
	}

	rec.Set("success", true)

	return &Storage{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

// PutObject uploads an object to S3 storage.
func (s *Storage) PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64) error {
	rec := event.Get(ctx).Sub("s3svc/put_object")
	rec.Set(
		"start_time", time.Now(),
		"bucket_name", s.bucketName,
		"object_key", objectKey,
		"size", size,
	)

	var putResult minio.UploadInfo
	err := rec.Operation("upload", func(rec *event.Record) error {
		var err error
		startTime := time.Now()
		putResult, err = s.client.PutObject(
			ctx,
			s.bucketName,
			objectKey,
			reader,
			size,
			minio.PutObjectOptions{},
		)
		rec.Set("duration_ms", time.Since(startTime).Milliseconds())

		if err == nil {
			rec.Set(
				"etag", putResult.ETag,
				"version_id", putResult.VersionID,
			)
		}

		return err
	})

	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	return nil
}

// RemoveObject deletes an object from S3 storage.
func (s *Storage) RemoveObject(ctx context.Context, objectKey string) error {
	rec := event.Get(ctx).Sub("s3svc/remove_object")
	rec.Set(
		"start_time", time.Now(),
		"bucket_name", s.bucketName,
		"object_key", objectKey,
	)

	err := rec.Operation("delete", func(*event.Record) error {
		return s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	})

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to remove object: %w", err))
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	return nil
}

// GetObjectURL returns a presigned URL for the object.
func (s *Storage) GetObjectURL(
	ctx context.Context,
	objectKey string,
	downloadName string,
	expires time.Duration,
) (string, error) {
	rec := event.Get(ctx).Sub("s3svc/get_object_url")
	rec.Set(
		"start_time", time.Now(),
		"bucket_name", s.bucketName,
		"object_key", objectKey,
		"expires", expires.String(),
	)

	params := make(url.Values)
	params.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadName))

	var presignedURL string
	err := rec.Operation("presign", func(rec *event.Record) error {
		var err error
		url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectKey, expires, params)
		if err == nil {
			presignedURL = url.String()
			rec.Set("url_generated", true)
		}
		return err
	})

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to generate presigned URL: %w", err))
		rec.Set("success", false)
		return "", err
	}

	rec.Set("success", true)
	return presignedURL, nil
}

// Compile-time check that Storage implements the ObjectStorage interface
var _ filesvc.ObjectStorage = (*Storage)(nil)
