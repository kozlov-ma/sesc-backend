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

// ScheduleDeletion schedules the deletion of a document (admin operation).
// Returns achievement.ErrDocumentNotFound if the document does not exist.
func (s *ACS) ScheduleDeletion(
	ctx context.Context,
	opt achievement.ScheduleDocumentDeletionOptions,
) error {
	rec := event.Get(ctx).Sub("sesc/schedule_deletion")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"document_id", opt.DocumentID,
	)

	err := txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		var doc *ent.AchievementDocument
		err := rec.Operation("query_document", func(_ *event.Record) error {
			start := time.Now()
			document, err := tx.AchievementDocument.Query().
				Where(
					achievementdocument.ID(opt.DocumentID),
				).
				Only(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return achievement.ErrDocumentNotFound
			}
			if err != nil {
				return fmt.Errorf("failed to query document: %w", err)
			}

			doc = document
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("schedule_document_deletion", func(_ *event.Record) error {
			start := time.Now()
			scheduledAt := time.Now()
			_, err := tx.AchievementDocument.UpdateOne(doc).
				SetStatus(achievement.DocumentStatusScheduled).
				SetScheduledDeletionAt(scheduledAt).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to schedule document deletion: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
