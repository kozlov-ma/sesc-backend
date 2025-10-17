package achsvc

import (
	"context"
	"database/sql"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/file"
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

func (s *ACS) GetDocumentStats(ctx context.Context, isCommon bool) (*DocumentStats, error) {
	rec := event.Get(ctx).Sub("achsvc/get_document_stats")

	rec.Sub("params").Set(
		"start_time", time.Now(),
		"is_common", isCommon,
	)

	var stats DocumentStats

	err := txwrapper.WithTx(ctx, s.client, sql.LevelRepeatableRead, rec, func(tx *ent.Tx) error {
		var txErr error

		baseQuery := tx.AchievementDocument.Query().Where(achievementdocument.HasFile())
		if isCommon {
			baseQuery = baseQuery.Where(achievementdocument.HasFileWith(file.OwnerIDIsNil()))
		} else {
			baseQuery = baseQuery.Where(achievementdocument.HasFileWith(file.OwnerIDNotNil()))
		}

		notScheduledDocs, txErr := baseQuery.Clone().
			Where(achievementdocument.Status(achievement.DocumentStatusActive)).
			All(ctx)
		if txErr != nil {
			return txErr
		}
		stats.NotScheduled = len(notScheduledDocs)

		scheduledDocs, txErr := baseQuery.Clone().
			Where(achievementdocument.Status(achievement.DocumentStatusScheduled)).
			All(ctx)
		if txErr != nil {
			return txErr
		}
		stats.ScheduledForDeletion = len(scheduledDocs)

		deletedDocs, txErr := baseQuery.Clone().
			Where(achievementdocument.Status(achievement.DocumentStatusDeleted)).
			All(ctx)
		if txErr != nil {
			return txErr
		}
		stats.DeletedDocuments = len(deletedDocs)

		stats.TotalDocuments = stats.NotScheduled + stats.ScheduledForDeletion

		readyDocs, txErr := baseQuery.Clone().
			Where(
				achievementdocument.Status(achievement.DocumentStatusScheduled),
				achievementdocument.ScheduledDeletionAtLTE(time.Now()),
			).
			All(ctx)
		if txErr != nil {
			return txErr
		}

		stats.ReadyForDeletion = len(readyDocs)

		return nil
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	rec.Set("success", true)
	rec.Set("stats", stats)
	rec.Set("total_duration_ms", time.Since(rec.Sub("params").Value("start_time").(time.Time)).Milliseconds())

	return &stats, nil
}
