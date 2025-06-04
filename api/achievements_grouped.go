package api

import (
	"net/http"

	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// UsersWithAchievementsResponse represents a list of users with their achievement counts
type UsersWithAchievementsResponse struct {
	Items      []UserWithAchievementsResponse `json:"items"      validate:"required"`
	TotalCount int                            `json:"totalCount" validate:"required"`
	Offset     int                            `json:"offset"     validate:"required"`
	Limit      int                            `json:"limit"      validate:"required"`
}

// UserWithAchievementsResponse represents a user with achievement summary
type UserWithAchievementsResponse struct {
	ID         string `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	FirstName  string `json:"firstName"  example:"Иван"                                 validate:"required"`
	LastName   string `json:"lastName"   example:"Иванов"                               validate:"required"`
	MiddleName string `json:"middleName" example:"Иванович"`
	Role       string `json:"role"       example:"teacher"                              validate:"required"`
}

// GetUsersWithAchievements godoc
// @Summary Get users with achievements
// @Description Retrieves users with achievements based on role permissions
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Param limit query int false "Pagination limit" default(10) minimum(1) maximum(100)
// @Success 200 {object} UsersWithAchievementsResponse
// @Failure 400 {object} respond.Error "Invalid request parameters"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/users [get]
func (a *API) GetUsersWithAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("http/get_users_with_achievements")

	// Get user from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
		return
	}

	// Check if user has permission to view users with achievements
	// Only department heads and deputies should have access
	//nolint:exhaustive // cuz fuck it here.
	switch user.Role {
	case sesc.Dephead,
		sesc.OlympiadDeputy,
		sesc.DevelopmentDeputy,
		sesc.ScientificDeputy,
		sesc.AcademicDirector,
		sesc.ChiefEconomist:
	default:
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrInvalidRole))
		return
	}

	// Parse pagination parameters
	offset, limit, err := param.ParsePagination(r)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Get users with achievements
	users, totalCount, err := a.sesc.GetUsersWithAchievements(ctx, user.ID, offset, limit)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := UsersWithAchievementsResponse{
		Items:      make([]UserWithAchievementsResponse, 0, len(users)),
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}

	for _, u := range users {
		response.Items = append(response.Items, UserWithAchievementsResponse{
			ID:         u.ID.String(),
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			MiddleName: u.MiddleName,
			Role:       u.Role.String(),
		})
	}

	// Write response
	a.writeJSON(ctx, w, response)
}
