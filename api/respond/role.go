package respond

import (
	"github.com/kozlov-ma/sesc-backend/sesc"
)

func WithRole(r sesc.Role) Role {
	return Role{
		ID:       int(r),
		Name:     r.Name(),
		CodeName: r.String(),
	}
}

type Role struct {
	ID       int    `json:"id"       example:"1"             validate:"required"`
	Name     string `json:"name"     example:"Преподаватель" validate:"required"`
	CodeName string `json:"codeName" example:"teacher"       validate:"required"`
}
