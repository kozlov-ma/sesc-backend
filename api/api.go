package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	_ "github.com/kozlov-ma/sesc-backend/api/docs" // This blank import is needed to serve the swagger scheme.
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title SESC Management API
// @version 1.0
// @description API for managing SESC departments, users and permissions
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter 'Bearer ' followed by your token

type API struct {
	sesc      SESC
	iam       IAMService
	eventSink EventSink
}

func New(sesc SESC, iam IAMService, eventSink EventSink) *API {
	return &API{sesc: sesc, iam: iam, eventSink: eventSink}
}

// Helper functions
func (a *API) writeJSON(ctx context.Context, w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		rec := event.Get(ctx)
		rec.Add(events.Error, fmt.Errorf("couldn't write json: %w", err))
	}
}

func writeError[T SpecificError](ctx context.Context, w http.ResponseWriter, apiError T) {
	rec := event.Get(ctx)

	genericError := ToError(apiError)

	// Set default status code if not provided
	statusCode := http.StatusInternalServerError
	if genericError.StatusCode != 0 {
		statusCode = genericError.StatusCode
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	rec.Sub("http").Add("error_response", genericError)

	err := json.NewEncoder(w).Encode(genericError)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("couldn't write json: %w", err))
	}
}

func convertDepartment(d sesc.Department) Department {
	return Department{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
	}
}

const allowAllOriginsNow = true

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")
		if allowAllOriginsNow || isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
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
	r.Use(a.EventMiddleware)

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
			r.With(a.CurrentUserMiddleware).Get("/me", a.GetCurrentUser)
			r.Get("/", a.GetUsers)
			r.Get("/{id}", a.GetUser)
		})
	})

	// Admin-only routes
	r.Group(func(r chi.Router) {
		r.Use(a.RequireAuthMiddleware)
		r.Use(a.RoleMiddleware("admin"))

		r.Post("/dev/fakedata", a.FakeData)

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
