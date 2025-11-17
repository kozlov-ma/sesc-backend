package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"net/url"

	"github.com/go-chi/chi/v5"
	_ "github.com/kozlov-ma/sesc-backend/api/docs" // This blank import is needed to serve the swagger scheme.
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	httpSwagger "github.com/swaggo/http-swagger"
)

type API struct {
	sesc      SESC
	iam       AuthService
	file      FileService
	eventSink EventSink
}

func New(sesc SESC, iam AuthService, file FileService, eventSink EventSink) *API {
	return &API{sesc: sesc, iam: iam, file: file, eventSink: eventSink}
}

type statusCoder interface {
	AsHTTPStatusCode() int
}

// Helper functions
func (a *API) writeJSON(ctx context.Context, w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	if sc, ok := data.(statusCoder); ok {
		w.WriteHeader(sc.AsHTTPStatusCode())
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		rec := event.Get(ctx)
		rec.Add(events.Error, fmt.Errorf("couldn't write json: %w", err))
	}

	if e, ok := data.(error); ok {
		event.Get(ctx).Sub("http").Add("error_response", e)
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

// HealthCheck returns a simple OK response for health checking
func (a *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := respond.NewHealthCheckOK()
	a.writeJSON(r.Context(), w, response)
}

func (a *API) RegisterRoutes(r chi.Router) {
	r.Use(
		a.EventMiddleware,
		corsMiddleware,
	)

	r.Get("/up", a.HealthCheck)

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		r.Post("/auth/login", a.Login)

		r.Get("/departments/{id}", a.GetDepartment)
		r.Get("/departments", a.Departments)
		r.Get("/roles", a.Roles)
	})

	// Routes that require the user that performs this action to be authorized
	r.Group(func(r chi.Router) {
		r.Use(a.CurrentUserMiddleware)

		r.Get("/files/{id}/download", a.DownloadFile)
		r.Group(func(r chi.Router) {
			// Token validation
			r.Get("/auth/validate", a.ValidateToken)

			// User routes with current user context
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", a.GetCurrentUser)
				r.Get("/", a.GetUsers)
				r.Get("/{id}", a.GetUser)
			})

			// Achievement groups and templates
			r.Get("/achievement-groups", a.GetAchievementGroups)
			r.Get("/achievement-templates", a.GetAchievementTemplates)

			// Achievement routes
			r.Route("/achievements", func(r chi.Router) {
				// r.Use(a.CurrentUserMiddleware)
				// Routes for current user's achievements
				r.Get("/", a.GetUserAchievements)
				r.Post("/", a.CreateAchievement)

				r.Get("/users", a.GetUsersWithAchievements)

				// Routes for specific achievement
				r.Route("/{id}", func(r chi.Router) {
					r.Use(a.AchievementMiddleware)
					r.Get("/", a.GetAchievement)
					r.Delete("/", a.DeleteAchievement)

					// Document management
					r.Post("/documents", a.AddDocument)
					r.Delete("/documents/{documentId}", a.RemoveDocument)

					// Achievement status management
					r.Post("/submit", a.SubmitAchievement)
					r.Post("/review", a.ReviewAchievement)
					r.Post("/submit-with-new-points", a.SubmitWithNewPoints)
				})
			})

			// File routes
			r.Route("/files", func(r chi.Router) {
				r.Get("/{id}", a.GetFileByID)
				r.Get("/", a.SearchFiles)
				r.Post("/", a.UploadFile)
				r.Delete("/{id}", a.DeleteFile)
				r.Delete("/{id}", a.DeleteFile)
			})
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			// Achievement groups management (admin only)
			r.Post("/achievement-groups", a.CreateAchievementGroup)
			r.Patch("/achievement-groups/{id}", a.PatchAchievementGroup)

			// Achievement templates management (admin only)
			r.Post("/achievement-templates", a.CreateAchievementTemplate)
			r.Patch("/achievement-templates/{id}", a.PatchAchievementTemplate)

			// Credential management
			// TODO r.Get("/auth/credentials/{id}", a.GetCredentials)
		})

		// Reports routes (economist-only)
		r.Group(func(r chi.Router) {
			r.Get("/reports/user-points", a.GenerateUserPointsReport)
			r.Post("/reports/mark-all-accounted", a.MarkAllDoneAchievementsAsAccounted)
		})

		// Swagger UI
		r.Get("/swagger/*", httpSwagger.WrapHandler)

		// Profiler
		r.Group(func(r chi.Router) {
			// TODO not in prod or firewall maybe
			r.Get("/debug/pprof/", pprof.Index)
			r.Get("/debug/pprof/cmdline", pprof.Cmdline)
			r.Get("/debug/pprof/profile", pprof.Profile)
			r.Get("/debug/pprof/symbol", pprof.Symbol)
			r.Get("/debug/pprof/trace", pprof.Trace)
		})
	})
}
