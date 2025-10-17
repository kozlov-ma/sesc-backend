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

type ObjectStorage interface {
	RemoveObject(ctx context.Context, objectKey string) error
}

func (s *ACS) ProcessScheduledDocumentDeletions(ctx context.Context, storage ObjectStorage, delay time.Duration) error {
	rec := event.Get(ctx).Sub("achsvc/process_scheduled_document_deletions")
	statsRec := event.Root(ctx).Sub("stats")

	rec.Sub("params").Set(
		"start_time", time.Now(),
		"delay", delay,
	)

	cutoffTime := time.Now().Add(-delay)
	rec.Set("cutoff_time", cutoffTime)

	var filesToDeleteFromS3 []string
	var documentsProcessed int

	err := txwrapper.WithTx(ctx, s.client, sql.LevelSerializable, rec, func(tx *ent.Tx) error {
		start := time.Now()
		scheduledDocs, err := tx.AchievementDocument.Query().
			Where(
				achievementdocument.Status(achievement.DocumentStatusScheduled),
				achievementdocument.ScheduledDeletionAtLTE(cutoffTime),
			).
			WithFile().
			All(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		rec.Set("scheduled", len(scheduledDocs))

		if err != nil {
			rec.Add(events.Error, fmt.Errorf("query error: %w", err))
			return fmt.Errorf("failed to query scheduled documents: %w", err)
		}

		rec.Set("documents_found", len(scheduledDocs))
		rec.Set("cutoff_time_used", cutoffTime.Format(time.RFC3339))

		if len(scheduledDocs) == 0 {
			rec.Set("reason", "no_documents_found")
			return nil
		}

		fileIDsToCheck := make(map[uuid.UUID]string)
		for _, doc := range scheduledDocs {
			if doc.FileID == nil {
				continue
			}

			loadedFile, err := doc.QueryFile().Only(ctx)
			if err != nil {
				continue
			}

			fileIDsToCheck[loadedFile.ID] = loadedFile.S3ObjectKey
		}

		rec.Set("files_to_check", len(fileIDsToCheck))

		start = time.Now()
		updated, err := tx.AchievementDocument.Update().
			Where(
				achievementdocument.Status(achievement.DocumentStatusScheduled),
				achievementdocument.ScheduledDeletionAtLTE(cutoffTime),
			).
			SetStatus(achievement.DocumentStatusDeleted).
			ClearScheduledDeletionAt().
			ClearFileID().
			Save(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			rec.Add(events.Error, fmt.Errorf("update error: %w", err))
			return fmt.Errorf("failed to mark documents as deleted: %w", err)
		}

		documentsProcessed = updated
		rec.Set("documents_marked_deleted", updated)

		if updated == 0 {
			rec.Set("warning", "no_documents_updated")
		}

		orphanedFileIDs := []uuid.UUID{}
		for fileID, s3Key := range fileIDsToCheck {
			start = time.Now()
			remainingRefs, err := tx.AchievementDocument.Query().
				Where(achievementdocument.FileID(fileID)).
				Count(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				rec.Add(events.Error, fmt.Errorf("failed to count refs for file %s: %w", fileID, err))
				continue
			}

			if remainingRefs == 0 {
				orphanedFileIDs = append(orphanedFileIDs, fileID)
				filesToDeleteFromS3 = append(filesToDeleteFromS3, s3Key)
			}
		}

		rec.Set("orphaned_files_found", len(orphanedFileIDs))
		rec.Set("files_to_delete_from_s3", len(filesToDeleteFromS3))

		if len(orphanedFileIDs) > 0 {
			start = time.Now()
			deletedFiles, err := tx.File.Delete().
				Where(file.IDIn(orphanedFileIDs...)).
				Exec(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				rec.Add(events.Error, fmt.Errorf("db delete error: %w", err))
				return fmt.Errorf("failed to delete orphaned files from db: %w", err)
			}

			rec.Set("files_deleted_from_db", deletedFiles)

			if deletedFiles != len(orphanedFileIDs) {
				rec.Set(
					"warning",
					fmt.Sprintf("expected to delete %d files but deleted %d", len(orphanedFileIDs), deletedFiles),
				)
			}
		} else {
			rec.Set("info", "no_orphaned_files_to_delete")
		}

		return nil
	})

	if err != nil {
		rec.Add(events.Error, err)
		rec.Add(events.Error, fmt.Errorf("transaction error: %w", err))
		return err
	}

	rec.Set("transaction_committed", true)

	s3Failures := 0
	for _, s3Key := range filesToDeleteFromS3 {
		if err := storage.RemoveObject(ctx, s3Key); err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to delete s3 object %s: %w", s3Key, err))
			s3Failures++
		}
	}

	rec.Set("documents_processed", documentsProcessed)
	rec.Set("files_deleted_from_s3", len(filesToDeleteFromS3)-s3Failures)
	rec.Set("s3_failures", s3Failures)
	rec.Set("s3_keys_attempted", filesToDeleteFromS3)
	rec.Set("success", true)

	if documentsProcessed == 0 {
		rec.Set("warning", "no_documents_were_processed")
	}
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return nil
}
