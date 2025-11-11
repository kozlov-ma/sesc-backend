package api

import (
	"net/http"

	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// GetUsersWithAchievements godoc
// @Summary Get users with achievements
// @Description Retrieves users with achievements based on role permissions
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Param limit query int false "Pagination limit" default(10) minimum(1) maximum(100)
// @Param search query string false "Search by name"
// @Success 200 {object} respond.Users
// @Failure 400 {object} respond.Error "Invalid request parameters"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/users [get]
func (a *API) GetUsersWithAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("http/get_users_with_achievements")

	// Get user from context
	user := CurrentUser(ctx)

	search := param.QueryStringOrZero(r, "search")

	// Parse pagination parameters
	offset, limit, err := param.ParsePagination(r)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Get users with achievements
	users, totalCount, err := a.sesc.GetUsersWithAchievements(ctx, user, offset, limit, search)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithUsers(users, totalCount)
	a.writeJSON(ctx, w, response)
}
