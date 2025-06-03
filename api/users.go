package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type CreateUserRequest struct {
	FirstName    string     `json:"firstName"             example:"Anna"                                 validate:"required"`
	LastName     string     `json:"lastName"              example:"Smirnova"                             validate:"required"`
	MiddleName   string     `json:"middleName"            example:"Olegovna"`
	Role         int        `json:"role"                  example:"2"                                    validate:"required"`
	PictureURL   string     `json:"pictureUrl,omitzero"   example:"/images/users/ivan.jpg"`
	DepartmentID *uuid.UUID `json:"departmentId,omitzero" example:"550e8400-e29b-41d4-a716-446655440000"`

	Subdivision       string  `json:"subdivision"            example:"Кафедра информатики"       validate:"required"`
	JobTitle          string  `json:"jobTitle"               example:"Профессор"                 validate:"required"`
	EmploymentRate    float64 `json:"employmentRate"         example:"1.0"                       validate:"required"`
	AcademicDegree    int     `json:"academicDegree"         example:"2"`
	PersonnelCategory int     `json:"personnelCategory"      example:"1"                         validate:"required"`
	EmploymentType    int     `json:"employmentType"         example:"1"                         validate:"required"`
	AcademicTitle     string  `json:"academicTitle,omitzero" example:"Профессор"`
	Honors            string  `json:"honors,omitzero"        example:"Заслуженный деятель науки"`
	Category          string  `json:"category,omitzero"      example:"Высшая"`

	DateOfEmployment time.Time `json:"dateOfEmployment,omitzero" example:"2020-01-15T00:00:00Z"`
	UnemploymentDate time.Time `json:"unemploymentDate,omitzero" example:"2023-12-31T00:00:00Z"`
}

// GetUser godoc
// @Summary Get user details
// @Description Retrieves detailed information about a user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User UUID"
// @Success 200 {object} respond.User
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "User not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users/{id} [get]
func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	user, err := a.sesc.User(ctx, userID)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithUser(user))
}

// GetUsers godoc
// @Summary Get all users registered in the system
// @Description Retrieves detailed information about all users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Param limit query int false "Pagination limit" default(10) minimum(1) maximum(100)
// @Param search query string false "Search by name"
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} respond.Users
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users [get]
func (a *API) GetUsers(w http.ResponseWriter, r *http.Request) {
	var (
		offset, limit int
		search        = r.PathValue("search")
	)

	offset, _ = strconv.Atoi(r.PathValue("offset"))
	limit, _ = strconv.Atoi(r.PathValue("limit"))
	if limit == 0 || limit >= 500 {
		limit = 100
	}

	ctx := r.Context()
	rec := event.Get(ctx)
	users, total, err := a.sesc.Users(ctx, offset, limit, search)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithUsers(users, total))
}

// CreateUser godoc
// @Summary Create new user
// @Description Creates a new user with specified role (non-teacher)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param request body CreateUserRequest true "User details"
// @Success 201 {object} respond.User
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 400 {object} respond.Error "Invalid role ID specified"
// @Failure 400 {object} respond.Error "Invalid name specified"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users [post]
func (a *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	user, err := a.sesc.CreateUser(ctx, sesc.UserUpdateOptions{
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		MiddleName:        req.MiddleName,
		PictureURL:        req.PictureURL,
		DepartmentID:      req.DepartmentID,
		NewRole:           sesc.Role(req.Role),
		Subdivision:       req.Subdivision,
		JobTitle:          req.JobTitle,
		EmploymentRate:    req.EmploymentRate,
		AcademicDegree:    sesc.AcademicDegree(req.AcademicDegree),
		PersonnelCategory: sesc.PersonnelCategory(req.PersonnelCategory),
		EmploymentType:    sesc.EmploymentType(req.EmploymentType),
		AcademicTitle:     req.AcademicTitle,
		Honors:            req.Honors,
		Category:          req.Category,
		DateOfEmployment:  req.DateOfEmployment,
		UnemploymentDate:  req.UnemploymentDate,
	})

	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithStatus(respond.WithUser(user), http.StatusCreated))
}

