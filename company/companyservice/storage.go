package companyservice

import (
	"context"
	"strings"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
)

type storage struct {
	users       []company.User
	usersByID   map[string]company.User
	departments []company.Department
	deptsByID   map[string]company.Department
}

func newStorage(users []company.User, departments []company.Department) *storage {
	s := &storage{
		users:       users,
		usersByID:   make(map[string]company.User, len(users)),
		departments: departments,
		deptsByID:   make(map[string]company.Department, len(departments)),
	}

	for _, u := range users {
		s.usersByID[u.ID] = u
	}

	for _, d := range departments {
		s.deptsByID[d.ID] = d
	}

	return s
}

func (s *storage) getUserByID(id string) (company.User, bool) {
	u, ok := s.usersByID[id]
	return u, ok
}

func (s *storage) getDepartmentByID(id string) (company.Department, bool) {
	d, ok := s.deptsByID[id]
	return d, ok
}

func (s *storage) usersWithIDs(ctx context.Context, ids []string) ([]company.User, error) {
	result := make([]company.User, 0, len(ids))
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if u, ok := s.usersByID[id]; ok {
			result = append(result, u)
		} else {
			result = append(result, company.ExEmployee(id))
		}
	}
	return result, nil
}

func (s *storage) queryUser(ctx context.Context, q companyquery.User, verifyPassword func(userID, password string) error) (company.User, error) {
	if u, ok := s.usersByID[q.ID]; ok {
		if err := ctx.Err(); err != nil {
			return company.User{}, err
		}

		if q.Password != "" && verifyPassword != nil {
			if err := verifyPassword(u.ID, q.Password); err != nil {
				return company.User{}, company.ErrUserNotFound
			}
		}
		return u, nil
	}

	return company.User{}, company.ErrUserNotFound
}

func (s *storage) queryUsers(ctx context.Context, q companyquery.Users) ([]company.User, error) {
	hasExactFilters := q.DepartmentID != "" || q.RoleID != ""
	hasSubstringFilters := q.Department != "" || q.FullName != "" || q.RoleName != ""

	var depNames []string
	if q.Department != "" {
		for _, d := range s.departments {
			if strings.Contains(strings.ToLower(d.Name), strings.ToLower(q.Department)) {
				depNames = append(depNames, strings.ToLower(d.Name))
			}
		}
	}

	var uu []company.User
	for _, u := range s.users {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var accepted bool

		if !hasExactFilters && !hasSubstringFilters {
			accepted = true
		} else if hasExactFilters {
			accepted = true
			if q.DepartmentID != "" {
				accepted = accepted && u.DepartmentID == q.DepartmentID
			}
			if q.RoleID != "" {
				accepted = accepted && u.HasRole(company.Role(q.RoleID))
			}
		} else {
			if q.FullName != "" && strings.Contains(strings.ToLower(u.FullName), strings.ToLower(q.FullName)) {
				accepted = true
			}
			if !accepted && q.Department != "" {
				for _, dn := range depNames {
					if strings.Contains(dn, strings.ToLower(q.Department)) {
						accepted = true
						break
					}
				}
			}
			if !accepted && q.RoleName != "" {
				roles := roleNameHeuristic(q.RoleName)
				if len(roles) > 0 {
					accepted = u.HasRole(roles...)
				}
			}
		}

		if accepted {
			uu = append(uu, u)
		}
	}

	return uu, nil
}

func (s *storage) queryDepartment(ctx context.Context, q companyquery.Department) (company.Department, error) {
	if err := ctx.Err(); err != nil {
		return company.Department{}, err
	}

	if d, ok := s.deptsByID[q.ID]; ok {
		return d, nil
	}

	return company.Department{}, company.ErrDepartmentNotFound
}

func (s *storage) queryDepartments(ctx context.Context, q companyquery.Departments) ([]company.Department, error) {
	var deps []company.Department
	for _, d := range s.departments {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if strings.Contains(strings.ToLower(d.Name), strings.ToLower(q.Name)) {
			deps = append(deps, d)
		}
	}

	return deps, nil
}
