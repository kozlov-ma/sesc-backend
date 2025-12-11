package companyservice

import (
	"context"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
)

type DataSource interface {
	User(context.Context, companyquery.User) (company.User, error)
	Users(context.Context, companyquery.Users) ([]company.User, error)
	UsersWithIDs(context.Context, []string) ([]company.User, error)
	Department(context.Context, companyquery.Department) (company.Department, error)
	Departments(context.Context, companyquery.Departments) ([]company.Department, error)
}

type S struct {
	DataSource
}

func New(ds DataSource) S {
	return S{ds}
}

func NewLocal(users []company.User, userPasswords map[string]string, departments []company.Department) S {
	return New(newLocalDS(users, userPasswords, departments))
}

// NewLDAPService creates a new company service backed by LDAP
func NewLDAPService(ctx context.Context, config LDAPConfig, eventSink EventSink) (S, error) {
	ds, err := NewLDAP(ctx, config, eventSink)
	if err != nil {
		return S{}, err
	}
	return New(ds), nil
}
