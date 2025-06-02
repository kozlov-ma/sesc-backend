package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
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

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.OwnerID,
		"achievement_id", opt.AchievementID,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Get achievement to check status
	var achievementEntity *ent.Achievement
	err := rec.Operation("query_achievement", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"achievement_id", opt.AchievementID,
			"owner_id", opt.OwnerID,
		)

		queryStart := time.Now()
		entity, err := s.client.Achievement.Query().
			Where(
				entAchievement.ID(opt.AchievementID),
				entAchievement.OwnerID(opt.OwnerID),
			).
			Only(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if ent.IsNotFound(err) {
			opRec.Add(events.Error, "achievement not found")
			return achievement.ErrAchievementNotFound
		}
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
			return err
		}

		achievementEntity = entity
		opRec.Set("status", achievementEntity.Status)
		return nil
	})
	if err != nil {
		return err
	}

	// Validate achievement status
	err = rec.Operation("validate_status", func(opRec *event.Record) error {
		opRec.Set("current_status", achievementEntity.Status)
		opRec.Set("required_status", string(achievement.StatusDraft))

		// Check if achievement is in draft status
		if achievementEntity.Status != string(achievement.StatusDraft) {
			opRec.Add(events.Error, "achievement is not in draft status")
			return achievement.ErrWrongAchievementStatus
		}

		opRec.Set("valid", true)
		return nil
	})
	if err != nil {
		return err
	}

	// Delete achievement and its documents
	var tx *ent.Tx
	err = rec.Operation("delete_achievement", func(opRec *event.Record) error {
		// Start a transaction
		queryStart := time.Now()
		txn, err := s.client.Tx(ctx)
		queryCount++
		opRec.Add("tx_start_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to start transaction: %w", err))
			return err
		}
		tx = txn

		// Delete achievement documents
		queryStart = time.Now()
		result, err := tx.AchievementDocument.Delete().
			Where(achievementdocument.AchievementID(opt.AchievementID)).
			Exec(ctx)
		queryCount++
		opRec.Add("delete_documents_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to delete achievement documents: %w", err))
			return rollback(tx, err)
		}
		opRec.Set("documents_deleted", result)

		// Delete achievement
		queryStart = time.Now()
		result, err = tx.Achievement.Delete().
			Where(
				entAchievement.ID(opt.AchievementID),
				entAchievement.OwnerID(opt.OwnerID),
			).
			Exec(ctx)
		queryCount++
		opRec.Add("delete_achievement_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to delete achievement: %w", err))
			return rollback(tx, err)
		}
		opRec.Set("achievements_deleted", result)

		return nil
	})
	if err != nil {
		return err
	}

	// Commit transaction
	err = rec.Operation("commit_transaction", func(opRec *event.Record) error {
		queryStart := time.Now()
		err := tx.Commit()
		queryCount++
		opRec.Add("commit_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to commit transaction: %w", err))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"success", true,
		"achievement_id", opt.AchievementID,
	)

	rec.Add("success", true)
	return nil
}
