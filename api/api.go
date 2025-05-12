package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/kozlov-ma/sesc-backend/api/docs"
)

// @title SESC Management API
// @version 1.0
// @description API for managing SESC departments, users and permissions
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter 'Bearer ' followed by your token

type API struct {
	log  *slog.Logger
	sesc SESC
	iam  IAMService
}

func New(log *slog.Logger, sesc SESC, iam IAMService) *API {
	return &API{log: log, sesc: sesc, iam: iam}
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")
		if isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
			w.Header().Set("Access-Control-Allow-Headers", requestedHeaders)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	hostname := u.Hostname()
	return hostname == "localhost"
}

func (a *API) RegisterRoutes(r chi.Router) {
	// Apply global middlewares
	r.Use(corsMiddleware)
	r.Use(a.AuthMiddleware)

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Auth endpoints
		r.Post("/auth/login", a.Login)
		r.Post("/auth/admin/login", a.LoginAdmin)

		// Public endpoints
		r.Get("/departments", a.Departments)
		r.Get("/roles", a.Roles)
		r.Get("/permissions", a.Permissions)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		r.Use(a.RequireAuthMiddleware)

		// Token validation
		r.Get("/auth/validate", a.ValidateToken)

		// User routes with current user context
		r.Route("/users", func(r chi.Router) {
			// Add current user to context for user routes
			r.Use(a.CurrentUserMiddleware)

			r.Get("/me", a.GetCurrentUser)
			r.Get("/", a.GetUsers)
			r.Get("/{id}", a.GetUser)
		})
	})

	// Admin-only routes
	r.Group(func(r chi.Router) {
		r.Use(a.RequireAuthMiddleware)
		r.Use(a.RoleMiddleware("admin"))

		// Setting credentials for a user
		r.Put("/users/{id}/credentials", a.RegisterUser)

		// Department management
		r.Post("/departments", a.CreateDepartment)
		r.Put("/departments/{id}", a.UpdateDepartment)
		r.Delete("/departments/{id}", a.DeleteDepartment)

		// User management
		r.Post("/users", a.CreateUser)
		r.Patch("/users/{id}", a.PatchUser)

		// Credential management
		r.Delete("/auth/credentials/{id}", a.DeleteCredentials)
		r.Get("/auth/credentials/{id}", a.GetCredentials)
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)
}
