package achsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// ScheduleDocumentDeletionAll schedules all documents for deletion after a specified delay
func (s *ACS) ScheduleDocumentDeletionAll(ctx context.Context, delay time.Duration) error {
	rec := event.Get(ctx).Sub("achsvc/schedule_document_deletion_all")
	statsRec := event.Root(ctx).Sub("stats")

	rec.Sub("params").Set(
		"delay", delay,
		"start_time", time.Now(),
	)

	deletionTime := time.Now().Add(delay)
	var scheduledCount int

	err := txwrapper.WithTx(ctx, s.client, sql.LevelSerializable, rec, func(tx *ent.Tx) error {
		start := time.Now()
		documents, err := tx.AchievementDocument.Query().
			Where(
				achievementdocument.Status(achievement.DocumentStatusActive),
			).
			All(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			rec.Add(events.Error, err)
			return err
		}

		rec.Set("documents_count", len(documents))

		for i, doc := range documents {
			docRec := rec.Sub(fmt.Sprintf("document_%d", i))
			docRec.Set("document_id", doc.ID.String())

			docRec.Set("action", "schedule_deletion")
			start := time.Now()
			_, err := tx.AchievementDocument.UpdateOneID(doc.ID).
				SetStatus(achievement.DocumentStatusScheduled).
				SetScheduledDeletionAt(deletionTime).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				docRec.Add(events.Error, fmt.Errorf("failed to schedule document %s: %w", doc.ID.String(), err))
				return err
			}
			docRec.Set("success", true)
			scheduledCount++
		}

		return nil
	})

	if err != nil {
		return err
	}

	rec.Set("scheduled_count", scheduledCount)
	rec.Set("scheduled_deletion_at", deletionTime)
	rec.Set("success", true)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return nil
}
