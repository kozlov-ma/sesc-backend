// Package sescsvc provides services for managing SESC employees and departments.
package sescsvc

import (
	"context"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
	"github.com/kozlov-ma/sesc-backend/company/companyservice"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/internal/services/achsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/atsvc"
)

type SESC struct {
	*achsvc.ACS
	*atsvc.ATS
	company companyservice.S
}

func New(client *ent.Client, c companyservice.S) *SESC {
	return &SESC{
		ACS:     achsvc.New(client, c),
		ATS:     atsvc.New(client),
		company: c,
	}
}

func (s *SESC) Departments(ctx context.Context) ([]company.Department, error) {
	return s.company.Departments(ctx, companyquery.Departments{})
}

func (s *SESC) DepartmentByID(ctx context.Context, id string) (company.Department, error) {
	return s.company.Department(ctx, companyquery.Department{ID: id})
}

func (s *SESC) Users(ctx context.Context, search string) ([]company.User, error) {
	return s.company.Users(ctx, companyquery.Users{
		Department: search,
		FullName:   search,
		RoleName:   search,
	})
}

func (s *SESC) User(ctx context.Context, id string) (company.User, error) {
	return s.company.User(ctx, companyquery.User{ID: id})
}
