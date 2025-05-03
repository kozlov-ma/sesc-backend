package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type Department struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name"        example:"Mathematics"`
	Description string    `json:"description" example:"Math department"`
}

type CreateDepartmentRequest struct {
	Name        string `json:"name"        example:"Mathematics"`
	Description string `json:"description" example:"Math department"`
}

type CreateDepartmentResponse = Department

type DepartmentsResponse struct {
	Departments []Department `json:"departments"`
}

type UpdateDepartmentRequest struct {
	Name        string `json:"name"        example:"Mathematics"`
	Description string `json:"description" example:"Math department"`
}

type UpdateDepartmentResponse = Department

var (
	ErrDepartmentNotFound = APIError{
		Code:      "DEPARTMENT_NOT_FOUND",
		Message:   "Department not found",
		RuMessage: "Кафедра не найдена",
	}

	ErrInvalidDepartmentID = APIError{
		Code:      "INVALID_DEPARTMENT_ID",
		Message:   "Invalid department ID",
		RuMessage: "Некорректный идентификатор кафедры",
	}

	ErrInvalidDepartment = APIError{
		Code:      "INVALID_DEPARTMENT",
		Message:   "Invalid department data",
		RuMessage: "Некорректные данные кафедры",
	}

	ErrDepartmentExists = APIError{
		Code:      "DEPARTMENT_EXISTS",
		Message:   "Department with this name already exists",
		RuMessage: "Кафедра с таким названием уже существует",
	}

	ErrCannotRemoveDepartment = APIError{
		Code:      "CANNOT_REMOVE_DEPARTMENT",
		Message:   "Cannot remove department, it still has some users",
		RuMessage: "Невозможно удалить кафедру, так как она содержит пользователей",
	}
)

// CreateDepartment godoc
// @Summary Create new department
// @Description Creates a new department with provided name and description
// @Tags departments
// @Accept json
// @Produce json
// @Param request body CreateDepartmentRequest true "Department creation data"
// @Success 201 {object} CreateDepartmentResponse
// @Failure 400 {object} APIError "Invalid request format"
// @Failure 409 {object} APIError "Department with this name already exists"
// @Failure 500 {object} APIError "Internal server error"
// @Router /departments [post]
func (a *API) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateDepartmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	dep, err := a.sesc.CreateDepartment(ctx, req.Name, req.Description)
	if errors.Is(err, sesc.ErrInvalidDepartment) {
		a.writeError(w, ErrDepartmentExists, http.StatusConflict)
		return
	}
	if err != nil {
		a.writeError(w, APIError{
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
// @Failure 500 {object} APIError "Internal server error"
// @Router /departments [get]
func (a *API) Departments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deps, err := a.sesc.Departments(ctx)
	if err != nil {
		a.writeError(w, APIError{
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

func (a *API) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ids := r.PathValue("id")

	var id uuid.UUID
	if err := (&id).Parse(ids); err != nil {
		a.writeError(w, ErrInvalidDepartmentID, http.StatusBadRequest)
		return
	}

	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	if err := a.sesc.UpdateDepartment(ctx, id, req.Name, req.Description); err != nil {
		if errors.Is(err, sesc.ErrInvalidDepartment) {
			a.writeError(w, ErrDepartmentExists, http.StatusConflict)
			return
		}
		a.writeError(w, APIError{
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

func (a *API) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ids := r.PathValue("id")

	var id uuid.UUID
	if err := (&id).Parse(ids); err != nil {
		a.writeError(w, ErrInvalidDepartmentID, http.StatusBadRequest)
		return
	}

	if err := a.sesc.DeleteDepartment(ctx, id); err != nil {
		if errors.Is(err, sesc.ErrInvalidDepartment) {
			a.writeError(w, ErrInvalidDepartment, http.StatusNotFound)
			return
		} else if errors.Is(err, sesc.ErrCannotRemoveDepartment) {
			a.writeError(w, ErrCannotRemoveDepartment, http.StatusConflict)
			return
		}
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to delete department",
			RuMessage: "Ошибка удаления кафедры",
		}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
