package api

import (
	"fmt"
	"net/http"

	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// Departments godoc
// @Summary List all departments
// @Description Retrieves list of all registered departments
// @Tags departments
// @Produce json
// @Success 200 {object} respond.Departments
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /departments [get]
func (a *API) Departments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	deps, err := a.sesc.Departments(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("couldn't get departments: %w", err))
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithDepartments(deps, len(deps)))
}

// GetDepartment godoc
// @Summary Get department details
// @Description Retrieves detailed information about a department
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Department ID"
// @Success 200 {object} respond.Department
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 404 {object} respond.Error "Department not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /departments/{id} [get]
func (a *API) GetDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")

	dep, err := a.sesc.DepartmentByID(ctx, idStr)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithDepartment(dep))
}
