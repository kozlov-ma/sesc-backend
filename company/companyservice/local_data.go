package companyservice

import (
	"context"
	"strings"
	"time"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
)

var _ DataSource = (*localDS)(nil)

type localDS struct {
	storage              *storage
	userPasswords        map[string]string
	slowdownDuration     time.Duration
	longSlowdownDuration time.Duration
}

func newLocalDS(users []company.User, userPasswords map[string]string, departments []company.Department) *localDS {
	return &localDS{
		storage:              newStorage(users, departments),
		userPasswords:        userPasswords,
		slowdownDuration:     200 * time.Millisecond,
		longSlowdownDuration: 2 * time.Second,
	}
}

func (l *localDS) Department(ctx context.Context, q companyquery.Department) (company.Department, error) {
	l.slowdown()
	return l.storage.queryDepartment(ctx, q)
}

func (l *localDS) Departments(ctx context.Context, q companyquery.Departments) ([]company.Department, error) {
	l.longSlowdown()
	return l.storage.queryDepartments(ctx, q)
}

func (l *localDS) User(ctx context.Context, q companyquery.User) (company.User, error) {
	l.slowdown()

	verifyPassword := func(userID, password string) error {
		if expected, ok := l.userPasswords[userID]; ok && expected == password {
			return nil
		}
		return company.ErrUserNotFound
	}

	return l.storage.queryUser(ctx, q, verifyPassword)
}

func (l *localDS) Users(ctx context.Context, q companyquery.Users) ([]company.User, error) {
	l.longSlowdown()
	return l.storage.queryUsers(ctx, q)
}

func (l *localDS) UsersWithIDs(ctx context.Context, ids []string) ([]company.User, error) {
	l.longSlowdown()
	return l.storage.usersWithIDs(ctx, ids)
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
