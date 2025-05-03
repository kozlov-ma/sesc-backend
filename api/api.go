package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/kozlov-ma/sesc-backend/api/docs"
)

// @title SESC Management API
// @version 1.0
// @description API for managing SESC departments, users and permissions

type API struct {
	log  *slog.Logger
	sesc SESC
}

func New(log *slog.Logger, sesc SESC) *API {
	return &API{log: log, sesc: sesc}
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
	}, http.StatusOK)
}

// Helper functions
func (a *API) writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		a.log.Error("couldn't write json", "error", err)
	}
}

func (a *API) writeError(w http.ResponseWriter, apiError APIError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(struct {
		Error APIError `json:"error"`
	}{Error: apiError})

	if err != nil {
		a.log.Error("couldn't write json", "error", err)
	}
}

func convertDepartment(d sesc.Department) Department {
	return Department{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
	}
}

// RevokePermissions godoc
// @Summary Revoke permissions
// @Description Removes extra permissions from a user
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body RevokePermissionsRequest true "Permission IDs to revoke"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Invalid UUID or permission ID"
// @Failure 404 {object} APIError "User not found"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users/{id}/permissions [delete]
func (a *API) RevokePermissions(w http.ResponseWriter, r *http.Request) {
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

	var req RevokePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_REQUEST",
			Message:   "Invalid request body",
			RuMessage: "Некорректный формат запроса",
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

	permissions := make([]sesc.Permission, 0, len(req.Permissions))
	for _, pid := range req.Permissions {
		p, exists := sesc.PermissionByID(pid)
		if !exists {
			a.writeError(w, APIError{
				Code:      "INVALID_PERMISSION",
				Message:   "Invalid permission ID",
				RuMessage: "Некорректный ID разрешения",
			}, http.StatusBadRequest)
			return
		}
		permissions = append(permissions, p)
	}

	updatedUser, err := a.sesc.RevokeExtraPermissions(ctx, user, permissions...)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to revoke permissions",
			RuMessage: "Ошибка отзыва разрешений",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UserResponse{
		ID:         updatedUser.ID,
		FirstName:  updatedUser.FirstName,
		LastName:   updatedUser.LastName,
		MiddleName: updatedUser.MiddleName,
		PictureURL: updatedUser.PictureURL,
		Role:       convertRole(updatedUser.Role),
		Department: convertDepartment(updatedUser.Department),
	}, http.StatusOK)
}

// SetRole godoc
// @Summary Update user role
// @Description Changes the user's system role
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body SetRoleRequest true "New role ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Invalid role or UUID"
// @Failure 404 {object} APIError "User not found"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users/{id}/role [put]
func (a *API) SetRole(w http.ResponseWriter, r *http.Request) {
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

	var req SetRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_REQUEST",
			Message:   "Invalid request body",
			RuMessage: "Некорректный формат запроса",
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

	roles, err := a.sesc.Roles(ctx)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch roles",
			RuMessage: "Ошибка получения списка ролей",
		}, http.StatusInternalServerError)
		return
	}

	var newRole sesc.Role
	for _, r := range roles {
		if r.ID == req.RoleID {
			newRole = r
			break
		}
	}

	if newRole.ID == 0 {
		a.writeError(w, APIError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID",
			RuMessage: "Некорректный ID роли",
		}, http.StatusBadRequest)
		return
	}

	updatedUser, err := a.sesc.SetRole(ctx, user, newRole)
	if errors.Is(err, sesc.ErrInvalidRoleChange) {
		a.writeError(w, APIError{
			Code:      "INVALID_ROLE_CHANGE",
			Message:   "Cannot assign teacher role without department",
			RuMessage: "Невозможно назначить роль преподавателя без кафедры",
		}, http.StatusBadRequest)
		return
	}
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to update role",
			RuMessage: "Ошибка обновления роли",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UserResponse{
		ID:         updatedUser.ID,
		FirstName:  updatedUser.FirstName,
		LastName:   updatedUser.LastName,
		MiddleName: updatedUser.MiddleName,
		PictureURL: updatedUser.PictureURL,
		Role:       convertRole(updatedUser.Role),
		Department: convertDepartment(updatedUser.Department),
	}, http.StatusOK)
}

// SetDepartment godoc
// @Summary Update department
// @Description Updates the user's department assignment
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body SetDepartmentRequest true "Department ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Invalid department or UUID"
// @Failure 404 {object} APIError "User/department not found"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users/{id}/department [put]
func (a *API) SetDepartment(w http.ResponseWriter, r *http.Request) {
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

	var req SetDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_REQUEST",
			Message:   "Invalid request body",
			RuMessage: "Некорректный формат запроса",
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

	department, err := a.sesc.DepartmentByID(ctx, req.DepartmentID)
	if err != nil || department.ID == uuid.Nil {
		a.writeError(w, APIError{
			Code:      "DEPARTMENT_NOT_FOUND",
			Message:   "Department not found",
			RuMessage: "Кафедра не найдена",
		}, http.StatusNotFound)
		return
	}

	updatedUser, err := a.sesc.SetDepartment(ctx, user, department)
	if errors.Is(err, sesc.ErrInvalidDepartment) {
		a.writeError(w, APIError{
			Code:      "INVALID_DEPARTMENT",
			Message:   "Invalid department for user role",
			RuMessage: "Некорректная кафедра для текущей роли пользователя",
		}, http.StatusBadRequest)
		return
	}
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to update department",
			RuMessage: "Ошибка обновления кафедры",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, UserResponse{
		ID:         updatedUser.ID,
		FirstName:  updatedUser.FirstName,
		LastName:   updatedUser.LastName,
		MiddleName: updatedUser.MiddleName,
		PictureURL: updatedUser.PictureURL,
		Role:       convertRole(updatedUser.Role),
		Department: convertDepartment(department),
	}, http.StatusOK)
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /departments", a.CreateDepartment)
	mux.HandleFunc("GET /departments", a.Departments)
	mux.HandleFunc("GET /roles", a.Roles)
	mux.HandleFunc("GET /permissions", a.Permissions)
	mux.HandleFunc("POST /teachers", a.CreateTeacher)
	mux.HandleFunc("POST /users", a.CreateUser)
	mux.HandleFunc("GET /users/{id}", a.GetUser)
	mux.HandleFunc("PUT /users/{id}/role", a.SetRole)
	mux.HandleFunc("PUT /users/{id}/department", a.SetDepartment)
	mux.HandleFunc("POST /users/{id}/permissions", a.GrantPermissions)
	mux.HandleFunc("DELETE /users/{id}/permissions", a.RevokePermissions)

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}
