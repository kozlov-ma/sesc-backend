package respond

import (
	"github.com/kozlov-ma/sesc-backend/company"
)

type Department struct {
	ID          string `json:"id"          example:"kaf_bio"                                                             validate:"required"`
	Name        string `json:"name"        example:"Кафедра Химии и Биологии"                                            validate:"required"`
	Description string `json:"description" example:"Мы в сунце очень любим делать string id. Потому что зачем нам uuid?" validate:"required"`
}

type Departments struct {
	Departments []*Department `json:"departments" validate:"required"`
	Total       int           `json:"total"       validate:"required"`
}

func WithDepartment(d company.Department) *Department {
	if d == (company.Department{}) {
		return nil
	}
	return &Department{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
	}
}

func WithDepartments(dd []company.Department, total int) Departments {
	deps := make([]*Department, len(dd))
	for i, d := range dd {
		deps[i] = WithDepartment(d)
	}

	return Departments{
		Departments: deps,
		Total:       total,
	}
}
