package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type UserResponse struct {
	ID         uuid.UUID  `json:"id"                  example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	FirstName  string     `json:"firstName"           example:"Ivan"                                 validate:"required"`
	LastName   string     `json:"lastName"            example:"Petrov"                               validate:"required"`
	MiddleName string     `json:"middleName"          example:"Sergeevich"`
	PictureURL string     `json:"pictureUrl"          example:"/images/users/ivan.jpg"               validate:"required"`
	Role       Role       `json:"role"                                                               validate:"required"`
	Suspended  bool       `json:"suspended"                                                          validate:"required"`
	Department Department `json:"department,omitzero"`

	Subdivision       string  `json:"subdivision"       example:"Кафедра информатики"       validate:"required"`
	JobTitle          string  `json:"jobTitle"          example:"Профессор"                 validate:"required"`
	EmploymentRate    float64 `json:"employmentRate"    example:"1.0"                       validate:"required"`
	AcademicDegree    int     `json:"academicDegree"    example:"2"`
	PersonnelCategory int     `json:"personnelCategory" example:"1"                         validate:"required"`
	EmploymentType    int     `json:"employmentType"    example:"1"                         validate:"required"`
	AcademicTitle     string  `json:"academicTitle"     example:"Профессор"`
	Honors            string  `json:"honors"            example:"Заслуженный деятель науки"`
	Category          string  `json:"category"          example:"Высшая"`

	DateOfEmployment time.Time `json:"dateOfEmployment,omitzero" example:"2020-01-15T00:00:00Z"`
	UnemploymentDate time.Time `json:"unemploymentDate,omitzero" example:"2023-12-31T00:00:00Z"`
}

type CreateUserRequest struct {
	FirstName    string    `json:"firstName"             example:"Anna"                                 validate:"required"`
	LastName     string    `json:"lastName"              example:"Smirnova"                             validate:"required"`
	MiddleName   string    `json:"middleName"            example:"Olegovna"`
	Role         int       `json:"role"                  example:"2"                                    validate:"required"`
	PictureURL   string    `json:"pictureUrl,omitzero"   example:"/images/users/ivan.jpg"`
	DepartmentID uuid.UUID `json:"departmentId,omitzero" example:"550e8400-e29b-41d4-a716-446655440000"`

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
// @Success 200 {object} UserResponse
// @Failure 400 {object} InvalidUUIDError "Invalid UUID format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 404 {object} UserNotFoundError "User not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /users/{id} [get]
func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(ctx, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	user, err := a.sesc.User(ctx, userID)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, sescError(err))
		return
	}

	a.writeJSON(ctx, w, UserResponse{
		ID:                user.ID,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		MiddleName:        user.MiddleName,
		PictureURL:        user.PictureURL,
		Role:              convertRole(user.Role),
		Department:        convertDepartment(user.Department),
		Suspended:         user.Suspended,
		Subdivision:       user.Subdivision,
		JobTitle:          user.JobTitle,
		EmploymentRate:    user.EmploymentRate,
		AcademicDegree:    int(user.AcademicDegree),
		PersonnelCategory: int(user.PersonnelCategory),
		EmploymentType:    int(user.EmploymentType),
		AcademicTitle:     user.AcademicTitle,
		Honors:            user.Honors,
		Category:          user.Category,
		DateOfEmployment:  user.DateOfEmployment,
		UnemploymentDate:  user.UnemploymentDate,
	}, http.StatusOK)
}

type UsersResponse struct {
	Users []UserResponse `json:"users" validate:"required"`
}

// GetUsers godoc
// @Summary Get all users registered in the system
// @Description Retrieves detailed information about all users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} UsersResponse
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /users [get]
func (a *API) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)
	users, err := a.sesc.Users(ctx)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch users",
			RuMessage: "Ошибка получения данных пользователей",
		}.WithStatus(http.StatusInternalServerError))
		return
	}

	a.writeJSON(ctx, w, UsersResponse{
		Users: convertUsers(users),
	}, http.StatusOK)
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
// @Success 201 {object} UserResponse
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 400 {object} InvalidRoleError "Invalid role ID specified"
// @Failure 400 {object} InvalidNameError "Invalid name specified"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /users [post]
func (a *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest).WithDetails(err.Error()))
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
		writeError(ctx, w, sescError(err))
		return
	}

	a.writeJSON(ctx, w, UserResponse{
		ID:                user.ID,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		MiddleName:        user.MiddleName,
		PictureURL:        user.PictureURL,
		Role:              convertRole(user.Role),
		Department:        convertDepartment(user.Department),
		Suspended:         user.Suspended,
		Subdivision:       user.Subdivision,
		JobTitle:          user.JobTitle,
		EmploymentRate:    user.EmploymentRate,
		AcademicDegree:    int(user.AcademicDegree),
		PersonnelCategory: int(user.PersonnelCategory),
		EmploymentType:    int(user.EmploymentType),
		AcademicTitle:     user.AcademicTitle,
		Honors:            user.Honors,
		Category:          user.Category,
		DateOfEmployment:  user.DateOfEmployment,
		UnemploymentDate:  user.UnemploymentDate,
	}, http.StatusCreated)
}

// PatchUserRequest defines the fields that can be updated on a User.
// Fields are pointers so that only non‑nil values are applied to the user record.
// DepartmentID is only allowed to be set if the user's role is Teacher or Dephead.
type PatchUserRequest struct {
	FirstName    *string    `json:"firstName"             example:"Ivan"                                 validate:"required"`
	LastName     *string    `json:"lastName"              example:"Petrov"                               validate:"required"`
	MiddleName   *string    `json:"middleName,omitzero"   example:"Sergeevich"`
	PictureURL   *string    `json:"pictureUrl,omitzero"   example:"/images/users/ivan.jpg"`
	Suspended    *bool      `json:"suspended,omitzero"    example:"false"                                validate:"required"`
	DepartmentID *uuid.UUID `json:"departmentId,omitzero" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleChange   *int       `json:"roleId,omitzero"       example:"1"                                    validate:"required"`

	Subdivision       *string  `json:"subdivision,omitzero"       example:"Кафедра информатики"`
	JobTitle          *string  `json:"jobTitle,omitzero"          example:"Профессор"`
	EmploymentRate    *float64 `json:"employmentRate,omitzero"    example:"1.0"`
	AcademicDegree    *int     `json:"academicDegree,omitzero"    example:"2"`
	PersonnelCategory *int     `json:"personnelCategory,omitzero" example:"1"`
	EmploymentType    *int     `json:"employmentType,omitzero"    example:"1"`
	AcademicTitle     *string  `json:"academicTitle,omitzero"     example:"Профессор"`
	Honors            *string  `json:"honors,omitzero"            example:"Заслуженный деятель науки"`
	Category          *string  `json:"category,omitzero"          example:"Высшая"`

	DateOfEmployment *time.Time `json:"dateOfEmployment,omitzero" example:"2020-01-15T00:00:00Z"`
	UnemploymentDate *time.Time `json:"unemploymentDate,omitzero" example:"2023-12-31T00:00:00Z"`
}

