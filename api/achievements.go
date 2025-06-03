package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// Context key for achievement
type achievementContextKey string

const (
	achievementContextKeyValue achievementContextKey = "achievement"
)

func GetAchievementFromContext(ctx context.Context) (achievement.Achievement, bool) {
	ach, ok := ctx.Value(achievementContextKeyValue).(achievement.Achievement)
	return ach, ok
}

// AchievementMiddleware adds the achievement specified by ID in the path to the request context
func (a *API) AchievementMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx)
		rec.Sub("http").Set("route_requires_achievement", true)

		_, ok := GetUserFromContext(ctx)
		if !ok {
			a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
			return
		}

		id, err := param.PathUUID(r, "id")
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		ach, err := a.sesc.GetAchievement(ctx, id)
		if err != nil {
			rec.Add(events.Error, err)
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		ctx = context.WithValue(ctx, achievementContextKeyValue, ach)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper function to convert an achievement to a response
func convertAchievementToResponse(ach achievement.Achievement) AchievementResponse {
	response := AchievementResponse{
		ID:           ach.ID,
		OwnerID:      ach.Owner.ID,
		OwnerName:    fmt.Sprintf("%s %s %s", ach.Owner.LastName, ach.Owner.FirstName, ach.Owner.MiddleName),
		TemplateID:   ach.Template.ID,
		TemplateName: ach.Template.Name,
		Status:       ach.Status,
		Points:       ach.Points,
		Documents:    make([]DocumentResponse, 0, len(ach.Documents)),
		Reviews:      make([]ReviewResponse, 0, len(ach.Reviews)),
	}

	// Convert documents
	for _, doc := range ach.Documents {
		response.Documents = append(response.Documents, DocumentResponse{
			ID:     doc.ID,
			Name:   doc.Name,
			FileID: doc.FileID,
		})
	}

	// Convert reviews
	for _, rev := range ach.Reviews {
		response.Reviews = append(response.Reviews, ReviewResponse{
			ID:             rev.ID,
			ReviewerID:     rev.From.ID,
			ReviewerName:   fmt.Sprintf("%s %s %s", rev.From.LastName, rev.From.FirstName, rev.From.MiddleName),
			PointsAssigned: rev.PointsAssigned,
			Comment:        rev.Comment,
		})
	}

	return response
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
// @Success 200 {object} PaginatedAchievementsResponse
// @Failure 400 {object} respond.Error "Invalid request parameters"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements [get]
func (a *API) GetUserAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get user from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
		return
	}

	// Parse pagination parameters
	offset, limit, err := parsePaginationParams(r)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Get achievements for user with pagination
	achievements, total, err := a.sesc.GetUserAchievements(ctx, user.ID, offset, limit)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	items := make([]AchievementResponse, 0, len(achievements))
	for _, ach := range achievements {
		items = append(items, convertAchievementToResponse(ach))
	}

	// Create paginated response
	response := PaginatedAchievementsResponse{
		Items:      items,
		TotalCount: total,
		Offset:     offset,
		Limit:      limit,
	}

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
// @Success 200 {object} AchievementResponse
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
	response := convertAchievementToResponse(ach)
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
// @Param request body CreateAchievementRequest true "Achievement creation data"
// @Success 201 {object} AchievementResponse
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Template not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements [post]
func (a *API) CreateAchievement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get user from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
		return
	}

	var depID uuid.UUID
	if user.DepartmentID != nil {
		depID = *user.DepartmentID
	}

	sus := sesc.User{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		PictureURL: user.PictureURL,
		Suspended:  user.Suspended,
		Department: sesc.Department{
			ID: depID,
		},
		Role:              user.Role,
		Subdivision:       user.Subdivision,
		JobTitle:          user.JobTitle,
		EmploymentRate:    user.EmploymentRate,
		AcademicDegree:    user.AcademicDegree,
		PersonnelCategory: user.PersonnelCategory,
		EmploymentType:    user.EmploymentType,
		AcademicTitle:     user.AcademicTitle,
		Honors:            user.Honors,
		Category:          user.Category,
		DateOfEmployment:  user.DateOfEmployment,
		UnemploymentDate:  user.UnemploymentDate,
	}

	// Parse request
	var req CreateAchievementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Create achievement
	opt := achievement.CreateOptions{
		ForUser:    sus,
		TemplateID: req.TemplateID,
	}
	ach, err := a.sesc.CreateAchievement(ctx, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := convertAchievementToResponse(ach)
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

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Delete achievement
	opt := achievement.DeleteOptions{
		OwnerID:       ach.Owner.ID,
		AchievementID: ach.ID,
	}
	err := a.sesc.DeleteAchievement(ctx, opt)
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
// @Param request body AddDocumentRequest true "Document data"
// @Success 201 {object} DocumentResponse
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 409 {object} respond.Error "Wrong achievement status"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/documents [post]
func (a *API) AddDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Parse request
	var req AddDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Add document
	opt := achievement.AddDocumentOptions{
		OwnerID:       ach.Owner.ID,
		AchievementID: ach.ID,
		Name:          req.Name,
		FileID:        req.FileID,
	}
	doc, err := a.sesc.AddDocument(ctx, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := DocumentResponse{
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
		OwnerID:       ach.Owner.ID,
		AchievementID: ach.ID,
		DocumentID:    docID,
	}
	err = a.sesc.RemoveDocument(ctx, opt)
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
// @Success 200 {object} AchievementResponse
// @Failure 400 {object} respond.Error "Invalid UUID format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error "Achievement not found"
// @Failure 409 {object} respond.Error "Wrong achievement status"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievements/{id}/submit [post]
func (a *API) SubmitAchievement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Submit achievement
	opt := achievement.SubmitOptions{
		OwnerID:       ach.Owner.ID,
		AchievementID: ach.ID,
	}
	updatedAch, err := a.sesc.SubmitAchievement(ctx, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := convertAchievementToResponse(updatedAch)
	a.writeJSON(ctx, w, response)
}

// ReviewAchievement godoc
// @Summary Review an achievement
// @Description Reviews an achievement, setting points and optionally a comment
// @Tags achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Achievement UUID"
// @Param request body ReviewAchievementRequest true "Review data"
// @Success 200 {object} AchievementResponse
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 400 {object} respond.Error "Points assigned exceed the template's points limit"
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
	reviewer, ok := GetUserFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, sesc.ErrUserNotFound))
		return
	}

	// Get achievement from context (added by AchievementMiddleware)
	ach, ok := GetAchievementFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrAchievementNotFound))
		return
	}

	// Parse request
	var req ReviewAchievementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Review achievement
	opt := achievement.ReviewOptions{
		AchievementOwnerID: ach.Owner.ID,
		AchievementID:      ach.ID,
		ReviewerID:         reviewer.ID,
		PointsAssigned:     req.PointsAssigned,
		Comment:            req.Comment,
	}
	updatedAch, err := a.sesc.ReviewAchievement(ctx, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := convertAchievementToResponse(updatedAch)
	a.writeJSON(ctx, w, response)
}

