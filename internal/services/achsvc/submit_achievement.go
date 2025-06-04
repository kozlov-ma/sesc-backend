package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// SubmitAchievement submits an achievement for review.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *ACS) SubmitAchievement(
	ctx context.Context,
	opt achievement.SubmitOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/submit_achievement")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.OwnerID,
		"achievement_id", opt.AchievementID,
	)

	var updatedAch *ent.Achievement
	err := withTx(ctx, s.client, func(tx *ent.Tx) error {
		var ach *ent.Achievement
		err := rec.Operation("query_achievement", func(opRec *event.Record) error {
			opRec.Sub("params").Set(
				"achievement_id", opt.AchievementID,
				"owner_id", opt.OwnerID,
			)

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

		err = rec.Operation("validate_status", func(opRec *event.Record) error {
			opRec.Set("current_status", ach.Status)
			opRec.Set("required_status", string(achievement.StatusDraft))

			if ach.Status != string(achievement.StatusDraft) {
				return achievement.ErrWrongAchievementStatus
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("update_status", func(opRec *event.Record) error {
			start := time.Now()
			entity, err := tx.Achievement.UpdateOne(ach).
				SetStatus(string(achievement.StatusDepheadReview)).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to update achievement status: %w", err)
			}

			updatedAch = entity
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return updatedAch, err
}
