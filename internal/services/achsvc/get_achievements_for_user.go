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

func (s *ACS) GetAchievementsForUser(
	ctx context.Context,
	userID UUID,
	offset, limit int,
) ([]achievement.Achievement, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_achievements_for_user")
	rec.Add("user_id", userID)
	rec.Add("offset", offset)
	rec.Add("limit", limit)

	// Count total achievements for the user
	total, err := s.client.Achievement.Query().
		Where(entAchievement.OwnerID(userID)).
		Count(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to count achievements: %w", err))
		return nil, 0, err
	}

	// Get achievements with pagination
	achievementEntities, err := s.client.Achievement.Query().
		Where(entAchievement.OwnerID(userID)).
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
		Order(ent.Desc(entAchievement.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievements: %w", err))
		return nil, 0, err
	}

	// Convert to domain models
	results := make([]achievement.Achievement, 0, len(achievementEntities))
	for _, entity := range achievementEntities {
		ach, err := convertAchievementEntityToDomain(entity, rec)
		if err != nil {
			rec.Add(events.Error, err)
			return nil, 0, err
		}
		results = append(results, ach)
	}

	rec.Add("achievements_count", len(results))
	rec.Add("total_count", total)
	return results, total, nil
}

// GetUserAchievements retrieves all achievements for the current user with pagination.
func (s *ACS) GetUserAchievements(
	ctx context.Context,
	userID UUID,
	offset, limit int,
) ([]achievement.Achievement, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_user_achievements")

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", userID,
		"offset", offset,
		"limit", limit,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Count total achievements for the user
	var totalAchievements int
	err := rec.Operation("count_achievements", func(opRec *event.Record) error {
		opRec.Sub("params").Set("user_id", userID)

		queryStart := time.Now()
		count, err := s.client.Achievement.Query().
			Where(entAchievement.OwnerID(userID)).
			Count(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to count achievements: %w", err))
			return err
		}

		totalAchievements = count
		opRec.Set("total_count", totalAchievements)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Get achievements with pagination
	var achievementEntities []*ent.Achievement
	err = rec.Operation("query_achievements", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"user_id", userID,
			"offset", offset,
			"limit", limit,
		)

		queryStart := time.Now()
		entities, err := s.client.Achievement.Query().
			Where(entAchievement.OwnerID(userID)).
			WithTemplate(func(q *ent.AchievementTemplateQuery) {
				q.WithGroup()
			}).
			WithOwner(func(q *ent.UserQuery) {
				q.WithDepartment()
			}).
			WithDocuments(func(q *ent.AchievementDocumentQuery) {
				q.WithFile()
			}).
			WithReviews(func(q *ent.AchievementReviewQuery) {
				q.WithReviewer()
			}).
			Order(ent.Desc(entAchievement.FieldID)).
			Offset(offset).
			Limit(limit).
			All(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query achievements: %w", err))
			return err
		}

		achievementEntities = entities
		opRec.Set("achievements_count", len(achievementEntities))
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Convert to domain models
	result := make([]achievement.Achievement, 0, len(achievementEntities))
	err = rec.Operation("convert_achievements", func(opRec *event.Record) error {
		opRec.Set("entities_count", len(achievementEntities))

		for i, entity := range achievementEntities {
			achRec := opRec.Sub(fmt.Sprintf("achievement_%d", i))
			achRec.Set("id", entity.ID)

			ach, err := convertAchievementEntityToDomain(entity, achRec)
			if err != nil {
				achRec.Add(events.Error, err)
				return err
			}
			result = append(result, ach)
		}

		opRec.Set("converted_count", len(result))
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"achievements_count", len(result),
		"total_achievements", totalAchievements,
		"user_id", userID,
	)

	return result, totalAchievements, nil
}
