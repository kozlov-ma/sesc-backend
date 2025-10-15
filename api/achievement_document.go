package api

import (
	"net/http"
	"time"

	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

const (
	// DocumentDeletionDelay is the delay before documents are permanently deleted
	DocumentDeletionDelay = 24 * time.Hour
)

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
	err := a.sesc.ScheduleDocumentDeletionAll(ctx, DocumentDeletionDelay)
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

	a.writeJSON(ctx, w, respond.WithDocumentStats(stats, DocumentDeletionDelay.String()))
}
