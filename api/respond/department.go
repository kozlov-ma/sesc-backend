package respond

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
)

type Department struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name        string    `json:"name"        example:"Mathematics"                          validate:"required"`
	Description string    `json:"description" example:"Math department"                      validate:"required"`
}

type Departments struct {
	Departments []*Department `json:"departments" validate:"required"`
	Total       int           `json:"total"       validate:"required"`
}

func WithDepartment(d *ent.Department) *Department {
	if d == nil {
		return nil
	}
	return &Department{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
	}
}

func WithDepartments(dd ent.Departments, total int) Departments {
	deps := make([]*Department, len(dd))
	for i, d := range dd {
		deps[i] = WithDepartment(d)
	}

	return Departments{
		Departments: deps,
		Total:       total,
	}
}
