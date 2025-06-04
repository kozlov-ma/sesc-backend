package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type CreateDepartmentRequest struct {
	Name        string `json:"name"        example:"Mathematics"     validate:"required"`
	Description string `json:"description" example:"Math department" validate:"required"`
}

type UpdateDepartmentRequest struct {
	Name        string `json:"name"        example:"Mathematics"     validate:"required"`
	Description string `json:"description" example:"Math department" validate:"required"`
}

// CreateDepartment godoc
// @Summary Create a new department
// @Description Creates a new department with the given details
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param request body CreateDepartmentRequest true "Department details"
// @Success 201 {object} respond.Department
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 409 {object} respond.Error "Department with this name already exists"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /departments [post]
func (a *API) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateDepartmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	dep, err := a.sesc.CreateDepartment(ctx, req.Name, req.Description)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("couldn't create department: %w", err))
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithStatus(respond.WithDepartment(dep), http.StatusCreated))
}

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

// UpdateDepartment godoc
// @Summary Update department details
// @Description Updates an existing department with new details
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Department UUID"
// @Param request body UpdateDepartmentRequest true "Updated department details"
// @Success 200 {object} respond.Department
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 404 {object} respond.Error "Department not found"
// @Failure 409 {object} respond.Error "Department with this name already exists"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /departments/{id} [put]
func (a *API) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	rec := event.Get(ctx)

	var id uuid.UUID
	if err := (&id).Parse(idStr); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	dep, err := a.sesc.UpdateDepartment(ctx, id, req.Name, req.Description)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithDepartment(dep))
}

// DeleteDepartment godoc
// @Summary Delete a department
// @Description Deletes a department by its ID
// @Tags departments
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Department UUID"
// @Success 204 "No content"
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 404 {object} respond.Error "Department not found"
// @Failure 409 {object} respond.Error "Cannot remove department, it still has some users"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /departments/{id} [delete]
func (a *API) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	rec := event.Get(ctx)

	var id uuid.UUID
	if err := (&id).Parse(idStr); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	err := a.sesc.DeleteDepartment(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDepartment godoc
// @Summary Get department details
// @Description Retrieves detailed information about a department
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User UUID"
// @Success 200 {object} respond.Department
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 404 {object} respond.Error "Department not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /departments/{id} [get]
func (a *API) GetDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")

	depID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	dep, err := a.sesc.DepartmentByID(ctx, depID)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithDepartment(dep))
}
