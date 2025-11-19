package companyservice

import (
	"context"
	"time"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
)

type DataSource interface {
	User(context.Context, companyquery.User) (company.User, error)
	Users(context.Context, companyquery.Users) ([]company.User, error)
	Department(context.Context, companyquery.Department) (company.Department, error)
	Departments(context.Context, companyquery.Departments) ([]company.Department, error)

	// UsersWithIDs must return the the list of all users with the given ids.
	// The length of this list must be equal to the length of the list of ids.
	// The order of the users must be equal to the order of the originally provided ids.
	//
	// When a user does not exist, the corresponding element in the list must be
	// created using company.ExEmployee.
	UsersWithIDs(context.Context, []string) ([]company.User, error)
}

type S struct {
	DataSource
}

func New(ds DataSource) S {
	return S{ds}
}

func NewLocal(users []company.User, userPasswords map[string]string, departments []company.Department) S {
	return New(&localDS{
		users:                users,
		userPasswords:        userPasswords,
		departments:          departments,
		slowdownDuration:     200 * time.Millisecond,
		longSlowdownDuration: 2 * time.Second,
	})
}
