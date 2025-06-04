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
)

// GetAchievement retrieves an achievement by ID.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
func (s *ACS) GetAchievement(
	ctx context.Context,
	achievementID UUID,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/get_achievement")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set("achievement_id", achievementID)

	var ach *ent.Achievement
	err := rec.Operation("query_achievement", func(opRec *event.Record) error {
		opRec.Sub("params").Set("achievement_id", achievementID)

		start := time.Now()
		entity, err := s.client.Achievement.Query().
			Where(entAchievement.ID(achievementID)).
			WithTemplate().
			WithOwner(func(q *ent.UserQuery) {
				q.WithDepartment()
			}).
			WithDocuments(func(q *ent.AchievementDocumentQuery) {
				q.WithFile()
			}).
			WithReviews(func(q *ent.AchievementReviewQuery) {
				q.WithReviewer()
			}).
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

	return ach, err
}
