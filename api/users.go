package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
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
}

type CreateUserRequest struct {
	FirstName    string    `json:"firstName"             example:"Anna"                                 validate:"required"`
	LastName     string    `json:"lastName"              example:"Smirnova"                             validate:"required"`
	MiddleName   string    `json:"middleName"            example:"Olegovna"`
	RoleID       int32     `json:"roleId"                example:"2"                                    validate:"required"`
	PictureURL   string    `json:"pictureUrl,omitzero"   example:"/images/users/ivan.jpg"`
	DepartmentID uuid.UUID `json:"departmentId,omitzero" example:"550e8400-e29b-41d4-a716-446655440000"`
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
	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(a, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}, http.StatusBadRequest)
		return
	}

	user, err := a.sesc.User(ctx, userID)
	switch {
	case errors.Is(err, sesc.ErrUserNotFound):
		writeError(a, w, UserNotFoundError{
			Code:      "USER_NOT_FOUND",
			Message:   "User not found",
			RuMessage: "Пользователь не найден",
		}, http.StatusNotFound)
		return
	case err != nil:
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch user",
			RuMessage: "Ошибка получения данных пользователя",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		PictureURL: user.PictureURL,
		Role:       convertRole(user.Role),
		Department: convertDepartment(user.Department),
		Suspended:  user.Suspended,
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
	users, err := a.sesc.Users(ctx)
	if err != nil {
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch users",
			RuMessage: "Ошибка получения данных пользователей",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UsersResponse{
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
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(a, w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	user, err := a.sesc.CreateUser(ctx, sesc.UserUpdateOptions{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		MiddleName:   req.MiddleName,
		PictureURL:   req.PictureURL,
		DepartmentID: req.DepartmentID,
		NewRoleID:    req.RoleID,
	})

	switch {
	case errors.Is(err, sesc.ErrInvalidRole):
		writeError(a, w, InvalidRoleError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}, http.StatusBadRequest)
		return
	case errors.Is(err, sesc.ErrInvalidName):
		writeError(a, w, InvalidNameError{
			Code:      "INVALID_NAME",
			Message:   "Invalid name specified",
			RuMessage: "Указано некорректное имя",
		}, http.StatusBadRequest)
		return
	case err != nil:
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to create user",
			RuMessage: "Ошибка создания пользователя",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		PictureURL: user.PictureURL,
		Role:       convertRole(user.Role),
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
	RoleID       *int32     `json:"roleId,omitzero"       example:"1"                                    validate:"required"`
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

	idStr := r.PathValue("id")
	userID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(a, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}, http.StatusBadRequest)
		return
	}

	var req PatchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(a, w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	existing, err := a.sesc.User(ctx, userID)
	if errors.Is(err, sesc.ErrUserNotFound) {
		writeError(a, w, UserNotFoundError{
			Code:      "USER_NOT_FOUND",
			Message:   "User not found",
			RuMessage: "Пользователь не найден",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch user",
			RuMessage: "Ошибка получения данных пользователя",
		}, http.StatusInternalServerError)
		return
	}

	upd := existing.UpdateOptions()
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
		newRoleIsBad := (req.RoleID != nil && *req.RoleID != sesc.Teacher.ID && *req.RoleID != sesc.Dephead.ID)
		noNewRoleAndOldIsBad := (req.RoleID == nil && existing.Role.ID != sesc.Teacher.ID && existing.Role.ID != sesc.Dephead.ID)
		if newRoleIsBad || noNewRoleAndOldIsBad {
			writeError(a, w, InvalidRoleError{
				Code:      "INVALID_ROLE",
				Message:   "Unable to assign department to selected role",
				RuMessage: "Нельзя указать департамент для выбранной роли",
			}, http.StatusBadRequest)
			return
		}

		upd.DepartmentID = *req.DepartmentID
	}
	if req.RoleID != nil {
		upd.NewRoleID = *req.RoleID
	}

	updated, err := a.sesc.UpdateUser(ctx, userID, upd)
	switch {
	case errors.Is(err, sesc.ErrUserNotFound):
		writeError(a, w, UserNotFoundError{
			Code:      "USER_NOT_FOUND",
			Message:   "User not found",
			RuMessage: "Пользователь не найден",
		}, http.StatusNotFound)
		return

	case errors.Is(err, sesc.ErrInvalidRole):
		writeError(a, w, InvalidRoleError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}, http.StatusBadRequest)
		return

	case errors.Is(err, sesc.ErrInvalidName):
		writeError(a, w, InvalidNameError{
			Code:      "INVALID_NAME",
			Message:   "Invalid first or last name",
			RuMessage: "Указано некорректное имя",
		}, http.StatusBadRequest)
		return

	case err != nil:
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to update user",
			RuMessage: "Ошибка обновления пользователя",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UserResponse{
		ID:         updated.ID,
		FirstName:  updated.FirstName,
		LastName:   updated.LastName,
		MiddleName: updated.MiddleName,
		PictureURL: updated.PictureURL,
		Role:       convertRole(updated.Role),
		Department: convertDepartment(updated.Department),
		Suspended:  updated.Suspended,
	}, http.StatusOK)
}

func convertUser(user sesc.User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		PictureURL: user.PictureURL,
		Role:       convertRole(user.Role),
		Department: convertDepartment(user.Department),
		Suspended:  user.Suspended,
	}
}

func convertUsers(users []sesc.User) []UserResponse {
	convertedUsers := make([]UserResponse, len(users))
	for _, user := range users {
		convertedUsers = append(convertedUsers, convertUser(user))
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

	// Get identity from context
	identity, ok := GetIdentityFromContext(ctx)
	if !ok {
		writeError(a, w, UnauthorizedError{
			Code:      "UNAUTHORIZED",
			Message:   "Authentication required",
			RuMessage: "Требуется авторизация",
		}, http.StatusUnauthorized)
		return
	}

	// Get user from sesc
	user, err := a.sesc.User(ctx, identity.ID)
	switch {
	case errors.Is(err, sesc.ErrUserNotFound):
		writeError(a, w, UserNotFoundError{
			Code:      "USER_NOT_FOUND",
			Message:   "User not found",
			RuMessage: "Пользователь не найден",
		}, http.StatusNotFound)
		return
	case err != nil:
		writeError(a, w, ServerError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch user data",
			RuMessage: "Ошибка получения данных пользователя",
		}, http.StatusInternalServerError)
		return
	}

	// Return user data
	a.writeJSON(w, UserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		PictureURL: user.PictureURL,
		Role:       convertRole(user.Role),
		Department: convertDepartment(user.Department),
		Suspended:  user.Suspended,
	}, http.StatusOK)
}
