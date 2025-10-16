package api

import (
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

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
	doc, err := a.sesc.AddDocument(ctx, opt)
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

// ScheduleDeletionAll schedules deletion for all documents
// @Summary Schedule deletion for all documents
// @Description Schedules deletion for all documents
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Success 204 "No Content"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 500 {object} respond.Error
// @Router /documents/schedule_deletion/all [post]
// @Security BearerAuth
func (a *API) ScheduleDeletionAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/schedule_deletion_all")
	// Schedule all documents for deletion
	err := a.sesc.ScheduleDocumentDeletionAll(ctx)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDocumentStats returns statistics about documents
// @Summary Get document statistics
// @Description Returns statistics about documents including total, deleted, scheduled, etc.
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} respond.DocumentStats
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 500 {object} respond.Error
// @Router /documents/stats [get]
// @Security BearerAuth
func (a *API) GetDocumentStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/get_document_stats")

	// Get statistics from achievement service
	stats, err := a.sesc.GetDocumentStats(ctx)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, respond.WithDocumentStats(stats, a.deletionDelay.String()))
}

// ScheduleDeletion godoc
// @Summary Schedule document deletion by ID (admin only)
// @Description Schedules a specific document for deletion
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param documentId path string true "Document UUID"
// @Success 204 "No Content"
// @Failure 400 {object} respond.Error "Invalid document ID"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error "Document not found"
// @Failure 500 {object} respond.Error
// @Router /documents/schedule_deletion/{documentId} [post]
// @Security BearerAuth
func (a *API) ScheduleDeletion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/schedule_deletion")

	// Get document ID from path
	docIDStr := r.PathValue("documentId")
	docID, err := uuid.FromString(docIDStr)
	if err != nil {
		rec.Add(events.Error, "invalid document ID format")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	rec.Sub("params").Set("document_id", docID)

	// Schedule document deletion
	opt := achievement.ScheduleDocumentDeletionOptions{
		DocumentID: docID,
	}
	err = a.sesc.ScheduleDeletion(ctx, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	rec.Set("success", true)
	w.WriteHeader(http.StatusNoContent)
}

// RemoveDocument godoc
// @Summary Remove a document from an achievement immediately
// @Description Immediately removes a document from an achievement (marks as deleted)
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

	// Delete document immediately
	opt := achievement.RemoveDocumentOptions{
		OwnerID:       ach.OwnerID,
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
