// Package filesvc provides services for managing files.
package filesvc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/file"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/predicate"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	UUID     = uuid.UUID
	FileOpts = sesc.FileCreateOptions
)

// FileService provides methods for managing files.
type FileService struct {
	client     *ent.Client
	storage    ObjectStorage
	bucketName string
}

// New creates a new FileService instance.
func New(client *ent.Client, storage ObjectStorage, bucketName string) *FileService {
	return &FileService{
		client:     client,
		storage:    storage,
		bucketName: bucketName,
	}
}

// withTransaction executes the given operation within a transaction and handles commit/rollback.
func (s *FileService) withTransaction(ctx context.Context, rec *event.Record, operation func(tx *ent.Tx) error) error {
	var tx *ent.Tx

	err := rec.Operation("start_transaction", func(*event.Record) error {
		var err error
		tx, err = s.client.Tx(ctx)
		return err
	})

	if err != nil {
		return err
	}

	// Execute the operation within the transaction
	if err := operation(tx); err != nil {
		return rollback(tx, err)
	}

	// Commit the transaction
	return rec.Operation("commit_transaction", func(*event.Record) error {
		return tx.Commit()
	})
}

// getStatsRecord retrieves the stats record from the root event record
func getStatsRecord(ctx context.Context) *event.Record {
	return event.Root(ctx).Sub("stats")
}

// recordDBOperation executes a database operation and records stats and timing
func recordDBOperation(ctx context.Context, rec *event.Record, name string, operation func() error) error {
	statsRec := getStatsRecord(ctx)

	return rec.Operation(name, func(*event.Record) error {
		startTime := time.Now()
		statsRec.Add(events.PostgresQueries, 1)

		err := operation()

		statsRec.Add(events.PostgresTime, time.Since(startTime))
		return err
	})
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %w", err, rerr)
	}
	return err
}

const objectKeyLengthBytes = 16

