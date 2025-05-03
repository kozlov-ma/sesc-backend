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

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	// departments
	mux.HandleFunc("POST /departments", a.CreateDepartment)
	mux.HandleFunc("GET /departments", a.Departments)
	mux.HandleFunc("PUT /departments/{id}", a.UpdateDepartment)
	mux.HandleFunc("DELETE /departments/{id}", a.DeleteDepartment)

	// roles & permissions
	mux.HandleFunc("GET /roles", a.Roles)
	mux.HandleFunc("GET /permissions", a.Permissions)

	// users
	mux.HandleFunc("POST /users", a.CreateUser)
	mux.HandleFunc("PATCH /users/{id}", a.PatchUser)
	mux.HandleFunc("GET /users/{id}", a.GetUser)

	// swagger UI
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}
