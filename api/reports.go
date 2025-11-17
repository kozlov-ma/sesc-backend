package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// GenerateUserPointsReport generates an Excel report with user achievement points
// @Summary Generate user points report
// @Description Generates an Excel report containing all users with their achievement points summary
// @Tags reports
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {file} binary "Excel file with user points report"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - Admin access required"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /reports/user-points [get]
// @Security BearerAuth
func (a *API) GenerateUserPointsReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/generate_user_points_report")

	// Get user from context
	user := CurrentUser(ctx)

	// Generate the Excel report
	excelBuffer, err := a.sesc.GenerateUserPointsReport(ctx, user)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Set response headers for Excel file download
	filename := fmt.Sprintf("user_points_report_%s.xlsx", time.Now().Format("2006-01-02_15-04-05"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.Itoa(excelBuffer.Len()))

	// Write the Excel file content
	_, err = w.Write(excelBuffer.Bytes())
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to write excel response: %w", err))
		return
	}

	rec.Set("success", true)
	rec.Set("file_size_bytes", excelBuffer.Len())
	rec.Set("filename", filename)
}

// MarkAllDoneAchievementsAsAccounted marks all achievements with "done" status as "accounted"
// @Summary Mark all done achievements as accounted
// @Description Marks all achievements with "done" status as "accounted" in the system
// @Tags reports
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - Economist access required"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /reports/mark-all-accounted [post]
// @Security BearerAuth
func (a *API) MarkAllDoneAchievementsAsAccounted(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/mark_all_done_achievements_as_accounted")

	// Get user from context
	user := CurrentUser(ctx)

	// Mark all done achievements as accounted
	count, err := a.sesc.MarkAllDoneAchievementsAsAccounted(ctx, user)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "All done achievements marked as accounted successfully",
		"count":   count,
	}

	a.writeJSON(ctx, w, response)

	rec.Set("success", true)
	rec.Set("marked_count", count)
}
