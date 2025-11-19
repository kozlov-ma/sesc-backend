package companyservice

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
)

var _ DataSource = (*localDS)(nil)

type localDS struct {
	users                []company.User
	userPasswords        map[string]string
	departments          []company.Department
	slowdownDuration     time.Duration
	longSlowdownDuration time.Duration
}

// Department implements DataSource.
func (l *localDS) Department(ctx context.Context, q companyquery.Department) (company.Department, error) {
	l.slowdown()
	for _, d := range l.departments {
		if err := ctx.Err(); err != nil {
			return company.Department{}, err
		}
		if d.ID == q.ID {
			return d, nil
		}
	}

	return company.Department{}, company.ErrDepartmentNotFound
}

// Departments implements DataSource.
func (l *localDS) Departments(ctx context.Context, q companyquery.Departments) ([]company.Department, error) {
	l.longSlowdown()
	var deps []company.Department
	for _, d := range l.departments {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if strings.Contains(strings.ToLower(d.Name), strings.ToLower(q.Name)) {
			deps = append(deps, d)
		}
	}

	return deps, nil
}

// User implements DataSource.
func (l *localDS) User(ctx context.Context, q companyquery.User) (company.User, error) {
	l.slowdown()

	for _, u := range l.users {
		if err := ctx.Err(); err != nil {
			return company.User{}, err
		}
		if q.Password == "" && u.ID == q.ID {
			return u, nil
		}
		if u.ID == q.ID && l.userPasswords[u.ID] == q.Password {
			return u, nil
		}
	}

	return company.User{}, company.ErrUserNotFound
}

// Users implements DataSource.
func (l *localDS) Users(ctx context.Context, q companyquery.Users) ([]company.User, error) {
	deps, err := l.Departments(ctx, companyquery.Departments{Name: q.Department})
	if err != nil {
		return nil, fmt.Errorf("failed to query departments: %w", err)
	}

	var depNames []string
	for _, d := range deps {
		depNames = append(depNames, strings.ToLower(d.Name))
	}

	roles := roleNameHeuristic(q.RoleName)
	var uu []company.User
	for _, u := range l.users {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var accepted bool
		if depID := q.DepartmentID; depID != "" {
			accepted = accepted && u.DepartmentID == depID
		}
		if roleID := q.RoleID; roleID != "" {
			accepted = accepted && u.HasRole(company.Role(roleID))
		}

		if dep := strings.ToLower(q.Department); dep != "" {
			var depExists bool
			for _, dn := range depNames {
				if strings.Contains(dn, strings.ToLower(dep)) {
					depExists = true
				}
				break
			}

			accepted = accepted || depExists
		}

		accepted = accepted || strings.Contains(strings.ToLower(u.FullName), strings.ToLower(q.FullName))

		accepted = accepted || u.HasRole(roles...)

		if accepted {
			uu = append(uu, u)
		}
	}

	return uu, nil
}

func (l *localDS) UsersWithIDs(ctx context.Context, ids []string) ([]company.User, error) {
	l.longSlowdown()
	var uu []company.User
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if i := slices.IndexFunc(l.users, func(u company.User) bool {
			return u.ID == id
		}); i != -1 {
			uu = append(uu, l.users[i])
		} else {
			uu = append(uu, company.ExEmployee(id))
		}
	}

	return uu, nil
}

func (l *localDS) slowdown() {
	time.Sleep(l.slowdownDuration)
}

func (l *localDS) longSlowdown() {
	time.Sleep(l.longSlowdownDuration)
}

var allRoles = company.Roles

func roleNameHeuristic(containsFold string) []company.Role {
	if containsFold == "" {
		return allRoles
	}

	var rr []company.Role
	for _, r := range allRoles {
		if strings.Contains(strings.ToLower(r.Name()), strings.ToLower(containsFold)) ||
			strings.Contains(r.String(), strings.ToLower(containsFold)) {
			rr = append(rr, r)
		}
	}

	return rr
}
