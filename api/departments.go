package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type Department struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name        string    `json:"name"        example:"Mathematics"                          validate:"required"`
	Description string    `json:"description" example:"Math department"                      validate:"required"`
}

type CreateDepartmentRequest struct {
	Name        string `json:"name"        example:"Mathematics"     validate:"required"`
	Description string `json:"description" example:"Math department" validate:"required"`
}

type CreateDepartmentResponse = Department

type DepartmentsResponse struct {
	Departments []Department `json:"departments" validate:"required"`
}

type UpdateDepartmentRequest struct {
	Name        string `json:"name"        example:"Mathematics"     validate:"required"`
	Description string `json:"description" example:"Math department" validate:"required"`
}

type UpdateDepartmentResponse = Department

type DepartmentNotFoundError struct {
	Code      string `json:"code"             example:"DEPARTMENT_NOT_FOUND"`
	Message   string `json:"message"          example:"Department not found"`
	RuMessage string `json:"ruMessage"        example:"Кафедра не найдена"`
	Details   string `json:"details,omitzero"`
}

func (e DepartmentNotFoundError) WithDetails(details string) DepartmentNotFoundError {
	e.Details = details
	return e
}

type InvalidDepartmentIDError struct {
	Code      string `json:"code"             example:"INVALID_DEPARTMENT_ID"`
	Message   string `json:"message"          example:"Invalid department ID"`
	RuMessage string `json:"ruMessage"        example:"Некорректный идентификатор кафедры"`
	Details   string `json:"details,omitzero"`
}

func (e InvalidDepartmentIDError) WithDetails(details string) InvalidDepartmentIDError {
	e.Details = details
	return e
}

type InvalidDepartmentError struct {
	Code      string `json:"code"             example:"INVALID_DEPARTMENT"`
	Message   string `json:"message"          example:"Invalid department data"`
	RuMessage string `json:"ruMessage"        example:"Некорректные данные кафедры"`
	Details   string `json:"details,omitzero"`
}

func (e InvalidDepartmentError) WithDetails(details string) InvalidDepartmentError {
	e.Details = details
	return e
}

type DepartmentExistsError struct {
	Code      string `json:"code"             example:"DEPARTMENT_EXISTS"`
	Message   string `json:"message"          example:"Department with this name already exists"`
	RuMessage string `json:"ruMessage"        example:"Кафедра с таким названием уже существует"`
	Details   string `json:"details,omitzero"`
}

func (e DepartmentExistsError) WithDetails(details string) DepartmentExistsError {
	e.Details = details
	return e
}

type CannotRemoveDepartmentError struct {
	Code      string `json:"code"             example:"CANNOT_REMOVE_DEPARTMENT"`
	Message   string `json:"message"          example:"Cannot remove department, it still has some users"`
	RuMessage string `json:"ruMessage"        example:"Невозможно удалить кафедру, так как она содержит пользователей"`
	Details   string `json:"details,omitzero"`
}

func (e CannotRemoveDepartmentError) WithDetails(details string) CannotRemoveDepartmentError {
	e.Details = details
	return e
}

var (
	ErrDepartmentNotFound = DepartmentNotFoundError{
		Code:      "DEPARTMENT_NOT_FOUND",
		Message:   "Department not found",
		RuMessage: "Кафедра не найдена",
	}
	ErrInvalidDepartmentID = InvalidDepartmentIDError{
		Code:      "INVALID_DEPARTMENT_ID",
		Message:   "Invalid department ID",
		RuMessage: "Некорректный идентификатор кафедры",
	}
	ErrInvalidDepartment = InvalidDepartmentError{
		Code:      "INVALID_DEPARTMENT",
		Message:   "Invalid department data",
		RuMessage: "Некорректные данные кафедры",
	}
	ErrDepartmentExists = DepartmentExistsError{
		Code:      "DEPARTMENT_EXISTS",
		Message:   "Department with this name already exists",
		RuMessage: "Кафедра с таким названием уже существует",
	}
	ErrCannotRemoveDepartment = CannotRemoveDepartmentError{
		Code:      "CANNOT_REMOVE_DEPARTMENT",
		Message:   "Cannot remove department, it still has some users",
		RuMessage: "Невозможно удалить кафедру, так как она содержит пользователей",
	}
)

