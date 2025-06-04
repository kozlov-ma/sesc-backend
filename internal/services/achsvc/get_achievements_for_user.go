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

// GetUserAchievements retrieves all achievements for the current user with pagination.
// Results are ordered based on the asking user's role and review responsibilities.
func (s *ACS) GetUserAchievements(
	ctx context.Context,
	userID UUID,
	whosAsking UUID,
	offset, limit int,
) (ent.Achievements, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_user_achievements")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", userID,
		"whos_asking", whosAsking,
		"offset", offset,
		"limit", limit,
	)

	var totalAchievements int
	var achievementEntities []*ent.Achievement

	err := rec.Operation("count_achievements", func(opRec *event.Record) error {
		// Get asking user for role-based filtering
		start := time.Now()
		askingUser, err := s.client.User.Query().
			Where(user.ID(whosAsking)).
			WithDepartment().
			Only(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to get asking user: %w", err)
		}

		// Apply role-based filtering for count
		query := s.client.Achievement.Query().
			Where(entAchievement.OwnerID(userID))

		// Apply role-based filters
		roleFilter := s.buildRoleBasedFilters(askingUser)
		roleFilter(query)

		start = time.Now()
		count, err := query.Count(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to count achievements: %w", err)
		}

		totalAchievements = count
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	err = rec.Operation("query_achievements", func(opRec *event.Record) error {
		// Get asking user for role-based filtering
		start := time.Now()
		askingUser, err := s.client.User.Query().
			Where(user.ID(whosAsking)).
			WithDepartment().
			Only(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to get asking user: %w", err)
		}

		// Build query with role-based filtering
		query := s.client.Achievement.Query().
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
			})

		// Apply role-based filters
		roleFilter := s.buildRoleBasedFilters(askingUser)
		roleFilter(query)

		start = time.Now()
		entities, err := query.
			Order(ent.Desc(entAchievement.FieldCreatedAt)).
			Offset(offset).
			Limit(limit).
			All(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to query achievements: %w", err)
		}

		achievementEntities = entities
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return achievementEntities, totalAchievements, nil
}
