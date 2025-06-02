package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// GetGroupedAchievements retrieves all achievements grouped by user with pagination.
func (s *ACS) GetGroupedAchievements(
	ctx context.Context,
	offset, limit int,
) (map[UUID][]achievement.Achievement, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_grouped_achievements")

	// Group parameters together
	rec.Sub("params").Set(
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

	// Count total unique users with achievements
	var totalUsers int
	err := rec.Operation("count_users", func(opRec *event.Record) error {
		queryStart := time.Now()
		count, err := s.client.User.Query().
			Where(user.HasAchievements()).
			Count(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to count users with achievements: %w", err))
			return err
		}

		totalUsers = count
		opRec.Set("total_users", totalUsers)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Get users with achievements with pagination
	var users []*ent.User
	err = rec.Operation("query_users", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"offset", offset,
			"limit", limit,
		)

		queryStart := time.Now()
		userList, err := s.client.User.Query().
			Where(user.HasAchievements()).
			WithDepartment().
			Offset(offset).
			Limit(limit).
			All(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query users with achievements: %w", err))
			return err
		}

		users = userList
		opRec.Set("users_count", len(users))
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Create a map to store achievements grouped by user ID
	groupedAchievements := make(map[UUID][]achievement.Achievement)

	// For each user, get their achievements (excluding drafts)
	err = rec.Operation("get_user_achievements", func(opRec *event.Record) error {
		opRec.Set("users_count", len(users))

		for i, usr := range users {
			userRec := opRec.Sub(fmt.Sprintf("user_%d", i))
			userRec.Set("user_id", usr.ID)

			// Query achievements for this user
			queryStart := time.Now()
			achievementEntities, err := s.client.Achievement.Query().
				Where(
					entAchievement.OwnerID(usr.ID),
					entAchievement.StatusNEQ(string(achievement.StatusDraft)), // Exclude drafts
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
				Order(ent.Desc(entAchievement.FieldID)).
				All(ctx)
			queryCount++
			userRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

			if err != nil {
				userRec.Add(events.Error, fmt.Errorf("failed to query achievements for user %s: %w", usr.ID, err))
				return err
			}
			userRec.Set("achievements_count", len(achievementEntities))

			// Convert to domain models
			userAchievements := make([]achievement.Achievement, 0, len(achievementEntities))
			for j, entity := range achievementEntities {
				achRec := userRec.Sub(fmt.Sprintf("achievement_%d", j))
				achRec.Set("id", entity.ID)

				ach, err := convertAchievementEntityToDomain(entity, achRec)
				if err != nil {
					achRec.Add(events.Error, err)
					return err
				}
				userAchievements = append(userAchievements, ach)
			}

			// Add to the grouped map
			groupedAchievements[usr.ID] = userAchievements
			userRec.Set("converted_count", len(userAchievements))
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"users_count", len(users),
		"total_users", totalUsers,
		"total_groups", len(groupedAchievements),
	)

	return groupedAchievements, totalUsers, nil
}
