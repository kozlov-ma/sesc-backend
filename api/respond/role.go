package respond

import (
	"github.com/kozlov-ma/sesc-backend/company"
)

func WithRole(r company.Role) Role {
	return Role{
		Name:     r.Name(),
		CodeName: r.String(),
	}
}

type Role struct {
	Name     string `json:"name"     example:"Преподаватель" validate:"required"`
	CodeName string `json:"codeName" example:"teacher"       validate:"required"`
}
