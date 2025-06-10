package achsvc

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/department"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/predicate"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// GetUsersWithAchievements retrieves users that have achievements with pagination.
func (s *ACS) GetUsersWithAchievements(
	ctx context.Context,
	whosAsking UUID,
	offset, limit int,
	search string,
) (ent.Users, int, error) {
	rec := event.Get(ctx).Sub("achsvc/get_users_with_achievements")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"whos_asking", whosAsking,
		"offset", offset,
		"limit", limit,
		"search", search,
	)

	var totalUsers int
	var users []*ent.User

	err := withTx(ctx, s.client, func(tx *ent.Tx) error {
		err := rec.Operation("count_users", func(_ *event.Record) error {
			// Get asking user for role-based filtering
			start := time.Now()
			askingUser, err := tx.User.Query().
				Where(user.ID(whosAsking)).
				WithDepartment().
				Only(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to get asking user: %w", err)
			}

			roleFilter := s.buildRoleBasedFilters(askingUser)

			filters := userFilters(search)
			uq := tx.User.Query().WithDepartment()
			if len(filters) > 0 {
				uq = uq.Where(user.Or(filters...))
			}

			start = time.Now()
			count, err := uq.
				Where(user.HasAchievementsWith(roleFilter)).
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

		err = rec.Operation("query_users", func(_ *event.Record) error {
			// Get asking user to determine ordering
			start := time.Now()
			askingUser, err := tx.User.Query().
				Where(user.ID(whosAsking)).
				WithDepartment().
				Only(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to get asking user: %w", err)
			}

			// Apply role-based filtering
			roleFilter := s.buildRoleBasedFilters(askingUser)

			filters := userFilters(search)
			uq := tx.User.Query().WithDepartment()
			if len(filters) > 0 {
				uq = uq.Where(user.Or(filters...))
			}

			start = time.Now()
			userList, err := uq.
				Where(user.HasAchievementsWith(roleFilter)).
				WithDepartment().
				WithAchievements(func(q *ent.AchievementQuery) {
					q.Where(roleFilter).
						WithTemplate(func(tq *ent.AchievementTemplateQuery) {
							tq.WithGroup()
						}).
						WithDocuments(func(dq *ent.AchievementDocumentQuery) {
							dq.WithFile()
						}).
						WithReviews(func(rq *ent.AchievementReviewQuery) {
							rq.WithReviewer()
						}).
						Order(ent.Desc(entAchievement.FieldStatus))
				}).
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
			return err
		}

		return nil
	})

	return users, totalUsers, err
}

func userFilters(search string) []predicate.User {
	return slices.Concat(
		nameFilter(search),
		roleFilter(search),
		departmentFilter(search),
	)
}

func nameFilter(search string) []predicate.User {
	if search == "" {
		return nil
	}

	split := strings.Fields(search)

	if len(split) > 3 {
		return nil
	}

	var pre []predicate.User

	switch len(split) {
	case 3:
		n1, n2, n3 := split[0], split[1], split[2]
		pre = append(
			pre,
			user.And(
				user.LastNameContainsFold(n1),
				user.FirstNameContainsFold(n2),
				user.MiddleNameContainsFold(n3),
			),
			user.And(
				user.FirstNameContainsFold(n1),
				user.MiddleNameContainsFold(n2),
				user.LastNameContainsFold(n3),
			),
		)
	case 2:
		n1, n2 := split[0], split[1]
		pre = append(
			pre,
			user.And(
				user.LastNameContainsFold(n1),
				user.FirstNameContainsFold(n2),
			),
			user.And(
				user.FirstNameContainsFold(n1),
				user.LastNameContainsFold(n2),
			),
			user.And(
				user.FirstNameContainsFold(n1),
				user.MiddleNameContainsFold(n2),
			),
		)
	case 1:
		n1 := split[0]
		pre = append(
			pre,
			user.LastNameContainsFold(n1),
			user.FirstNameContainsFold(n1),
		)
	default:
		return nil
	}

	return pre
}

func departmentFilter(search string) []predicate.User {
	if search == "" {
		return nil
	}

	return []predicate.User{
		user.HasDepartmentWith(department.NameContainsFold(search)),
	}
}

func roleFilter(search string) []predicate.User {
	if search == "" {
		return nil
	}

	lowerRoleName := strings.ToLower(search)
	var pre []predicate.User

	if containsAnyOf(lowerRoleName, "преп", "tea") {
		pre = append(pre, user.Role(sesc.Teacher))
	}

	if containsAnyOf(lowerRoleName, "зав", "dep") {
		pre = append(pre, user.Role(sesc.Dephead))
	}

	if containsAnyOf(lowerRoleName, "оли", "oly", "cont") {
		pre = append(pre, user.Role(sesc.OlympiadDeputy))
	}

	if containsAnyOf(lowerRoleName, "науч", "наук", "sci") {
		pre = append(pre, user.Role(sesc.ScientificDeputy))
	}

	if containsAnyOf(lowerRoleName, "разв", "deve") {
		pre = append(pre, user.Role(sesc.DevelopmentDeputy))
	}

	if containsAnyOf(lowerRoleName, "ака", "aca") {
		pre = append(pre, user.Role(sesc.AcademicDirector))
	}

	if containsAnyOf(lowerRoleName, "эко", "eco", "вед", "chi") {
		pre = append(pre, user.Role(sesc.ChiefEconomist))
	}

	if len(pre) > 0 {
		return pre
	}

	if containsAnyOf(lowerRoleName, "дир") {
		pre = append(
			pre,
			user.Role(sesc.AcademicDirector),
			user.Role(sesc.DevelopmentDeputy),
			user.Role(sesc.OlympiadDeputy),
			user.Role(sesc.ScientificDeputy),
		)
	}

	return pre
}

func containsAnyOf(s string, variants ...string) bool {
	for _, v := range variants {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}
