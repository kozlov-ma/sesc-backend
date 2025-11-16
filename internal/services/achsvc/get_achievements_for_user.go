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

// getUserAchievements retrieves all achievements for the current user with pagination.
// Results are ordered based on the asking user's role and review responsibilities.
func (s *ACS) getUserAchievements(
	ctx context.Context,
	targetUserID string,
	offset, limit int,
	requireChanges bool,
) (ent.Achievements, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_user_achievements")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"whos_asking", targetUserID,
		"offset", offset,
		"limit", limit,
		"require_changes", requireChanges,
	)

	var totalAchievements int
	var achievementEntities []*ent.Achievement

	var askingUser company.User
	err := rec.Operation("query_asking_user", func(rec *event.Record) (err error) {
		rec.Sub("params").Set(
			"asking_user_id", targetUserID,
		)

		askingUser, err = s.company.User(ctx, companyquery.User{ID: targetUserID})
		if err != nil {
			return fmt.Errorf("failed to get asking user: %w", err)
		}

		rec.Set("asking_user", askingUser)

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	err = txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		err := rec.Operation("count_achievements", func(rec *event.Record) error {
			rec.Sub("params").Set(
				"asking_user_roles", askingUser.Roles,
				"owner_id", targetUserID,
			)
			roleFilter := s.buildRoleBasedFilters(askingUser, requireChanges)

			start := time.Now()
			count, err := tx.Achievement.Query().
				Where(entAchievement.OwnerID(targetUserID)).
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
				"asking_user_roles", askingUser.Roles,
				"limit", limit,
				"offset", offset,
				"owner_id", targetUserID,
			)
			roleFilter := s.buildRoleBasedFilters(askingUser, requireChanges)

			start := time.Now()
			entities, err := tx.Achievement.Query().
				Where(entAchievement.OwnerID(targetUserID)).
				Where(roleFilter).
				Order(ent.Desc(entAchievement.FieldStatus)).
				Offset(offset).
				Limit(limit).
				WithDocuments().
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
