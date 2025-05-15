package api

import (
	"net/http"

	"github.com/kozlov-ma/sesc-backend/sesc"
)

type RolesResponse struct {
	Roles []Role `json:"roles"`
}

type Role struct {
	ID          int32        `json:"id"          example:"1"             validate:"required"`
	Name        string       `json:"name"        example:"Преподаватель" validate:"required"`
	Permissions []Permission `json:"permissions"                         validate:"required"`
}

type PermissionsResponse struct {
	Permissions []Permission `json:"permissions" validate:"required"`
}

type Permission struct {
	ID          int32  `json:"id"          example:"1"                                      validate:"required"`
	Name        string `json:"name"        example:"draft_achievement_list"                 validate:"required"`
	Description string `json:"description" example:"Создание и заполнение листа достижений" validate:"required"`
}

// Roles godoc
// @Summary List all roles
// @Description Retrieves all system roles with their permissions
// @Tags roles
// @Produce json
// @Success 200 {object} RolesResponse
// @Failure 500 {object} Error "Internal server error"
// @Router /roles [get]
func (a *API) Roles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response := RolesResponse{
		Roles: make([]Role, len(sesc.Roles)),
	}
	for i, role := range sesc.Roles {
		response.Roles[i] = convertRole(role)
	}

	a.writeJSON(ctx, w, response, http.StatusOK)
}

// Permissions godoc
// @Summary List all permissions
// @Description Retrieves all available system permissions
// @Tags permissions
// @Produce json
// @Success 200 {object} PermissionsResponse
// @Router /permissions [get]
func (a *API) Permissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	perms := sesc.Permissions
	response := PermissionsResponse{
		Permissions: make([]Permission, len(perms)),
	}

	for i, p := range perms {
		response.Permissions[i] = Permission{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	a.writeJSON(ctx, w, response, http.StatusOK)
}

func convertRole(r sesc.Role) Role {
	return Role{
		ID:          r.ID,
		Name:        r.Name,
		Permissions: convertPermissions(r.Permissions),
	}
}

func convertPermissions(perms []sesc.Permission) []Permission {
	res := make([]Permission, len(perms))
	for i, p := range perms {
		res[i] = Permission{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}
	return res
}
