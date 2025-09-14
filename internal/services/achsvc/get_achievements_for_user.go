package achsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/internal/services/txhelper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// GetUserAchievements retrieves all achievements for the current user with pagination.
// Results are ordered based on the asking user's role and review responsibilities.
// todo limit pagination discuss tx
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

	err := txhelper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		var askingUser *ent.User
		err := rec.Operation("query_asking_user", func(rec *event.Record) (err error) {
			rec.Sub("params").Set(
				"asking_user_id", whosAsking,
			)

			start := time.Now()
			askingUser, err = tx.User.Query().
				Where(user.ID(whosAsking)).
				WithDepartment().
				Only(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return sesc.ErrUserNotFound
			}

			if err != nil {
				return fmt.Errorf("failed to get asking user: %w", err)
			}

			rec.Set("asking_user", askingUser)

			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("count_achievements", func(rec *event.Record) error {
			rec.Sub("params").Set(
				"asking_user_role", askingUser.Role.String(),
				"owner_id", userID,
			)
			roleFilter := s.buildRoleBasedFilters(askingUser)

			start := time.Now()
			count, err := tx.Achievement.Query().
				Where(entAchievement.OwnerID(userID)).
				Where(roleFilter).
				Order(ent.Desc(entAchievement.FieldStatus)).
				Count(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to count achievements: %w", err)
			}

			totalAchievements = count
			rec.Set("total", count)
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("query_achievements", func(rec *event.Record) error {
			rec.Sub("params").Set(
				"asking_user_role", askingUser.Role.String(),
				"limit", limit,
				"offset", offset,
				"owner_id", userID,
			)
			roleFilter := s.buildRoleBasedFilters(askingUser)

			start := time.Now()
			entities, err := tx.Achievement.Query().
				Where(entAchievement.OwnerID(userID)).
				Where(roleFilter).
				Order(ent.Desc(entAchievement.FieldStatus)).
				Offset(offset).
				Limit(limit).
				WithDocuments().
				WithOwner().
				WithReviews().
				WithTemplate().
				All(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to query achievements: %w", err)
			}

			rec.Set("n_enities", len(entities))
			achievementEntities = entities
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return achievementEntities, totalAchievements, err
}