// Helper function to parse pagination parameters from request
func parsePaginationParams(r *http.Request) (offset, limit int, err error) {
	// Default values
	offset = 0
	limit = 10

	// Parse offset
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return 0, 0, errors.New("invalid offset parameter")
		}
	}

	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, errors.New("invalid limit parameter")
		}
	}

	return offset, limit, nil
}

// PaginatedAchievementsResponse represents a paginated list of achievements
type PaginatedAchievementsResponse struct {
	Items      []AchievementResponse `json:"items"      validate:"required"`
	TotalCount int                   `json:"totalCount" validate:"required"`
	Offset     int                   `json:"offset"     validate:"required"`
	Limit      int                   `json:"limit"      validate:"required"`
}

// AchievementResponse represents the API response for an achievement
type AchievementResponse struct {
	ID           uuid.UUID          `json:"id"           example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	OwnerID      uuid.UUID          `json:"ownerId"      example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	OwnerName    string             `json:"ownerName"    example:"Иванов Иван Иванович"                 validate:"required"`
	TemplateID   uuid.UUID          `json:"templateId"   example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	TemplateName string             `json:"templateName" example:"регионального уровня"                 validate:"required"`
	Status       achievement.Status `json:"status"       example:"draft"                                validate:"required"`
	Points       int                `json:"points"       example:"10"                                   validate:"required"`
	Documents    []DocumentResponse `json:"documents"                                                   validate:"required"`
	Reviews      []ReviewResponse   `json:"reviews"                                                     validate:"required"`
}

// DocumentResponse represents the API response for a document
type DocumentResponse struct {
	ID     uuid.UUID `json:"id"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name   string    `json:"name"   example:"Publication proof"                    validate:"required"`
	FileID uuid.UUID `json:"fileId" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
}

// ReviewResponse represents the API response for a review
type ReviewResponse struct {
	ID             uuid.UUID `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	ReviewerID     uuid.UUID `json:"reviewerId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	ReviewerName   string    `json:"reviewerName"   example:"Петров Петр Петрович"                 validate:"required"`
	PointsAssigned int       `json:"pointsAssigned" example:"8"                                    validate:"required"`
	Comment        string    `json:"comment"        example:"Good job, but could be better"        validate:"omitempty"`
}

// CreateAchievementRequest represents the request to create a new achievement
type CreateAchievementRequest struct {
	TemplateID uuid.UUID `json:"templateId" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
}

// AddDocumentRequest represents the request to add a document to an achievement
type AddDocumentRequest struct {
	Name   string    `json:"name"   example:"Publication proof"                    validate:"required"`
	FileID uuid.UUID `json:"fileId" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
}

// ReviewAchievementRequest represents the request to review an achievement
type ReviewAchievementRequest struct {
	PointsAssigned int    `json:"pointsAssigned" example:"8"                             validate:"required"`
	Comment        string `json:"comment"        example:"Good job, but could be better" validate:"omitempty"`
}
