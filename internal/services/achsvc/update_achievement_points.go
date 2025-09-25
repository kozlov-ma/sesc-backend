package achsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// UpdateAchievementPoints updates the points for an achievement when requested by reviewer.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in the correct status for points update.
func (s *ACS) UpdateAchievementPoints(
	ctx context.Context,
	opt achievement.UpdatePointsOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/update_achievement_points")
	rec.Sub("params").Set(
		"achievement_id", opt.AchievementID,
		"owner_id", opt.OwnerID,
		"points", opt.Points,
	)

	statsRec := event.Root(ctx).Sub("stats")

	var updatedAch *ent.Achievement
	err := txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		var ach *ent.Achievement
		err := rec.Operation("query_achievement", func(_ *event.Record) error {
			start := time.Now()
			entity, err := tx.Achievement.Query().
				Where(
					entAchievement.ID(opt.AchievementID),
					entAchievement.OwnerID(opt.OwnerID),
				).
				WithTemplate().
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

		err = rec.Operation("validate_status", func(rec *event.Record) error {
			rec.Sub("params").Set(
				"current_status", ach.Status,
			)

			currentStatus := achievement.Status(ach.Status)
			// Allow points update only when reviewer requested it
			if currentStatus != achievement.StatusDepheadPointsChange &&
				currentStatus != achievement.StatusInspectorPointsChange {
				return achievement.ErrWrongAchievementStatus
			}

			// Validate points against template limit
			if opt.Points > ach.Edges.Template.PointsLimit {
				return achievement.ErrPointsLimitExceeded
			}
			if opt.Points < 0 {
				return errors.New("points cannot be negative")
			}

			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("update_achievement", func(_ *event.Record) error {
			start := time.Now()
			_, err := tx.Achievement.UpdateOne(ach).
				SetPoints(opt.Points).
				SetStatus(string(achievement.StatusDraft)).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to update achievement: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("query_updated", func(_ *event.Record) error {
			var err error
			updatedAch, err = tx.Achievement.Query().
				Where(entAchievement.ID(ach.ID)).
				WithDocuments().
				WithOwner().
				WithReviews().
				WithTemplate().
				Only(ctx)
			if err != nil {
				return fmt.Errorf("failed to query updated achievement: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return updatedAch, err
}
