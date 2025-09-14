package achsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/internal/services/txhelper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// DeleteAchievement deletes an achievement.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *ACS) DeleteAchievement(
	ctx context.Context,
	opt achievement.DeleteOptions,
) error {
	rec := event.Get(ctx).Sub("achsvc/delete_achievement")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.OwnerID,
		"achievement_id", opt.AchievementID,
	)

	err := txhelper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		var ach *ent.Achievement
		err := rec.Operation("query_achievement", func(_ *event.Record) error {
			start := time.Now()
			entity, err := tx.Achievement.Query().
				Where(
					entAchievement.ID(opt.AchievementID),
					entAchievement.OwnerID(opt.OwnerID),
				).
				Only(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return achievement.ErrAchievementNotFound
			}
			if err != nil {
				return fmt.Errorf("failed to query achievement: %w", err)
			}

			ach = entity
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("validate_status", func(_ *event.Record) error {
			if ach.Status != string(achievement.StatusDraft) {
				return achievement.ErrWrongAchievementStatus
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("delete_documents", func(_ *event.Record) error {
			start := time.Now()
			_, err := tx.AchievementDocument.Delete().
				Where(achievementdocument.AchievementID(opt.AchievementID)).
				Exec(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to delete achievement documents: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("delete_achievement", func(_ *event.Record) error {
			start := time.Now()
			_, err := tx.Achievement.Delete().
				Where(
					entAchievement.ID(opt.AchievementID),
					entAchievement.OwnerID(opt.OwnerID),
				).
				Exec(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to delete achievement: %w", err)
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
