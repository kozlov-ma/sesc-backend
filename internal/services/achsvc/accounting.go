package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

func (s *ACS) MarkAchievementsAsAccounted(ctx context.Context, achievementIDs []uuid.UUID) error {
	rec := event.Get(ctx).Sub("sesc/mark_achievements_as_accounted")

	if len(achievementIDs) == 0 {
		return nil
	}

	statsRec := event.Get(ctx).Sub("stats")
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", 1)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	err := rec.Operation("update_achievements_status", func(opRec *event.Record) error {
		opRec.Set("achievement_count", len(achievementIDs))

		queryStart := time.Now()
		count, err := s.client.Achievement.Update().
			Where(
				entAchievement.IDIn(achievementIDs...),
				entAchievement.StatusEQ(string(achievement.StatusDone)),
			).
			SetStatus(string(achievement.StatusAccounted)).
			Save(ctx)
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to update achievement statuses: %w", err))
			return err
		}

		opRec.Set("updated_count", count)
		return nil
	})

	if err != nil {
		return err
	}

	rec.Set("success", true)
	rec.Set("achievement_ids", achievementIDs)
	return nil
}

// markAllDoneAchievementsAsAccounted marks all achievements with "done" status as "accounted"
func (s *ACS) markAllDoneAchievementsAsAccounted(ctx context.Context) (int, error) {
	rec := event.Get(ctx).Sub("sesc/mark_all_done_achievements_as_accounted")

	statsRec := event.Get(ctx).Sub("stats")
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", 1)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	var count int
	err := rec.Operation("update_all_done_achievements_status", func(opRec *event.Record) error {
		queryStart := time.Now()
		updatedCount, err := s.client.Achievement.Update().
			Where(entAchievement.StatusEQ(string(achievement.StatusDone))).
			SetStatus(string(achievement.StatusAccounted)).
			Save(ctx)

		statsRec.Add("total_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to mark all done achievements as accounted: %w", err))
			return err
		}

		count = updatedCount
		opRec.Set("updated_count", count)
		return nil
	})

	if err != nil {
		return 0, err
	}

	rec.Set("success", true)
	rec.Set("updated_count", count)
	return count, nil
}
