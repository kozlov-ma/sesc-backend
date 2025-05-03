package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type UserResponse struct {
	ID         uuid.UUID  `json:"id"                  example:"550e8400-e29b-41d4-a716-446655440000"`
	FirstName  string     `json:"firstName"           example:"Ivan"`
	LastName   string     `json:"lastName"            example:"Petrov"`
	MiddleName string     `json:"middleName"          example:"Sergeevich"`
	PictureURL string     `json:"pictureUrl"          example:"/images/users/ivan.jpg"`
	Role       Role       `json:"role"`
	Department Department `json:"department,omitzero"`
}

type CreateUserRequest struct {
	FirstName  string `json:"firstName"  example:"Anna"`
	LastName   string `json:"lastName"   example:"Smirnova"`
	MiddleName string `json:"middleName" example:"Olegovna"`
	PictureURL string `json:"pictureUrl" example:"/images/users/anna.jpg"`
	RoleID     int32  `json:"roleId"     example:"2"`
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
