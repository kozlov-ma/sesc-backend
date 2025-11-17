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

func WithRoles(rr []company.Role) []Role {
	var res []Role
	for _, r := range rr {
		res = append(res, WithRole(r))
	}
	return res
}

type Role struct {
	Name     string `json:"name"     example:"Преподаватель" validate:"required"`
	CodeName string `json:"codeName" example:"teacher"       validate:"required"`
}
