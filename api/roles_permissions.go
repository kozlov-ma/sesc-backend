package api

import (
	"net/http"

	"github.com/kozlov-ma/sesc-backend/sesc"
)

type RolesResponse struct {
	Roles []Role `json:"roles"`
}

type Role struct {
	ID       int    `json:"id"       example:"1"             validate:"required"`
	Name     string `json:"name"     example:"Преподаватель" validate:"required"`
	CodeName string `json:"codeName" example:"teacher"       validate:"required"`
}

// Roles godoc
// @Summary List all roles
// @Description Retrieves all system roles with their permissions
// @Tags roles
// @Produce json
// @Success 200 {object} RolesResponse
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /roles [get]
func (a *API) Roles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response := RolesResponse{
		Roles: make([]Role, len(sesc.Roles)),
	}
	for i, role := range sesc.Roles {
		response.Roles[i] = convertRole(role)
	}

	a.writeJSON(ctx, w, response)
}

func convertRole(r sesc.Role) Role {
	return Role{
		ID:       int(r),
		Name:     r.Name(),
		CodeName: r.String(),
	}
}
