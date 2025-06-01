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

// GetAchievement retrieves an achievement by ID and owner ID.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
func (s *ACS) GetAchievement(
	ctx context.Context,
	achievementID UUID,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/get_achievement")
	// Group parameters together
	rec.Sub("params").Set("achievement_id", achievementID)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Get achievement with all related data
	var achievementEntity *ent.Achievement
	err := rec.Operation("query_achievement", func(opRec *event.Record) error {
		opRec.Sub("params").Set("achievement_id", achievementID)

		queryStart := time.Now()
		entity, err := s.client.Achievement.Query().
			Where(
				entAchievement.ID(achievementID),
			).
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
		opRec.Set("found", true)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert to domain model
	var result achievement.Achievement
	err = rec.Operation("convert_to_domain", func(opRec *event.Record) error {
		domainModel, err := convertAchievementEntityToDomain(achievementEntity, opRec)
		if err != nil {
			opRec.Add(events.Error, err)
			return err
		}
		result = domainModel
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	rec.Set("achievement", result)
	return result, nil
}
