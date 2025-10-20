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

// Variable to load achievements with users for testing purposes.
//
// Preferrably, should be removed.
var includeAchievementsForTests = false

// ACCESSTODO
// GetUsersWithAchievements retrieves users that have achievements with pagination.
func (s *ACS) GetUsersWithAchievements(
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

	uu, err := s.company.UsersWithIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users with ids: %w", err)
	}

	return uu, totalUsers, err
}

// func VuserFilters(search string) []predicate.User {
// 	return slices.Concat(
// 		nameFilter(search),
// 		roleFilter(search),
// 		departmentFilter(search),
// 	)
// }

// func nameFilter(search string) []predicate.User {
// 	if search == "" {
// 		return nil
// 	}

// 	split := strings.Fields(search)

// 	if len(split) > 3 {
// 		return nil
// 	}

// 	var pre []predicate.User

// 	switch len(split) {
// 	case 3:
// 		n1, n2, n3 := split[0], split[1], split[2]
// 		pre = append(
// 			pre,
// 			user.And(
// 				user.LastNameContainsFold(n1),
// 				user.FirstNameContainsFold(n2),
// 				user.MiddleNameContainsFold(n3),
// 			),
// 			user.And(
// 				user.FirstNameContainsFold(n1),
// 				user.MiddleNameContainsFold(n2),
// 				user.LastNameContainsFold(n3),
// 			),
// 		)
// 	case 2:
// 		n1, n2 := split[0], split[1]
// 		pre = append(
// 			pre,
// 			user.And(
// 				user.LastNameContainsFold(n1),
// 				user.FirstNameContainsFold(n2),
// 			),
// 			user.And(
// 				user.FirstNameContainsFold(n1),
// 				user.LastNameContainsFold(n2),
// 			),
// 			user.And(
// 				user.FirstNameContainsFold(n1),
// 				user.MiddleNameContainsFold(n2),
// 			),
// 		)
// 	case 1:
// 		n1 := split[0]
// 		pre = append(
// 			pre,
// 			user.LastNameContainsFold(n1),
// 			user.FirstNameContainsFold(n1),
// 		)
// 	default:
// 		return nil
// 	}

// 	return pre
// }

// func departmentFilter(search string) []predicate.User {
// 	if search == "" {
// 		return nil
// 	}

// 	return []predicate.User{
// 		user.HasDepartmentWith(department.NameContainsFold(search)),
// 	}
// }

// func roleFilter(search string) []predicate.User {
// 	if search == "" {
// 		return nil
// 	}

// 	lowerRoleName := strings.ToLower(search)
// 	var pre []predicate.User

// 	if containsAnyOf(lowerRoleName, "преп", "tea") {
// 		pre = append(pre, user.Role(sesc.Teacher))
// 	}

// 	if containsAnyOf(lowerRoleName, "зав", "dep") {
// 		pre = append(pre, user.Role(sesc.Dephead))
// 	}

// 	if containsAnyOf(lowerRoleName, "оли", "oly", "cont") {
// 		pre = append(pre, user.Role(sesc.OlympiadDeputy))
// 	}

// 	if containsAnyOf(lowerRoleName, "науч", "наук", "sci") {
// 		pre = append(pre, user.Role(sesc.ScientificDeputy))
// 	}

// 	if containsAnyOf(lowerRoleName, "разв", "deve") {
// 		pre = append(pre, user.Role(sesc.DevelopmentDeputy))
// 	}

// 	if containsAnyOf(lowerRoleName, "ака", "aca") {
// 		pre = append(pre, user.Role(sesc.AcademicDirector))
// 	}

// 	if containsAnyOf(lowerRoleName, "эко", "eco", "вед", "chi") {
// 		pre = append(pre, user.Role(sesc.ChiefEconomist))
// 	}

// 	if len(pre) > 0 {
// 		return pre
// 	}

// 	if containsAnyOf(lowerRoleName, "дир") {
// 		pre = append(
// 			pre,
// 			user.Role(sesc.AcademicDirector),
// 			user.Role(sesc.DevelopmentDeputy),
// 			user.Role(sesc.OlympiadDeputy),
// 			user.Role(sesc.ScientificDeputy),
// 		)
// 	}

// 	return pre
// }

// func containsAnyOf(s string, variants ...string) bool {
// 	for _, v := range variants {
// 		if strings.Contains(s, v) {
// 			return true
// 		}
// 	}
// 	return false
// }
