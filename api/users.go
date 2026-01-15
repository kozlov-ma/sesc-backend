package api

import (
	"errors"
	"net/http"

	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// GetUser godoc
// @Summary Get user details
// @Description Retrieves detailed information about a user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User ID"
// @Success 200 {object} respond.User
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "User not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users/{id} [get]
func (a *API) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Ensure user who asked done auth
	_ = CurrentUser(ctx)

	idStr := r.PathValue("id")

	user, err := a.sesc.User(ctx, idStr)
	if err != nil {
		if errors.Is(err, company.ErrUserNotFound) {
			a.writeJSON(ctx, w, respond.WithUser(company.ExEmployee(idStr)))
			return
		}
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithUser(user))
}

// GetUsers godoc
// @Summary Get all users registered in the system
// @Description Retrieves detailed information about all users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by name"
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} respond.Users
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users [get]
func (a *API) GetUsers(w http.ResponseWriter, r *http.Request) {
	search := param.QueryStringOrZero(r, "search")

	ctx := r.Context()
	rec := event.Get(ctx)

	// Ensure user who asked done auth
	_ = CurrentUser(ctx)

	users, err := a.sesc.Users(ctx, search)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithUsers(users, len(users)))
}

// GetCurrentUser godoc
// @Summary Get current user information
// @Description Returns information about the current authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} respond.User
// @Failure 401 {object} respond.Error "Unauthorized or invalid token"
// @Failure 404 {object} respond.Error "User not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /users/me [get]
func (a *API) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := CurrentUser(ctx)

	a.writeJSON(ctx, w, respond.WithUser(user))
}
