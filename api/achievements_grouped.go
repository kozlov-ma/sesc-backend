package api

import (
	"net/http"

	"github.com/google/uuid"
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

// UserAchievementsResponse represents achievements for a specific user
type UserAchievementsResponse struct {
	UserID       string                `json:"userId"     validate:"required"`
	Achievements []AchievementResponse `json:"achievements" validate:"required"`
	TotalCount   int                   `json:"totalCount" validate:"required"`
	Offset       int                   `json:"offset"     validate:"required"`
	Limit        int                   `json:"limit"      validate:"required"`
}

// UsersWithAchievementsResponse represents users who have at least one achievement
type UsersWithAchievementsResponse struct {
	Users      []UserWithAchievementCount `json:"users"      validate:"required"`
	TotalCount int                        `json:"totalCount" validate:"required"`
	Offset     int                        `json:"offset"     validate:"required"`
	Limit      int                        `json:"limit"      validate:"required"`
}

// UserWithAchievementCount represents a user with their achievement count
type UserWithAchievementCount struct {
	UserID           string `json:"userId"           validate:"required"`
	UserName         string `json:"userName"         validate:"required"`
	AchievementCount int    `json:"achievementCount" validate:"required"`
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

// GetUserAchievementsByID godoc
// @Summary Get all achievements for a specific user by ID
// @Description Retrieves all achievements for a specific user with pagination
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param userId path string true "User ID"
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Param limit query int false "Pagination limit" default(10) minimum(1) maximum(100)
// @Success 200 {object} UserAchievementsResponse
// @Failure 400 {object} InvalidRequestError "Invalid request parameters"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden"
// @Failure 404 {object} NotFoundError "User not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /achievements/user/{userId} [get]
func (a *API) GetUserAchievementsByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("http/get_user_achievements_by_id")

	// Get user from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		writeError(ctx, w, ErrUnauthorized.WithStatus(http.StatusUnauthorized))
		return
	}

	// Get user ID from URL path
	userID := r.PathValue("userId")
	if userID == "" {
		rec.Add(events.Error, "missing user ID in path")
		writeError(ctx, w, InvalidRequestError{
			Code:      "INVALID_REQUEST",
			Message:   "User ID is required",
			RuMessage: "Требуется ID пользователя",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	// Parse user ID
	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, InvalidRequestError{
			Code:      "INVALID_USER_ID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}.WithDetails(err.Error()).WithStatus(http.StatusBadRequest))
		return
	}

	// Convert to sesc.UUID type
	targetUserID := sesc.UUID(parsedUUID)

	// Check permissions - users can view their own achievements,
	// department heads and deputies can view all achievements
	canViewAchievements := false
	switch {
	case user.ID == targetUserID:
		// Users can view their own achievements
		canViewAchievements = true
	case user.Role.ID == sesc.Dephead.ID ||
		user.Role.ID == sesc.ContestDeputy.ID ||
		user.Role.ID == sesc.DevelopmentDeputy.ID ||
		user.Role.ID == sesc.ScientificDeputy.ID:
		// Department heads and deputies can view all achievements
		canViewAchievements = true
	}

	if !canViewAchievements {
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

	// Get user achievements - изменил вызов метода
	achievements, totalCount, err := a.sesc.GetUserAchievementsByID(ctx, targetUserID, offset, limit)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
		return
	}

	// Convert to response format
	achievementResponses := make([]AchievementResponse, 0, len(achievements))
	for _, ach := range achievements {
		achievementResponses = append(achievementResponses, convertAchievementToResponse(ach))
	}

	response := UserAchievementsResponse{
		UserID:       userID,
		Achievements: achievementResponses,
		TotalCount:   totalCount,
		Offset:       offset,
		Limit:        limit,
	}

	// Write response
	a.writeJSON(ctx, w, response, http.StatusOK)
}

// GetUsersWithAchievements godoc
// @Summary Get all users who have at least one achievement
// @Description Retrieves all users who have at least one achievement with pagination
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Param limit query int false "Pagination limit" default(10) minimum(1) maximum(100)
// @Success 200 {object} UsersWithAchievementsResponse
// @Failure 400 {object} InvalidRequestError "Invalid request parameters"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /achievements/users [get]
func (a *API) GetUsersWithAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("http/get_users_with_achievements")

	// Get user from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		writeError(ctx, w, ErrUnauthorized.WithStatus(http.StatusUnauthorized))
		return
	}

	// Check if user has permission to view users with achievements
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

	// Get users with achievements
	usersWithAchievements, totalCount, err := a.sesc.GetUsersWithAchievements(ctx, offset, limit)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
		return
	}

	// Convert to response format - исправил конверсию
	userResponses := make([]UserWithAchievementCount, 0, len(usersWithAchievements))
	for _, userWithCount := range usersWithAchievements {
		// Не нужно конвертировать UserWithAchievementCount в UserWithAchievementCount
		// Но нужно конвертировать UUID в string для JSON ответа
		userResponses = append(userResponses, UserWithAchievementCount{
			UserID:           userWithCount.UserID.String(),
			UserName:         userWithCount.UserName,
			AchievementCount: userWithCount.AchievementCount,
		})
	}

	response := UsersWithAchievementsResponse{
		Users:      userResponses,
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}

	// Write response
	a.writeJSON(ctx, w, response, http.StatusOK)
}