// generateSecureObjectKey generates a cryptographically secure random string for S3 object key.
func (s *FileService) generateSecureObjectKey() (string, error) {
	b := make([]byte, objectKeyLengthBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Create creates a new file entry and stores the file in the storage.
func (s *FileService) Create(ctx context.Context, reader io.Reader, opts FileOpts) (*ent.File, error) {
	rec := event.Get(ctx).Sub("file/create")

	rec.Sub("params").Set(
		"name", opts.FileName,
		"size", opts.FileSize,
		"owner_id", opts.OwnerID,
		"start_time", time.Now(),
	)

	// Validate input
	var id UUID
	var objectKey string

	err := rec.Operation("validate_input", func(*event.Record) error {
		return opts.Validate()
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	// Generate UUID
	err = rec.Operation("generate_uuid", func(rec *event.Record) error {
		var err error
		id, err = uuid.NewV7()
		if err == nil {
			rec.Set("uuid", id.String())
		}
		return err
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	// Generate object key
	err = rec.Operation("generate_object_key", func(rec *event.Record) error {
		var err error
		objectKey, err = s.generateSecureObjectKey()
		if err == nil {
			rec.Set("object_key", objectKey)
		}
		return err
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	var f *ent.File

	err = s.withTransaction(ctx, rec, func(tx *ent.Tx) error {
		dbErr := recordDBOperation(ctx, rec, "db_create_file", func() error {
			f, err = tx.File.Create().
				SetID(id).
				SetS3ObjectKey(objectKey).
				SetName(opts.FileName).
				SetSize(opts.FileSize).
				SetNillableOwnerID(opts.OwnerID).
				Save(ctx)

			if err == nil {
				rec.Sub("db_create_file").Set("file_id", f.ID.String())
			}

			return err
		})

		if dbErr != nil {
			rec.Add(events.Error, dbErr)
			return dbErr
		}

		uploadErr := rec.Operation("storage_upload", func(rec *event.Record) error {
			rec.Set(
				"object_key", objectKey,
				"file_size", opts.FileSize,
			)
			return s.storage.PutObject(ctx, objectKey, reader, int64(opts.FileSize))
		})

		if uploadErr != nil {
			rec.Add(events.Error, uploadErr)
			return uploadErr
		}

		return nil
	})

	if err != nil {
		_ = s.storage.RemoveObject(ctx, objectKey)
		return nil, err
	}

	rec.Set("success", true)
	rec.Set("file_id", f.ID.String())
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return f, nil
}

// Delete deletes a file by ID.
func (s *FileService) Delete(ctx context.Context, id UUID) error {
	rec := event.Get(ctx).Sub("file/delete")

	rec.Sub("params").Set(
		"id", id.String(),
		"start_time", time.Now(),
	)

	// Get file to obtain object key
	f, err := s.getFile(ctx, rec, id)
	if err != nil {
		rec.Add(events.Error, err)
		return err
	}

	objectKey := f.S3ObjectKey

	// Execute database and storage operations in a transaction
	err = s.withTransaction(ctx, rec, func(tx *ent.Tx) error {
		// Delete from database
		dbErr := recordDBOperation(ctx, rec, "db_delete_file", func() error {
			_, err := tx.File.Delete().Where(file.ID(id)).Exec(ctx)
			return err
		})

		if dbErr != nil {
			rec.Add(events.Error, dbErr)
			return dbErr
		}

		// Delete from storage
		storageErr := rec.Operation("storage_delete", func(rec *event.Record) error {
			rec.Set("object_key", objectKey)
			return s.storage.RemoveObject(ctx, objectKey)
		})

		if storageErr != nil {
			rec.Add(events.Error, storageErr)
			return storageErr
		}

		return nil
	})

	if err != nil {
		return err
	}

	rec.Set("success", true)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return nil
}

// buildFilePredicates builds the query predicates for file search
func buildFilePredicates(opts sesc.FileSearchOptions, rec *event.Record) []predicate.File {
	buildRec := rec.Sub("build_predicates")
	buildRec.Set("start_time", time.Now())

	predicates := []predicate.File{}

	// Filter by name if provided
	if opts.Name != "" {
		predicates = append(predicates, file.NameContainsFold(opts.Name))
		buildRec.Set("name_filter", true)
	}

	// Filter by owner or common files
	if opts.OwnerID != nil {
		predicates = append(predicates, file.OwnerID(*opts.OwnerID))
		buildRec.Set("owner_filter", true)
		buildRec.Set("owner_id", opts.OwnerID.String())
	} else if opts.Common {
		predicates = append(predicates, file.OwnerIDIsNil())
		buildRec.Set("common_filter", true)
	}

	buildRec.Set(
		"predicates_count", len(predicates),
		"end_time", time.Now(),
	)

	return predicates
}

// Search returns a paginated list of files filtered by the given options.
func (s *FileService) Search(ctx context.Context, opts sesc.FileSearchOptions) (ent.Files, int, error) {
	rec := event.Get(ctx).Sub("file/search")

	rec.Sub("params").Set(
		"name", opts.Name,
		"owner_id", opts.OwnerID,
		"common", opts.Common,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"start_time", time.Now(),
	)

	// Set default limit if not provided
	if opts.Limit <= 0 {
		opts.Limit = 50 // Default limit
	}

	// Build query predicates
	predicates := buildFilePredicates(opts, rec)

	// Count total matching files
	var totalCount int
	err := recordDBOperation(ctx, rec, "count_total", func() error {
		var err error
		totalCount, err = s.client.File.Query().
			Where(predicates...).
			Count(ctx)

		if err == nil {
			rec.Sub("count_total").Set("total_count", totalCount)
		}

		return err
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, 0, err
	}

	// Get paginated results
	var files ent.Files
	err = recordDBOperation(ctx, rec, "query_files", func() error {
		var err error
		files, err = s.client.File.Query().
			Where(predicates...).
			Order(ent.Desc(file.FieldID)).
			Offset(opts.Offset).
			Limit(opts.Limit).
			All(ctx)

		if err == nil {
			rec.Sub("query_files").Set(
				"result_count", len(files),
			)
		}

		return err
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, 0, err
	}

	rec.Set("success", true)
	rec.Set("result_count", len(files))
	rec.Set("total_count", totalCount)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return files, totalCount, nil
}

// ByID returns a file by ID.
func (s *FileService) ByID(ctx context.Context, id UUID) (*ent.File, error) {
	rec := event.Get(ctx).Sub("file/by_id")

	rec.Sub("params").Set(
		"id", id.String(),
		"start_time", time.Now(),
	)

	// Get file from database
	f, err := s.getFile(ctx, rec, id)
	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	rec.Set("success", true)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return f, nil
}

// getFile is a helper method to retrieve a file by ID with proper event recording
func (s *FileService) getFile(ctx context.Context, rec *event.Record, id UUID) (*ent.File, error) {
	var file *ent.File

	err := recordDBOperation(ctx, rec, "get_file", func() error {
		var err error
		file, err = s.client.File.Get(ctx, id)

		if ent.IsNotFound(err) {
			rec.Sub("get_file").Set("found", false)
			return sesc.ErrFileNotFound
		}

		if err == nil {
			rec.Sub("get_file").Set(
				"found", true,
				"object_key", file.S3ObjectKey,
			)
		}

		return err
	})

	return file, err
}