// validateDepartmentAssignment validates that department can be assigned to the role
func (a *API) validateDepartmentAssignment(req *PatchUserRequest, existing sesc.User) bool {
	if req.DepartmentID == nil {
		return true
	}

	newRoleIsBad := (req.RoleChange != nil && *req.RoleChange != int(sesc.Teacher) && *req.RoleChange != int(sesc.Dephead))
	noNewRoleAndOldIsBad := (req.RoleChange == nil && existing.Role != sesc.Teacher && existing.Role != sesc.Dephead)

	return !newRoleIsBad && !noNewRoleAndOldIsBad
}

// applyPatchUserRequest applies the patch request to the user update options
func (a *API) applyPatchUserRequest(req *PatchUserRequest, upd *sesc.UserUpdateOptions) {
	if req.FirstName != nil {
		upd.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		upd.LastName = *req.LastName
	}
	if req.MiddleName != nil {
		upd.MiddleName = *req.MiddleName
	}
	if req.PictureURL != nil {
		upd.PictureURL = *req.PictureURL
	}
	if req.Suspended != nil {
		upd.Suspended = *req.Suspended
	}
	if req.DepartmentID != nil {
		upd.DepartmentID = *req.DepartmentID
	}
	if req.RoleChange != nil {
		upd.NewRole = sesc.Role(*req.RoleChange)
	}
	if req.Subdivision != nil {
		upd.Subdivision = *req.Subdivision
	}
	if req.JobTitle != nil {
		upd.JobTitle = *req.JobTitle
	}
	if req.EmploymentRate != nil {
		upd.EmploymentRate = *req.EmploymentRate
	}
	if req.AcademicDegree != nil {
		upd.AcademicDegree = sesc.AcademicDegree(*req.AcademicDegree)
	}
	if req.PersonnelCategory != nil {
		upd.PersonnelCategory = sesc.PersonnelCategory(*req.PersonnelCategory)
	}
	if req.EmploymentType != nil {
		upd.EmploymentType = sesc.EmploymentType(*req.EmploymentType)
	}
	if req.AcademicTitle != nil {
		upd.AcademicTitle = *req.AcademicTitle
	}
	if req.Honors != nil {
		upd.Honors = *req.Honors
	}
	if req.Category != nil {
		upd.Category = *req.Category
	}
	if req.DateOfEmployment != nil {
		upd.DateOfEmployment = *req.DateOfEmployment
	}
	if req.UnemploymentDate != nil {
		upd.UnemploymentDate = *req.UnemploymentDate
	}
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
// @Success 200 {object} UserResponse
// @Failure 400 {object} InvalidUUIDError "Invalid UUID format"
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 400 {object} InvalidRoleError "Invalid role"
// @Failure 400 {object} InvalidNameError "Invalid name"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} UserNotFoundError "User not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /users/{id} [patch]
func (a *API) PatchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")
	userID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(ctx, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	var req PatchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	existing, err := a.sesc.User(ctx, userID)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, sescError(err))
		return
	}

	if !a.validateDepartmentAssignment(&req, existing) {
		writeError(ctx, w, InvalidRoleError{
			Code:      "INVALID_ROLE",
			Message:   "Unable to assign department to selected role",
			RuMessage: "Нельзя указать департамент для выбранной роли",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	upd := existing.UpdateOptions()
	a.applyPatchUserRequest(&req, &upd)

	updated, err := a.sesc.UpdateUser(ctx, userID, upd)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, sescError(err))
		return
	}

	a.writeJSON(ctx, w, convertUser(updated), http.StatusOK)
}

