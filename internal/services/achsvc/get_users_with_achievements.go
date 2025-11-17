package achsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// getUsersWithAchievements retrieves users that have achievements with pagination.
func (s *ACS) getUsersWithAchievements(
	ctx context.Context,
	whosAsking string,
	offset, limit int,
	search string,
) ([]company.User, int, error) {
	rec := event.Get(ctx).Sub("achsvc/get_users_with_achievements")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"whos_asking", whosAsking,
		"offset", offset,
		"limit", limit,
		"search", search,
	)

	var askingUser company.User
	err := rec.Operation("query_asking_user", func(rec *event.Record) (err error) {
		rec.Sub("params").Set(
			"asking_user_id", whosAsking,
		)

		askingUser, err = s.company.User(ctx, companyquery.User{ID: whosAsking})
		if err != nil {
			return fmt.Errorf("failed to get asking user: %w", err)
		}

		rec.Set("asking_user", askingUser)

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	var totalUsers int
	var userIDs []string

	roleFilter := s.buildRoleBasedFilters(askingUser, false)

	err = txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		err := rec.Operation("count_users", func(_ *event.Record) error {
			start := time.Now()
			count, err := tx.Achievement.Query().
				Where(roleFilter).
				Select(entAchievement.FieldOwnerID).
				Unique(true).
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
			return err
		}

		err = rec.Operation("query_user_ids", func(_ *event.Record) error {
			var err error
			userIDs, err = tx.Achievement.Query().
				Where(roleFilter).
				Unique(true).
				Limit(limit).
				Offset(offset).
				Select(entAchievement.FieldOwnerID).
				Strings(ctx)

			statsRec.Add(events.PostgresQueries, 1)

			if err != nil {
				return fmt.Errorf("failed to query users with achievements: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	users, err := s.company.UsersWithIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users with ids: %w", err)
	}
	return users, totalUsers, nil
}
