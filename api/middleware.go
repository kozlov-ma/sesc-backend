package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/kozlov-ma/sesc-backend/iam"
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
			writeError(a, w, ErrUnauthorized.WithDetails("you are suspended"), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddleware checks for a token in the Authorization header and adds the identity to the context
func (a *API) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authHeader := r.Header.Get("Authorization")

		// Skip auth check if no Authorization header
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Validate Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(a, w, ErrInvalidAuthHeader, http.StatusUnauthorized)
			return
		}

		// Extract token and validate
		token := authHeader[7:]
		identity, err := a.iam.ImWatermelon(ctx, token)
		if err != nil {
			if errors.Is(err, iam.ErrInvalidToken) {
				writeError(a, w, ErrInvalidToken, http.StatusUnauthorized)
				return
			}
			if errors.Is(err, iam.ErrUserNotFound) {
				writeError(a, w, ErrUnauthorized, http.StatusUnauthorized)
				return
			}
			writeError(a, w, ErrServerError, http.StatusInternalServerError)
			return
		}

		// Add identity to context and continue
		ctx = context.WithValue(ctx, identityContextKey, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthMiddleware ensures the request has a valid token
func (a *API) RequireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Check if auth header exists
		if authHeader == "" {
			a.log.DebugContext(r.Context(), "got a request without an auth header", "header", r.Header)
			writeError(a, w, UnauthorizedError{
				Code:      "UNAUTHORIZED",
				Message:   "Authentication required",
				RuMessage: "Требуется аутентификация",
				Details:   "Authentication required",
			}, http.StatusUnauthorized)
			return
		}

		// Continue with auth chain
		next.ServeHTTP(w, r)
	})
}

// RequireAdminRoleMiddleware ensures the request is from a user with Admin role
func (a *API) RequireAdminRoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := GetIdentityFromContext(r.Context())

		// Check if we have identity in the context
		if !ok {
			writeError(a, w, UnauthorizedError{
				Code:      "UNAUTHORIZED",
				Message:   "Authentication required",
				RuMessage: "Требуется аутентификация",
				Details:   "Authentication required",
			}, http.StatusUnauthorized)
			return
		}

		// Check if user has admin role
		if string(identity.Role) != "admin" {
			writeError(a, w, ErrForbidden, http.StatusForbidden)
			return
		}

		// Continue with admin access granted
		next.ServeHTTP(w, r)
	})
}

// RoleMiddleware restricts access to endpoints based on user role
func (a *API) RoleMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get identity from context
			identity, ok := GetIdentityFromContext(ctx)
			if !ok {
				writeError(a, w, UnauthorizedError{
					Code:      "UNAUTHORIZED",
					Message:   "Authentication required",
					RuMessage: "Требуется аутентификация",
					Details:   "Authentication required",
				}, http.StatusUnauthorized)
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if string(identity.Role) == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				writeError(a, w, ErrForbidden, http.StatusForbidden)
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

		identity, ok := GetIdentityFromContext(ctx)
		if !ok {
			a.log.DebugContext(ctx, "identity was not in the context, exiting")
			writeError(a, w, ErrUnauthorized, http.StatusUnauthorized)
			return
		}

		if string(identity.Role) == "user" {
			user, err := a.sesc.User(ctx, identity.ID)
			if err != nil {
				if errors.Is(err, sesc.ErrUserNotFound) {
					a.log.DebugContext(ctx, "sesc user not found, exiting")
					writeError(a, w, ErrUnauthorized, http.StatusUnauthorized)
					return
				}

				writeError(a, w, ErrServerError.WithDetails("Error fetching user data"), http.StatusInternalServerError)
				return
			}

			ctx = context.WithValue(ctx, userContextKey, user)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
