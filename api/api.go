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

var (
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
		a.writeError(w, APIError{
			Code:      "INVALID_REQUEST",
			Message:   "Invalid request body",
			RuMessage: "Некорректный формат запроса",
		}, http.StatusBadRequest)
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

// GrantPermissions godoc
// @Summary Grant permissions to user
// @Description Adds extra permissions to a user's account
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body GrantPermissionsRequest true "List of permission IDs"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Invalid UUID or permission ID"
// @Failure 404 {object} APIError "User not found"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users/{id}/permissions [post]
func (a *API) GrantPermissions(w http.ResponseWriter, r *http.Request) {
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

	var req GrantPermissionsRequest
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

	updatedUser, err := a.sesc.GrantExtraPermissions(ctx, user, permissions...)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to grant permissions",
			RuMessage: "Ошибка назначения разрешений",
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

func convertRole(r sesc.Role) Role {
	return Role{
		ID:          r.ID,
		Name:        r.Name,
		Permissions: convertPermissions(r.Permissions),
	}
}

func convertPermissions(perms []sesc.Permission) []Permission {
	res := make([]Permission, len(perms))
	for i, p := range perms {
		res[i] = Permission{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}
	return res
}

func convertDepartment(d sesc.Department) Department {
	return Department{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
	}
}

// Roles godoc
// @Summary List all roles
// @Description Retrieves all system roles with their permissions
// @Tags roles
// @Produce json
// @Success 200 {object} RolesResponse
// @Failure 500 {object} APIError "Internal server error"
// @Router /roles [get]
func (a *API) Roles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roles, err := a.sesc.Roles(ctx)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch roles",
			RuMessage: "Ошибка получения списка ролей",
		}, http.StatusInternalServerError)
		return
	}

	response := RolesResponse{
		Roles: make([]Role, len(roles)),
	}
	for i, role := range roles {
		response.Roles[i] = convertRole(role)
	}

	a.writeJSON(w, response, http.StatusOK)
}

// Permissions godoc
// @Summary List all permissions
// @Description Retrieves all available system permissions
// @Tags permissions
// @Produce json
// @Success 200 {object} PermissionsResponse
// @Router /permissions [get]
func (a *API) Permissions(w http.ResponseWriter, r *http.Request) {
	perms := sesc.Permissions
	response := PermissionsResponse{
		Permissions: make([]Permission, len(perms)),
	}

	for i, p := range perms {
		response.Permissions[i] = Permission{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	a.writeJSON(w, response, http.StatusOK)
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
		a.writeError(w, APIError{
			Code:      "INVALID_REQUEST",
			Message:   "Invalid request body",
			RuMessage: "Некорректный формат запроса",
		}, http.StatusBadRequest)
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

	var selectedRole sesc.Role
	for _, r := range roles {
		if r.ID == req.RoleID && r.ID != sesc.Teacher.ID {
			selectedRole = r
			break
		}
	}

	if selectedRole.ID == 0 {
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
		PictureURL: req.PictureURL,
	}, selectedRole)

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

// PatchUser applies a partial update to the User record identified by userID.
// Only the fields provided (non-nil in the request) are updated.
//
// If DepartmentID is provided then it is verified that the user’s role is either Teacher
// or Dephead. Otherwise, PatchUser returns ErrInvalidDepartment.
//
// If the user does not exist, PatchUser returns ErrUserNotFound.

// PatchUser godoc
// @Summary Patch user
// @Description A partial update to the user identified by userID
// @Tags users
// @Accept json
// @Produce json
// @Param userID path string true "User UUID"
// @Param request body PatchUserRequest true "User fields to update"
// @Success 200 {object} UserResponse
// @Failure 400 {object} APIError "Invalid department"
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
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to fetch user",
			RuMessage: "Ошибка получения данных пользователя",
		}, http.StatusInternalServerError)
		return
	}

	if req.DepartmentID != nil {
		dep, err := a.sesc.DepartmentByID(ctx, *req.DepartmentID)
		if err != nil {
			a.writeError(w, ErrInvalidDepartment, http.StatusBadRequest)
			return
		}
		updated, err := a.sesc.SetDepartment(ctx, user, dep)
		if errors.Is(err, sesc.ErrInvalidDepartment) {
			a.writeError(w, ErrInvalidDepartment, http.StatusBadRequest)
			return
		}
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
				Message:   "Failed to update department",
				RuMessage: "Ошибка установки кафедры",
			}, http.StatusInternalServerError)
			return
		}
		user = updated
	}

	patched := user // copy
	if req.FirstName != nil {
		patched.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		patched.LastName = *req.LastName
	}
	if req.MiddleName != nil {
		patched.MiddleName = *req.MiddleName
	}
	if req.PictureURL != nil {
		patched.PictureURL = *req.PictureURL
	}
	if req.Suspended != nil {
		patched.Suspended = *req.Suspended
	}

	updatedUser, err := a.sesc.UpdateUser(ctx, patched)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to update user",
			RuMessage: "Ошибка обновления пользователя",
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

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /departments", a.CreateDepartment)
	mux.HandleFunc("GET /departments", a.Departments)
	mux.HandleFunc("GET /roles", a.Roles)
	mux.HandleFunc("GET /permissions", a.Permissions)
	mux.HandleFunc("POST /users", a.CreateUser)
	mux.HandleFunc("PATCH /users/{id}", a.PatchUser)
	mux.HandleFunc("GET /users/{id}", a.GetUser)
	mux.HandleFunc("POST /users/{id}/permissions", a.GrantPermissions)
	mux.HandleFunc("DELETE /users/{id}/permissions", a.RevokePermissions)

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}
