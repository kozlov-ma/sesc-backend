package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

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