// PatchUserRequest defines the fields that can be updated on a User.
// Fields are pointers so that only non‑nil values are applied to the user record.
// DepartmentID is only allowed to be set if the user's role is Teacher or Dephead.
type PatchUserRequest struct {
	FirstName    string     `json:"firstName"             example:"Ivan"                                 validate:"required"`
	LastName     string     `json:"lastName"              example:"Petrov"                               validate:"required"`
	MiddleName   string     `json:"middleName,omitzero"   example:"Sergeevich"`
	PictureURL   string     `json:"pictureUrl,omitzero"   example:"/images/users/ivan.jpg"`
	Suspended    bool       `json:"suspended,omitzero"    example:"false"                                validate:"required"`
	DepartmentID *uuid.UUID `json:"departmentId,omitzero" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleID       int        `json:"roleId,omitzero"       example:"1"                                    validate:"required"`

	Subdivision       string  `json:"subdivision,omitzero"       example:"Кафедра информатики"`
	JobTitle          string  `json:"jobTitle,omitzero"          example:"Профессор"`
	EmploymentRate    float64 `json:"employmentRate,omitzero"    example:"1.0"`
	AcademicDegree    int     `json:"academicDegree,omitzero"    example:"2"`
	PersonnelCategory int     `json:"personnelCategory,omitzero" example:"1"`
	EmploymentType    int     `json:"employmentType,omitzero"    example:"1"`
	AcademicTitle     string  `json:"academicTitle,omitzero"     example:"Профессор"`
	Honors            string  `json:"honors,omitzero"            example:"Заслуженный деятель науки"`
	Category          string  `json:"category,omitzero"          example:"Высшая"`

	DateOfEmployment time.Time `json:"dateOfEmployment,omitzero" example:"2020-01-15T00:00:00Z"`
	UnemploymentDate time.Time `json:"unemploymentDate,omitzero" example:"2023-12-31T00:00:00Z"`
}

// PatchUser godoc
// @Summary Partially update user
// @Description Applies a partial update to the user identified by {id}. Only non-nil fields in the request are applied.
// Department can only be set for Teacher or Department-Head roles.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User UUID"
// @Param request body PatchUserRequest true "User fields to update"
// @Success 200 {object} respond.User
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 400 {object} respond.Error "Invalid role"
// @Failure 400 {object} respond.Error "Invalid name"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 404 {object} respond.Error "User not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users/{id} [patch]
func (a *API) PatchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")
	userID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	var req PatchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	opt := sesc.UserUpdateOptions{
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		MiddleName:        req.MiddleName,
		PictureURL:        req.PictureURL,
		Suspended:         req.Suspended,
		DepartmentID:      req.DepartmentID,
		NewRole:           sesc.Role(req.RoleID),
		Subdivision:       req.Subdivision,
		JobTitle:          req.JobTitle,
		EmploymentRate:    req.EmploymentRate,
		AcademicDegree:    sesc.AcademicDegree(req.AcademicDegree),
		PersonnelCategory: sesc.PersonnelCategory(req.PersonnelCategory),
		EmploymentType:    sesc.EmploymentType(req.EmploymentType),
		AcademicTitle:     req.AcademicTitle,
		Honors:            req.Honors,
		Category:          req.Category,
		DateOfEmployment:  req.DateOfEmployment,
		UnemploymentDate:  req.UnemploymentDate,
	}

	updated, err := a.sesc.UpdateUser(ctx, userID, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithUser(updated))
}

// GetCurrentUser godoc
// @Summary Get current user information
// @Description Returns information about the current authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} respond.User
// @Failure 401 {object} respond.Error "Unauthorized or invalid token"
// @Failure 404 {object} respond.Error "User not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users/me [get]
func (a *API) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := GetUserFromContext(ctx)

	// Return user data
	a.writeJSON(ctx, w, respond.WithUser(user))
}
