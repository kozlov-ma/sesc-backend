package usersvc

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/department"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/predicate"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// Users gets all users.
func (s *USS) Users(
	ctx context.Context,
	offset, limit int,
	search string,
) (ent.Users, int, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/users")
	rec.Sub("params").Set(
		"offset", offset,
		"limit", limit,
		"search", search,
	)

	statrec := event.Root(ctx).Sub("stats")

	var (
		users ent.Users
		total int
	)
	err := withTx(ctx, s.client, func(tx *ent.Tx) error {
		err := rec.Operation("query_users", func(_ *event.Record) error {
			filters := userFilters(search)
			uq := tx.User.Query().WithDepartment()
			if len(filters) > 0 {
				uq = uq.Where(user.Or(filters...))
			}

			statrec.Add(events.PostgresQueries, 1)

			var err error
			startTime := time.Now()
			users, err = uq.Offset(offset).Limit(limit).All(ctx)
			statrec.Add(events.PostgresTime, time.Since(startTime))
			if err != nil {
				return fmt.Errorf("couldn't query users: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("count_total_users", func(_ *event.Record) error {
			filters := userFilters(search)
			uq := tx.User.Query().WithDepartment()
			if len(filters) > 0 {
				uq = uq.Where(user.Or(filters...))
			}

			var err error
			startTime := time.Now()
			total, err = uq.Count(ctx)
			statrec.Add(events.PostgresTime, time.Since(startTime))
			if err != nil {
				return fmt.Errorf("couldn't query users: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return users, total, err
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

	var n1, n2, n3 string
	switch len(split) {
	case 3:
		n1, n2, n3 = split[0], split[1], split[2]
	case 2:
		n1, n2, n3 = split[0], split[1], split[1]
	case 1:
		n1, n2, n3 = split[0], split[0], split[0]
	default:
		return nil
	}

	return []predicate.User{
		user.FirstNameContainsFold(n1),
		user.LastNameContainsFold(n2),
		user.LastNameContainsFold(n3),
	}
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