// CreateDepartment godoc
// @Summary Create a new department
// @Description Creates a new department with the given details
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param request body CreateDepartmentRequest true "Department details"
// @Success 201 {object} Department
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 409 {object} DepartmentExistsError "Department with this name already exists"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /departments [post]
func (a *API) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateDepartmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(a, w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	dep, err := a.sesc.CreateDepartment(ctx, req.Name, req.Description)
	switch {
	case errors.Is(err, sesc.ErrInvalidDepartment):
		writeError(a, w, ErrDepartmentExists, http.StatusConflict)
		return
	case err != nil:
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to create department",
			RuMessage: "Ошибка создания кафедры",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, CreateDepartmentResponse{
		ID:          dep.ID,
		Name:        dep.Name,
		Description: dep.Description,
	}, http.StatusCreated)
}

// Departments godoc
// @Summary List all departments
// @Description Retrieves list of all registered departments
// @Tags departments
// @Produce json
// @Success 200 {object} DepartmentsResponse
// @Failure 500 {object} ServerError "Internal server error"
// @Router /departments [get]
func (a *API) Departments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deps, err := a.sesc.Departments(ctx)
	if err != nil {
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch departments",
			RuMessage: "Ошибка получения списка кафедр",
		}, http.StatusInternalServerError)
		return
	}

	response := DepartmentsResponse{
		Departments: make([]Department, len(deps)),
	}
	for i, d := range deps {
		response.Departments[i] = Department{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
		}
	}

	a.writeJSON(w, response, http.StatusOK)
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
// @Success 200 {object} Department
// @Failure 400 {object} InvalidDepartmentIDError "Invalid UUID format"
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} DepartmentNotFoundError "Department not found"
// @Failure 409 {object} DepartmentExistsError "Department with this name already exists"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /departments/{id} [put]
func (a *API) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	var id uuid.UUID
	if err := (&id).Parse(idStr); err != nil {
		writeError(a, w, ErrInvalidDepartmentID, http.StatusBadRequest)
		return
	}

	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(a, w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	err := a.sesc.UpdateDepartment(ctx, id, req.Name, req.Description)
	switch {
	case errors.Is(err, sesc.ErrInvalidDepartment):
		writeError(a, w, ErrDepartmentExists, http.StatusConflict)
		return
	case err != nil:
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to update department",
			RuMessage: "Ошибка обновления кафедры",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UpdateDepartmentResponse{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}, http.StatusOK)
}

// DeleteDepartment godoc
// @Summary Delete a department
// @Description Deletes a department by its ID
// @Tags departments
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Department UUID"
// @Success 204 "No content"
// @Failure 400 {object} InvalidDepartmentIDError "Invalid UUID format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} DepartmentNotFoundError "Department not found"
// @Failure 409 {object} CannotRemoveDepartmentError "Cannot remove department, it still has some users"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /departments/{id} [delete]
func (a *API) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	var id uuid.UUID
	if err := (&id).Parse(idStr); err != nil {
		writeError(a, w, ErrInvalidDepartmentID, http.StatusBadRequest)
		return
	}

	err := a.sesc.DeleteDepartment(ctx, id)
	switch {
	case errors.Is(err, sesc.ErrInvalidDepartment):
		writeError(a, w, ErrDepartmentNotFound, http.StatusNotFound)
		return
	case errors.Is(err, sesc.ErrCannotRemoveDepartment):
		writeError(a, w, ErrCannotRemoveDepartment, http.StatusConflict)
		return
	case err != nil:
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to delete department",
			RuMessage: "Ошибка удаления кафедры",
		}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
