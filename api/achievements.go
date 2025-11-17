package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// Context key for achievement
type achievementContextKey string

const (
	achievementContextKeyValue achievementContextKey = "achievement"
)

func GetAchievementFromContext(ctx context.Context) (*ent.Achievement, bool) {
	ach, ok := ctx.Value(achievementContextKeyValue).(*ent.Achievement)
	return ach, ok
}

// AchievementMiddleware adds the achievement specified by ID in the path to the request context
func (a *API) AchievementMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx)
		rec.Sub("http").Set("route_requires_achievement", true)

		user := CurrentUser(ctx)

		id, err := param.PathUUID(r, "id")
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		ach, err := a.sesc.GetAchievement(ctx, user, id)
		if err != nil {
			rec.Add(events.Error, err)
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		ctx = context.WithValue(ctx, achievementContextKeyValue, ach)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserAchievements godoc
// @Summary Get all achievements for the current user
// @Description Retrieves all achievements for the current user with pagination
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Param limit query int false "Pagination limit" default(10) minimum(1) maximum(100)
// @Param id query string false "User's ID"
// @Param requiring_changes query bool false "Filter achievements requiring changes" default(false)
// @Success 200 {object} respond.Achievements
// @Failure 400 {object} respond.Error "Invalid request parameters"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements [get]
func (a *API) GetUserAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	asking := CurrentUser(ctx)

	target, err := param.QueryString(r, "id")
	if err != nil {
		rec.Add(events.Error, err)
		target = asking.ID
	}

	// Parse pagination parameters
	offset, limit, err := param.QueryPagination(r)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Check if filtering for achievements requiring changes
	requireChangesParam := r.URL.Query().Get("requiring_changes")
	requireChanges := requireChangesParam == "true"

	// Get achievements for user with pagination
	achievements, total, err := a.sesc.GetUserAchievements(
		ctx,
		asking,
		target,
		offset,
		limit,
		requireChanges,
	)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievements(achievements, total)
	a.writeJSON(ctx, w, response)
}

// SubmitAchievementWithNewPoints godoc
// @Summary Submit achievement with updated points
// @Description Allows teachers to update achievement points and resubmit for review
// @Tags achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Param request body param.UpdateAchievementPointsRequest true "Update points request"
// @Success 200 {object} respond.Achievement
// @Failure 400 {object} respond.Error "Invalid request or achievement status"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - not the achievement owner"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/submit-with-new-points [post]
func (a *API) SubmitAchievementWithNewPoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
		return
	}

	// Get current user from context
	user := CurrentUser(ctx)

	// Parse request body
	var req param.UpdateAchievementPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Update achievement points and resubmit
	updatedAch, err := a.sesc.UpdateAchievementPoints(ctx, user, achievement.UpdatePointsOptions{
		AchievementID: ach.ID,
		Points:        req.Points,
		Comment:       req.Comment,
	})
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievement(updatedAch)
	a.writeJSON(ctx, w, response)
}

// GetAchievement godoc
// @Summary Get a specific achievement
// @Description Retrieves a specific achievement by ID
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Success 200 {object} respond.Achievement
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id} [get]
func (a *API) GetAchievement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
		return
	}

	// Convert to response format
	response := respond.WithAchievement(ach)
	a.writeJSON(ctx, w, response)
}

// CreateAchievement godoc
// @Summary Create a new achievement
// @Description Creates a new achievement for the current user
// @Tags achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param request body param.CreateAchievementRequest true "Achievement creation data"
// @Success 201 {object} respond.Achievement
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Template not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements [post]
func (a *API) CreateAchievement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get user from context
	user := CurrentUser(ctx)

	// Parse request
	var req param.CreateAchievementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Create achievement
	opt := achievement.CreateOptions{
		ForUserID:  user.ID,
		TemplateID: req.TemplateID,
	}
	ach, err := a.sesc.CreateAchievement(ctx, user, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	ach, err = a.sesc.GetAchievement(ctx, user, ach.ID)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievement(ach)
	a.writeJSON(ctx, w, respond.WithStatus(response, http.StatusCreated))
}

