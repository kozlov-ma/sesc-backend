package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/pkg/sesc"
)

// GetUsersWithAchievements retrieves users that have achievements with pagination.
func (s *ACS) GetUsersWithAchievements(
	ctx context.Context,
	whosAsking UUID,
	offset, limit int,
) (ent.Users, int, error) {
	rec := event.Get(ctx).Sub("achsvc/get_users_with_achievements")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"whos_asking", whosAsking,
		"offset", offset,
		"limit", limit,
	)

	var totalUsers int
	var users []*ent.User

	err := rec.Operation("count_users", func(opRec *event.Record) error {
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

		// Build count query with role-based filtering
		query := s.client.User.Query().
			Where(user.HasAchievementsWith(func(q *ent.AchievementQuery) {
				// Apply role-based filters to achievements
				roleFilter := s.buildRoleBasedFilters(askingUser)
				roleFilter(q)
			}))

		start = time.Now()
		count, err := query.Count(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to count users with achievements: %w", err)
		}

		totalUsers = count
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	err = rec.Operation("query_users", func(opRec *event.Record) error {
		// Get asking user to determine ordering
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
		query := s.client.User.Query().
			Where(user.HasAchievementsWith(func(q *ent.AchievementQuery) {
				// Apply role-based filters to achievements
				roleFilter := s.buildRoleBasedFilters(askingUser)
				roleFilter(q)
			})).
			WithDepartment().
			WithAchievements(func(q *ent.AchievementQuery) {
				q.WithTemplate(func(tq *ent.AchievementTemplateQuery) {
					tq.WithGroup()
				}).
					WithDocuments(func(dq *ent.AchievementDocumentQuery) {
						dq.WithFile()
					}).
					WithReviews(func(rq *ent.AchievementReviewQuery) {
						rq.WithReviewer()
					})

				// Apply role-based filters to achievements
				roleFilter := s.buildRoleBasedFilters(askingUser)
				roleFilter(q)

				// Order achievements by creation time (newest first)
				q.Order(ent.Desc(achievement.FieldCreatedAt))
			})

		// Order users by name for consistent pagination
		query = query.Order(ent.Asc(user.FieldLastName), ent.Asc(user.FieldFirstName))

		start = time.Now()
		userList, err := query.
			Offset(offset).
			Limit(limit).
			All(ctx)
		statsRec.Add(events.PostgresQueries, 1)
		statsRec.Add(events.PostgresTime, time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to query users with achievements: %w", err)
		}

		users = userList
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return users, totalUsers, nil
}
