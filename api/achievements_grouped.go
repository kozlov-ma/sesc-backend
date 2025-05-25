package api

import (
	"net/http"

	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// GroupedAchievementsResponse represents a map of user IDs to their achievements
type GroupedAchievementsResponse struct {
	Items      map[string][]AchievementResponse `json:"items"      validate:"required"`
	TotalCount int                              `json:"totalCount" validate:"required"`
	Offset     int                              `json:"offset"     validate:"required"`
	Limit      int                              `json:"limit"      validate:"required"`
}

// GetGroupedAchievements godoc
// @Summary Get achievements grouped by user
// @Description Retrieves all achievements grouped by user with pagination
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Param limit query int false "Pagination limit" default(10) minimum(1) maximum(100)
// @Success 200 {object} GroupedAchievementsResponse
// @Failure 400 {object} InvalidRequestError "Invalid request parameters"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /achievements/grouped [get]
func (a *API) GetGroupedAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("http/get_grouped_achievements")

	// Get user from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		writeError(ctx, w, ErrUnauthorized.WithStatus(http.StatusUnauthorized))
		return
	}

	// Check if user has permission to view grouped achievements
	// Only department heads and deputies should have access
	switch user.Role.ID {
	case sesc.Dephead.ID, sesc.ContestDeputy.ID, sesc.DevelopmentDeputy.ID, sesc.ScientificDeputy.ID:
	default:
		writeError(ctx, w, ErrForbidden.WithStatus(http.StatusForbidden))
		return
	}

	// Parse pagination parameters
	offset, limit, err := parsePaginationParams(r)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, InvalidRequestError{
			Code:      "INVALID_REQUEST",
			Message:   "Invalid pagination parameters",
			RuMessage: "Некорректные параметры пагинации",
		}.WithDetails(err.Error()).WithStatus(http.StatusBadRequest))
		return
	}

	// Get grouped achievements
	groupedAchievements, totalCount, err := a.sesc.GetGroupedAchievements(ctx, offset, limit)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
		return
	}

	// Convert to response format
	response := GroupedAchievementsResponse{
		Items:      make(map[string][]AchievementResponse),
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}

	for userID, achievements := range groupedAchievements {
		userAchievements := make([]AchievementResponse, 0, len(achievements))
		for _, ach := range achievements {
			userAchievements = append(userAchievements, convertAchievementToResponse(ach))
		}
		response.Items[userID.String()] = userAchievements
	}

	// Write response
	a.writeJSON(ctx, w, response, http.StatusOK)
}