// DeleteAchievement godoc
// @Summary Delete an achievement
// @Description Deletes an achievement
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Success 204 "No Content"
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 409 {object} respond.Error "Wrong achievement status"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id} [delete]
func (a *API) DeleteAchievement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	user := CurrentUser(ctx)
	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Delete achievement
	opt := achievement.DeleteOptions{
		OwnerID:       ach.OwnerID,
		AchievementID: ach.ID,
	}
	err := a.sesc.DeleteAchievement(ctx, user, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddDocument godoc
// @Summary Add a document to an achievement
// @Description Adds a document to an achievement
// @Tags achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Param request body param.AddDocumentRequest true "Document data"
// @Success 201 {object} respond.Document
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 409 {object} respond.Error "Wrong achievement status"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/documents [post]
func (a *API) AddDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	user := CurrentUser(ctx)

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Parse request
	var req param.AddDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Add document
	opt := achievement.AddDocumentOptions{
		OwnerID:       ach.OwnerID,
		AchievementID: ach.ID,
		Name:          req.Name,
		FileID:        req.FileID,
	}
	doc, err := a.sesc.AddDocument(ctx, user, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.Document{
		ID:     doc.ID,
		Name:   doc.Name,
		FileID: doc.FileID,
	}
	a.writeJSON(ctx, w, respond.WithStatus(response, http.StatusCreated))
}

// RemoveDocument godoc
// @Summary Remove a document from an achievement
// @Description Removes a document from an achievement
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Param documentId path string true "Document UUID"
// @Success 204 "No Content"
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 404 {object} respond.Error "Document not found"
// @Failure 409 {object} respond.Error "Wrong achievement status"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/documents/{documentId} [delete]
func (a *API) RemoveDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	user := CurrentUser(ctx)
	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Get document ID from path
	docIDStr := r.PathValue("documentId")
	docID, err := uuid.FromString(docIDStr)
	if err != nil {
		rec.Add(events.Error, "invalid document ID format")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Remove document
	opt := achievement.RemoveDocumentOptions{
		OwnerID:       ach.OwnerID,
		AchievementID: ach.ID,
		DocumentID:    docID,
	}
	err = a.sesc.RemoveDocument(ctx, user, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SubmitAchievement godoc
// @Summary Submit an achievement for review
// @Description Submits an achievement for review
// @Tags achievements
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Success 200 {object} respond.Achievement
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 409 {object} respond.Error "Wrong achievement status"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/submit [post]
func (a *API) SubmitAchievement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	user := CurrentUser(ctx)

	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
		return
	}

	// Submit achievement
	opt := achievement.SubmitOptions{
		OwnerID:       ach.OwnerID,
		AchievementID: ach.ID,
	}
	updatedAch, err := a.sesc.SubmitAchievement(ctx, user, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	updatedAch, err = a.sesc.GetAchievement(ctx, user, updatedAch.ID)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievement(updatedAch)
	a.writeJSON(ctx, w, response)
}

// ReviewAchievement godoc
// @Summary Review an achievement
// @Description Reviews an achievement with approve, disapprove, or request changes action
// @Tags achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Param request body param.ReviewAchievementRequest true "Review data"
// @Success 200 {object} respond.Achievement
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 400 {object} respond.Error "Invalid review action"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - reviewer role required"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 409 {object} respond.Error "Wrong achievement status"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/review [post]
func (a *API) ReviewAchievement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get user from context
	user := CurrentUser(ctx)

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Parse request
	var req param.ReviewAchievementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Review achievement
	opt := achievement.ReviewOptions{
		AchievementOwnerID: ach.OwnerID,
		AchievementID:      ach.ID,
		Action:             achievement.ReviewAction(req.Action),
		Comment:            req.Comment,
	}
	updatedAch, err := a.sesc.ReviewAchievement(ctx, user, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievement(updatedAch)
	a.writeJSON(ctx, w, response)
}

// SubmitWithNewPoints godoc
// @Summary Submit achievement with updated points
// @Description Allows teachers to submit achievement with updated points when changes are requested by reviewers
// @Tags achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Param request body param.UpdateAchievementPointsRequest true "Points update data"
// @Success 200 {object} respond.Achievement
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 400 {object} respond.Error "Points exceed template limit"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - only achievement owner can update points"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 409 {object} respond.Error "Wrong achievement status - changes not requested"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/submit-with-new-points [post]
func (a *API) SubmitWithNewPoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get user from context
	user := CurrentUser(ctx)

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Verify user is the owner of the achievement
	if ach.OwnerID != user.ID {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrInvalidRole))
		return
	}

	// Parse request
	var req param.UpdateAchievementPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Submit achievement with updated points
	opt := achievement.UpdatePointsOptions{
		AchievementID: ach.ID,
		OwnerID:       ach.OwnerID,
		Points:        req.Points,
		Comment:       req.Comment,
	}
	updatedAch, err := a.sesc.UpdateAchievementPoints(ctx, user, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievement(updatedAch)
	a.writeJSON(ctx, w, response)
}
