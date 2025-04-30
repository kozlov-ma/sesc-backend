package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kozlov-ma/sesc-backend/sesc"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/kozlov-ma/sesc-backend/api/docs"
)

type API struct {
	log  *slog.Logger
	sesc SESC
}

func New(log *slog.Logger, sesc SESC) *API {
	return &API{
		log:  log,
		sesc: sesc,
	}
}

// CreateDepartment handles the creation of a new department.
// @Summary Create a new department
// @Description Creates a new department with the provided name and description
// @Tags departments
// @Accept  json
// @Produce  json
// @Param   department body CreateDepartmentRequest true "Department information"
// @Success 200 {object} CreateDepartmentResponse "Response containing created department or error details"
// @Failure 400 {string} string "Bad Request - Invalid input format"
// @Failure 409 {string} string "Conflict - Department with this name already exists"
// @Failure 500 {string} string "Internal Server Error - Failed to process request"
// @Router /api/departments [post]
func (a *API) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	a.log.DebugContext(ctx, "got a new CreateDepartment request")

	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "wrong serialization", http.StatusBadRequest)
		return
	}

	var cde CreateDepartmentError
	dep, err := a.sesc.CreateDepartment(ctx, req.Name, req.Description)
	if errors.Is(err, sesc.ErrInvalidDepartment) {
		cde = CreateDepartmentError{
			APIError: APIError{
				Status:    http.StatusConflict,
				Message:   "department already exists",
				RuMessage: "кафедра с таким названием уже существует",
			},
		}
	} else if err != nil {
		cde = CreateDepartmentError{
			APIError: InternalServerError,
		}
	}

	res := CreateDepartmentResponse{
		Error:       cde,
		ID:          dep.ID,
		Name:        dep.Name,
		Description: dep.Description,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		a.log.ErrorContext(ctx, "failed to encode response", "error", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Departments list all currently registered departments.
// @Summary List all departments
// @Description Retrieves a list of all departments
// @Tags departments
// @Produce  json
// @Success 200 {object} DepartmentsResponse "Response containing created departments or error details"
// @Failure 500 {string} string "Internal Server Error - Failed to process request"
// @Router /api/departments [get]
func (a *API) Departments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cde DepartmentsError
	deps, err := a.sesc.Departments(ctx)
	if err != nil {
		a.log.ErrorContext(ctx, "couldn't get departments because of server error", "error", err)
		cde = DepartmentsError{
			APIError: InternalServerError,
		}
	}

	departments := make([]Department, len(deps))
	for i, dep := range deps {
		departments[i] = Department{
			ID:          dep.ID,
			Name:        dep.Name,
			Description: dep.Description,
		}
	}

	res := DepartmentsResponse{
		Error:       cde,
		Departments: departments,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		a.log.ErrorContext(ctx, "failed to encode response", "error", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/departments", a.Departments)
	mux.HandleFunc("POST /api/departments", a.CreateDepartment)

	mux.HandleFunc("/api/swagger/", httpSwagger.WrapHandler)
}
