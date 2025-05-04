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
	FirstName  string `json:"firstName"  example:"Anna"                   validate:"required"`
	LastName   string `json:"lastName"   example:"Smirnova"               validate:"required"`
	MiddleName string `json:"middleName" example:"Olegovna"`
	PictureURL string `json:"pictureUrl" example:"/images/users/anna.jpg"`
	RoleID     int32  `json:"roleId"     example:"2"                      validate:"required"`
}

// GetUser godoc
// @Summary Get user details
// @Description Retrieves detailed information about a user
// @Tags users
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Invalid UUID format"
// @Failure 404 {object} APIError "User not found"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users/{id} [get]
func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}, http.StatusBadRequest)
		return
	}

	user, err := a.sesc.User(ctx, userID)
	if errors.Is(err, sesc.ErrUserNotFound) {
		a.writeError(w, APIError{
			Code:      "USER_NOT_FOUND",
			Message:   "User not found",
			RuMessage: "Пользователь не найден",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		a.writeError(w, APIError{
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

// CreateUser godoc
// @Summary Create new user
// @Description Creates a new user with specified role (non-teacher)
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User details"
// @Success 201 {object} UserResponse
// @Failure 400 {object} APIError "Invalid role or request format"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users [post]
func (a *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	role, ok := sesc.RoleByID(req.RoleID)
	if !ok {
		a.writeError(w, APIError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}, http.StatusBadRequest)
		return
	}

	user, err := a.sesc.CreateUser(ctx, sesc.UserOptions{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		MiddleName: req.MiddleName,
	}, role)

	if errors.Is(err, sesc.ErrInvalidName) {
		a.writeError(w, APIError{
			Code:      "INVALID_NAME",
			Message:   "Invalid name specified",
			RuMessage: "Указано некорректное имя",
		}, http.StatusBadRequest)
		return
	}

	if err != nil {
		a.writeError(w, APIError{
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
// DepartmentID is only allowed to be set if the user’s role is Teacher or Dephead.
type PatchUserRequest struct {
	FirstName    *string    `json:"firstName"             example:"Ivan"                                 validate:"required"`
	LastName     *string    `json:"lastName"              example:"Petrov"                               validate:"required"`
	MiddleName   *string    `json:"middleName,omitzero"   example:"Sergeevich"`
	PictureURL   *string    `json:"pictureUrl,omitzero"   example:"/images/users/ivan.jpg"`
	Suspended    *bool      `json:"suspended,omitzero"    example:"false"                                validate:"required"`
	DepartmentID *uuid.UUID `json:"departmentId,omitzero" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleId       *int32     `json:"roleId,omitzero"       example:"1"                                    validate:"required"`
}

// PatchUser godoc
// @Summary Partially update user
// @Description Applies a partial update to the user identified by {id}. Only non-nil fields in the request are applied.
// Department can only be set for Teacher or Department-Head roles.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body PatchUserRequest true "User fields to update"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Invalid request format or invalid field value"
// @Failure 404 {object} APIError "User not found"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users/{id} [patch]
func (a *API) PatchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	userID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}, http.StatusBadRequest)
		return
	}

	var req PatchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	existing, err := a.sesc.User(ctx, userID)
	if errors.Is(err, sesc.ErrUserNotFound) {
		a.writeError(w, APIError{
			Code:      "USER_NOT_FOUND",
			Message:   "User not found",
			RuMessage: "Пользователь не найден",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		a.writeError(w, APIError{
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
		newRoleIsBad := (req.RoleId != nil && *req.RoleId != sesc.Teacher.ID && *req.RoleId != sesc.Dephead.ID)
		noNewRoleAndOldIsBad := (req.RoleId == nil && existing.Role.ID != sesc.Teacher.ID && existing.Role.ID != sesc.Dephead.ID)
		if newRoleIsBad || noNewRoleAndOldIsBad {
			a.writeError(w, APIError{
				Code:      "INVALID_ROLE",
				Message:   "Unable to assign department to selected role",
				RuMessage: "Нельзя указать департамент для выбранной роли",
			}, http.StatusBadRequest)
			return
		}

		upd.DepartmentID = *req.DepartmentID
	}
	if req.RoleId != nil {
		upd.NewRoleID = *req.RoleId
	}

	updated, err := a.sesc.UpdateUser(ctx, userID, upd)
	switch {
	case errors.Is(err, sesc.ErrUserNotFound):
		a.writeError(w, APIError{
			Code:      "USER_NOT_FOUND",
			Message:   "User not found",
			RuMessage: "Пользователь не найден",
		}, http.StatusNotFound)
		return

	case errors.Is(err, sesc.ErrInvalidRole):
		a.writeError(w, APIError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}, http.StatusBadRequest)
		return

	case errors.Is(err, sesc.ErrInvalidName):
		a.writeError(w, APIError{
			Code:      "INVALID_NAME",
			Message:   "Invalid first or last name",
			RuMessage: "Указано некорректное имя",
		}, http.StatusBadRequest)
		return

	case err != nil:
		a.writeError(w, APIError{
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
