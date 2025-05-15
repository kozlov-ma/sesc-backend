package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type contextKey string

const (
	identityContextKey contextKey = "identity"
	userContextKey     contextKey = "user"
)

// GetIdentityFromContext retrieves the identity from the request context if it exists
func GetIdentityFromContext(ctx context.Context) (iam.Identity, bool) {
	identity, ok := ctx.Value(identityContextKey).(iam.Identity)
	return identity, ok
}

// GetUserFromContext retrieves the user from the request context if it exists
func GetUserFromContext(ctx context.Context) (sesc.User, bool) {
	user, ok := ctx.Value(userContextKey).(sesc.User)
	return user, ok
}

func (a *API) UnauthorizeSuspendedUsersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, ok := GetUserFromContext(ctx)
		if ok && u.Suspended {
			writeError(ctx, w, ErrUnauthorized.WithDetails("you are suspended"), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddleware checks for a token in the Authorization header and adds the identity to the context
func (a *API) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx)

		authHeader := r.Header.Get("Authorization")

		rec.Sub("identity").Set("authorized", false)

		// Skip auth check if no Authorization header
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Validate Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(ctx, w, ErrInvalidAuthHeader, http.StatusUnauthorized)
			return
		}

		// Extract token and validate
		token := authHeader[7:]
		identity, err := a.iam.ImWatermelon(ctx, token)
		if err != nil {
			if errors.Is(err, iam.ErrInvalidToken) {
				writeError(ctx, w, ErrInvalidToken, http.StatusUnauthorized)
				return
			}
			if errors.Is(err, iam.ErrUserNotFound) {
				writeError(ctx, w, ErrUnauthorized, http.StatusUnauthorized)
				return
			}
			rec.Add(events.Error, err)
			writeError(ctx, w, ErrServerError, http.StatusInternalServerError)
			return
		}

		rec.Sub("identity").Set(
			"authorized", true,
			"auth_id", identity.AuthID,
			"id", identity.ID,
			"role", identity.Role,
		)

		ctx = context.WithValue(ctx, identityContextKey, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthMiddleware ensures the request has a valid token
func (a *API) RequireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx)

		rec.Sub("http").Set("route_requires_auth", true)

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			writeError(ctx, w, UnauthorizedError{
				Code:      "UNAUTHORIZED",
				Message:   "Authentication required",
				RuMessage: "Требуется аутентификация",
				Details:   "Authentication required",
			}, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdminRoleMiddleware ensures the request is from a user with Admin role
func (a *API) RequireAdminRoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx)

		rec.Sub("http").Set("admin_route", true)

		identity, ok := GetIdentityFromContext(ctx)

		// Check if we have identity in the context
		if !ok {
			writeError(ctx, w, UnauthorizedError{
				Code:      "UNAUTHORIZED",
				Message:   "Authentication required",
				RuMessage: "Требуется аутентификация",
				Details:   "Authentication required",
			}, http.StatusUnauthorized)
			return
		}

		// Check if user has admin role
		if string(identity.Role) != "admin" {
			writeError(ctx, w, ErrForbidden, http.StatusForbidden)
			return
		}

		rec.Set("admin", true)

		// Continue with admin access granted
		next.ServeHTTP(w, r)
	})
}

// RoleMiddleware restricts access to endpoints based on user role
func (a *API) RoleMiddleware(role iam.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			rec := event.Get(ctx)

			rec.Sub("http").Set("required_role", role)

			// Get identity from context
			identity, ok := GetIdentityFromContext(ctx)
			if !ok {
				writeError(ctx, w, UnauthorizedError{
					Code:      "UNAUTHORIZED",
					Message:   "Authentication required",
					RuMessage: "Требуется аутентификация",
					Details:   "Authentication required",
				}, http.StatusUnauthorized)
				return
			}

			// Check if user has required role
			if identity.Role != role {
				writeError(ctx, w, ErrForbidden, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CurrentUserMiddleware adds the current user to the request context if available
func (a *API) CurrentUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx)
		rec.Sub("http").Set("route_requires_auth", true)
		rec.Sub("http").Set("route_requires_current_user", true)

		identity, ok := GetIdentityFromContext(ctx)
		if !ok {
			writeError(ctx, w, ErrUnauthorized, http.StatusUnauthorized)
			return
		}

		if string(identity.Role) == "user" {
			user, err := a.sesc.User(ctx, identity.ID)
			if err != nil {
				if errors.Is(err, sesc.ErrUserNotFound) {
					writeError(ctx, w, ErrUnauthorized, http.StatusUnauthorized)
					return
				}

				rec.Add(events.Error, fmt.Errorf("couldn't get user data: %w", err))
				writeError(
					ctx,
					w,
					ErrServerError.WithDetails("Error fetching user data"),
					http.StatusInternalServerError,
				)
				return
			}

			rec.Set("user", user.EventRecord())

			ctx = context.WithValue(ctx, userContextKey, user)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) EventMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, rec := event.NewRecord(ctx, "http_request")

		defer func() {
			if r := recover(); r != nil {
				rec.Set("panic", r)
				rec.Set("panic_message", fmt.Sprintf("%v", r))
				a.eventSink.ProcessEvent(rec)
				panic(r)
			}
		}()

		httprec := rec.Sub("http")

		rec.Set(
			"time", time.Now(),
		)

		httprec.Sub("request").Set(
			"method", r.Method,
			"path", r.URL.Path,
			"proto", r.Proto,
			"authorization_header_present", r.Header.Get("Authorization") != "",
			"content_length", r.ContentLength,
			"host", r.Host,
			"form_values", formValues(r.Form),
			"remote_addr", r.RemoteAddr,
			"header", event.Group(
				"content_type", r.Header.Get("Content-Type"),
			),
		)

		m := httpsnoop.CaptureMetrics(next, w, r.WithContext(ctx))

		rec.Set(
			"processing_time", m.Duration,
		)

		httprec.Sub("response").Set(
			"code", m.Code,
			"bytes_written", m.Written,
			"header", event.Group(
				"content_type", w.Header().Get("Content-Type"),
				"access_control_allow_origin", w.Header().Get("Access-Control-Allow-Origin"),
				"access_control_allow_methods", w.Header().Get("Access-Control-Allow-Methods"),
				"access_control_allow_headers", w.Header().Get("Access-Control-Allow-Headers"),
			),
		)

		a.eventSink.ProcessEvent(rec)
	})
}

func formValues(vals url.Values) *event.Record {
	const recordValuesPerFormValue = 2
	values := make([]any, 0, len(vals)*recordValuesPerFormValue)
	for key, val := range vals {
		values = append(values, key, strings.Join(val, ","))
	}
	return event.Group(values...)
}
