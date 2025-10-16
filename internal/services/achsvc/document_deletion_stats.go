package achsvc

import (
	"context"
	"database/sql"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type DocumentStats struct {
	TotalDocuments       int
	DeletedDocuments     int
	ScheduledForDeletion int
	ReadyForDeletion     int
	NotScheduled         int
}

// GetDocumentStats returns statistics about documents and deletion delay
func (s *ACS) GetDocumentStats(ctx context.Context) (*DocumentStats, error) {
	rec := event.Get(ctx).Sub("achsvc/get_document_stats")

	rec.Sub("params").Set(
		"start_time", time.Now(),
	)

	var stats DocumentStats

	err := txwrapper.WithTx(ctx, s.client, sql.LevelRepeatableRead, rec, func(tx *ent.Tx) error {
		var txErr error
		// Count total documents
		stats.TotalDocuments, txErr = tx.AchievementDocument.Query().Count(ctx)
		if txErr != nil {
			return txErr
		}

		// Count deleted documents
		stats.DeletedDocuments, txErr = tx.AchievementDocument.Query().
			Where(achievementdocument.Status(achievement.DocumentStatusDeleted)).
			Count(ctx)
		if txErr != nil {
			return txErr
		}

		// Count scheduled documents
		stats.ScheduledForDeletion, txErr = tx.AchievementDocument.Query().
			Where(achievementdocument.Status(achievement.DocumentStatusScheduled)).
			Count(ctx)
		if txErr != nil {
			return txErr
		}

		// Count ready for deletion documents
		stats.ReadyForDeletion, txErr = tx.AchievementDocument.Query().
			Where(
				achievementdocument.Status(achievement.DocumentStatusScheduled),
				achievementdocument.ScheduledDeletionAtLTE(time.Now()),
			).
			Count(ctx)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	stats.NotScheduled = stats.TotalDocuments - stats.DeletedDocuments - stats.ScheduledForDeletion

	rec.Set("success", true)
	rec.Set("stats", stats)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return &stats, nil
}