func convertUser(user sesc.User) UserResponse {
	return UserResponse{
		ID:                user.ID,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		MiddleName:        user.MiddleName,
		PictureURL:        user.PictureURL,
		Role:              convertRole(user.Role),
		Department:        convertDepartment(user.Department),
		Suspended:         user.Suspended,
		Subdivision:       user.Subdivision,
		JobTitle:          user.JobTitle,
		EmploymentRate:    user.EmploymentRate,
		AcademicDegree:    int(user.AcademicDegree),
		PersonnelCategory: int(user.PersonnelCategory),
		EmploymentType:    int(user.EmploymentType),
		AcademicTitle:     user.AcademicTitle,
		Honors:            user.Honors,
		Category:          user.Category,
		DateOfEmployment:  user.DateOfEmployment,
		UnemploymentDate:  user.UnemploymentDate,
	}
}

func convertUsers(users []sesc.User) []UserResponse {
	convertedUsers := make([]UserResponse, len(users))
	for i, user := range users {
		convertedUsers[i] = convertUser(user)
	}
	return convertedUsers
}

// GetCurrentUser godoc
// @Summary Get current user information
// @Description Returns information about the current authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} UserResponse
// @Failure 401 {object} UnauthorizedError "Unauthorized or invalid token"
// @Failure 404 {object} UserNotFoundError "User not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /users/me [get]
func (a *API) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := GetUserFromContext(ctx)

	// Return user data
	a.writeJSON(ctx, w, UserResponse{
		ID:                user.ID,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		MiddleName:        user.MiddleName,
		PictureURL:        user.PictureURL,
		Role:              convertRole(user.Role),
		Department:        convertDepartment(user.Department),
		Suspended:         user.Suspended,
		Subdivision:       user.Subdivision,
		JobTitle:          user.JobTitle,
		EmploymentRate:    user.EmploymentRate,
		AcademicDegree:    int(user.AcademicDegree),
		PersonnelCategory: int(user.PersonnelCategory),
		EmploymentType:    int(user.EmploymentType),
		AcademicTitle:     user.AcademicTitle,
		Honors:            user.Honors,
		Category:          user.Category,
		DateOfEmployment:  user.DateOfEmployment,
		UnemploymentDate:  user.UnemploymentDate,
	}, http.StatusOK)
}
