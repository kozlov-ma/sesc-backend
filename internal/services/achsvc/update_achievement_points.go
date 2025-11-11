package achsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// updateAchievementPoints allows teachers to update achievement points when changes are requested.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in a status that allows point updates.
func (s *ACS) updateAchievementPoints(
	ctx context.Context,
	opt achievement.UpdatePointsOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/update_achievement_points")
	statsRec := event.Root(ctx).Sub("stats")

	rec.Sub("params").Set(
		"achievement_id", opt.AchievementID,
		"owner_id", opt.OwnerID,
		"points", opt.Points,
	)

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

		err = rec.Operation("validate_update", func(rec *event.Record) error {
			rec.Sub("params").Set(
				"current_status", ach.Status,
				"points_limit", ach.Edges.Template.PointsLimit,
			)

			currentStatus := achievement.Status(ach.Status)
			if currentStatus != achievement.StatusDepheadRequestedChanges &&
				currentStatus != achievement.StatusInspectorRequestedChanges {
				return achievement.ErrWrongAchievementStatus
			}

			if opt.Points < 0 {
				return errors.New("points cannot be negative")
			}

			pointsLimit := ach.Edges.Template.PointsLimit
			if opt.Points > pointsLimit {
				return achievement.ErrPointsLimitExceeded
			}

			return nil
		})
		if err != nil {
			return err
		}
		// Create a review entry for the teacher's update if comment provided
		if opt.Comment != "" {
			err = rec.Operation("create_update_review", func(_ *event.Record) error {
				reviewID, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("failed to generate review ID: %w", err)
				}

				start := time.Now()
				_, err = tx.AchievementReview.Create().
					SetID(reviewID).
					SetAchievementID(opt.AchievementID).
					SetReviewerID(opt.OwnerID). // Teacher is updating their own achievement
					SetPointsAssigned(opt.Points).
					SetComment(opt.Comment).
					Save(ctx)
				statsRec.Add(events.PostgresQueries, 1)
				statsRec.Add(events.PostgresTime, time.Since(start))

				if err != nil {
					return fmt.Errorf("failed to create update review: %w", err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

		err = rec.Operation("update_achievement", func(_ *event.Record) error {
			currentStatus := achievement.Status(ach.Status)

			var newStatus achievement.Status
			switch currentStatus {
			case achievement.StatusDepheadRequestedChanges:
				newStatus = achievement.StatusDepheadReview
			case achievement.StatusInspectorRequestedChanges:
				newStatus = achievement.StatusInspectorReview
			default:
				return fmt.Errorf("unexpected status: %s", currentStatus)
			}

			start := time.Now()
			_, err := tx.Achievement.UpdateOne(ach).
				SetStatus(string(newStatus)).
				SetPoints(opt.Points).
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
