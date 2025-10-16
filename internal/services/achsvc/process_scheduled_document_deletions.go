package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
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

	// Get documents ready for deletion (scheduled_deletion_at + delay <= now)
	cutoffTime := time.Now().Add(-delay)
	rec.Set("cutoff_time", cutoffTime)

	start := time.Now()
	documents, err := s.client.AchievementDocument.Query().
		Where(
			achievementdocument.Status(achievement.DocumentStatusScheduled),
			achievementdocument.ScheduledDeletionAtLTE(cutoffTime),
		).
		WithFile().
		All(ctx)
	statsRec.Add(events.PostgresQueries, 1)
	statsRec.Add(events.PostgresTime, time.Since(start))

	if err != nil {
		rec.Add(events.Error, err)
		return err
	}

	rec.Set("documents_to_process", len(documents))

	processedCount := 0
	for _, doc := range documents {
		// Get the associated file
		file := doc.Edges.File
		if file == nil {
			rec.Add(events.Error, fmt.Errorf("document %s has no associated file", doc.ID.String()))
			continue
		}

		// Delete from storage if the file has an S3 object key
		if file.S3ObjectKey != "" {
			storageErr := storage.RemoveObject(ctx, file.S3ObjectKey)
			if storageErr != nil {
				rec.Add(events.Error, fmt.Errorf("failed to delete object %s: %w", file.S3ObjectKey, storageErr))
				continue
			}
		}

		// Mark document as deleted and clear deletion schedule
		start := time.Now()
		_, err := s.client.AchievementDocument.UpdateOneID(doc.ID).
			SetStatus(achievement.DocumentStatusDeleted).
			ClearScheduledDeletionAt().
			Save(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to update document %s: %w", doc.ID.String(), err))
			continue
		}

		// Delete the file record from database
		start = time.Now()
		err = s.client.File.DeleteOneID(file.ID).Exec(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to delete file record %s: %w", file.ID.String(), err))
			continue
		}

		processedCount++
	}

	rec.Set("processed_count", processedCount)
	rec.Set("success", true)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return nil
}
