package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
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
		start := time.Now()
		count, err := s.client.User.Query().
			Where(user.HasAchievements()).
			Count(ctx)
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
		start := time.Now()
		userList, err := s.client.User.Query().
			Where(user.HasAchievements()).
			WithDepartment().
			Order(ent.Asc(user.FieldLastName), ent.Asc(user.FieldFirstName)).
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
