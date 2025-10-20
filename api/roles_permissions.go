package api

import (
	"net/http"

	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type RolesResponse struct {
	Roles []respond.Role `json:"roles"`
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
		Roles: make([]respond.Role, len(sesc.Roles)),
	}
	for i, role := range company.Roles {
		response.Roles[i] = respond.WithRole(role)
	}

	a.writeJSON(ctx, w, response)
}
