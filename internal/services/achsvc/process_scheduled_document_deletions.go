package achsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/file"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// ObjectStorage is an interface for object storage operations
type ObjectStorage interface {
	RemoveObject(ctx context.Context, objectKey string) error
}

// ProcessScheduledDocumentDeletions processes documents that are ready for deletion
func (s *ACS) ProcessScheduledDocumentDeletions(ctx context.Context, storage ObjectStorage, delay time.Duration) error {
	rec := event.Get(ctx).Sub("achsvc/process_scheduled_document_deletions")
	statsRec := event.Root(ctx).Sub("stats")

	rec.Sub("params").Set(
		"start_time", time.Now(),
		"delay", delay,
	)

	cutoffTime := time.Now().Add(-delay)
	rec.Set("cutoff_time", cutoffTime)

	type fileInfo struct {
		FileID      uuid.UUID
		S3ObjectKey string
	}

	var filesToDelete []fileInfo
	var processedCount int

	err := txwrapper.WithTx(ctx, s.client, sql.LevelSerializable, rec, func(tx *ent.Tx) error {
		start := time.Now()
		documents, err := tx.AchievementDocument.Query().
			Where(
				achievementdocument.Status(achievement.DocumentStatusScheduled),
				achievementdocument.ScheduledDeletionAtLTE(cutoffTime),
			).
			WithFile().
			All(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to query documents: %w", err)
		}

		rec.Set("documents_to_process", len(documents))

		for _, doc := range documents {
			if doc.Edges.File != nil && doc.Edges.File.S3ObjectKey != "" {
				filesToDelete = append(filesToDelete, fileInfo{
					FileID:      doc.Edges.File.ID,
					S3ObjectKey: doc.Edges.File.S3ObjectKey,
				})
			}
		}

		start = time.Now()
		updateCount, err := tx.AchievementDocument.Update().
			Where(
				achievementdocument.Status(achievement.DocumentStatusScheduled),
				achievementdocument.ScheduledDeletionAtLTE(cutoffTime),
			).
			SetStatus(achievement.DocumentStatusDeleted).
			ClearScheduledDeletionAt().
			Save(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to update documents: %w", err)
		}

		processedCount = updateCount
		rec.Set("documents_updated", updateCount)

		if len(filesToDelete) > 0 {
			fileIDs := make([]uuid.UUID, len(filesToDelete))
			for i, f := range filesToDelete {
				fileIDs[i] = f.FileID
			}

			start = time.Now()
			deletedCount, err := tx.File.Delete().
				Where(file.IDIn(fileIDs...)).
				Exec(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to delete files: %w", err)
			}

			rec.Set("files_deleted", deletedCount)
		}

		return nil
	})

	if err != nil {
		rec.Add(events.Error, err)
		return err
	}

	storageFailures := 0
	for _, f := range filesToDelete {
		if err := storage.RemoveObject(ctx, f.S3ObjectKey); err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to delete object %s: %w", f.S3ObjectKey, err))
			storageFailures++
		}
	}
	rec.Set("storage_failures", storageFailures)

	rec.Set("processed_count", processedCount)
	rec.Set("success", true)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return nil
}
