package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
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
// @Failure 401 {object} Error "Unauthorized"
// @Failure 403 {object} Error "Forbidden - Admin access required"
// @Failure 500 {object} Error "Internal server error"
// @Router /reports/user-points [get]
// @Security BearerAuth
func (a *API) GenerateUserPointsReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/generate_user_points_report")

	// Generate the Excel report
	excelBuffer, err := a.sesc.GenerateUserPointsReport(ctx)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ServerError{
			Code:      "REPORT_GENERATION_ERROR",
			Message:   "Failed to generate user points report",
			RuMessage: "Не удалось создать отчет по баллам пользователей",
			Details:   err.Error(),
		}.WithStatus(http.StatusInternalServerError))
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

// MarkAchievementsAsAccountedRequest represents the request body for marking achievements as accounted
type MarkAchievementsAsAccountedRequest struct {
	AchievementIDs []string `json:"achievementIds" binding:"required"`
}

// MarkAchievementsAsAccounted marks achievements with "done" status as "accounted"
// @Summary Mark achievements as accounted
// @Description Marks achievements with "done" status as "accounted" in the system
// @Tags reports
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param request body MarkAchievementsAsAccountedRequest true "Achievement IDs to mark as accounted"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} Error "Bad request"
// @Failure 401 {object} Error "Unauthorized"
// @Failure 403 {object} Error "Forbidden - Admin access required"
// @Failure 500 {object} Error "Internal server error"
// @Router /reports/mark-accounted [post]
// @Security BearerAuth
func (a *API) MarkAchievementsAsAccounted(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/mark_achievements_as_accounted")

	// Parse request body
	var req MarkAchievementsAsAccountedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to decode request body: %w", err))
		writeError(ctx, w, BadRequestError{
			Code:      "INVALID_REQUEST_BODY",
			Message:   "Invalid request body",
			RuMessage: "Неверное тело запроса",
			Details:   err.Error(),
		}.WithStatus(http.StatusBadRequest))
		return
	}

	if len(req.AchievementIDs) == 0 {
		writeError(ctx, w, BadRequestError{
			Code:      "EMPTY_ACHIEVEMENT_LIST",
			Message:   "Achievement IDs list cannot be empty",
			RuMessage: "Список ID достижений не может быть пустым",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	// Convert string IDs to UUIDs
	achievementIDs := make([]uuid.UUID, len(req.AchievementIDs))
	for i, idStr := range req.AchievementIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			rec.Add(events.Error, fmt.Errorf("invalid achievement ID format: %s", idStr))
			writeError(ctx, w, BadRequestError{
				Code:      "INVALID_ACHIEVEMENT_ID",
				Message:   fmt.Sprintf("Invalid achievement ID format: %s", idStr),
				RuMessage: fmt.Sprintf("Неверный формат ID достижения: %s", idStr),
			}.WithStatus(http.StatusBadRequest))
			return
		}
		achievementIDs[i] = id
	}

	// Mark achievements as accounted
	err := a.sesc.MarkAchievementsAsAccounted(ctx, achievementIDs)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ServerError{
			Code:      "MARK_ACCOUNTED_ERROR",
			Message:   "Failed to mark achievements as accounted",
			RuMessage: "Не удалось отметить достижения как учтенные",
			Details:   err.Error(),
		}.WithStatus(http.StatusInternalServerError))
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Achievements marked as accounted successfully",
		"count":   len(achievementIDs),
	}

	a.writeJSON(ctx, w, response, http.StatusOK)

	rec.Set("success", true)
	rec.Set("marked_count", len(achievementIDs))
}

// MarkAllDoneAchievementsAsAccounted marks all achievements with "done" status as "accounted"
// @Summary Mark all done achievements as accounted
// @Description Marks all achievements with "done" status as "accounted" in the system
// @Tags reports
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 401 {object} Error "Unauthorized"
// @Failure 403 {object} Error "Forbidden - Economist access required"
// @Failure 500 {object} Error "Internal server error"
// @Router /reports/mark-all-accounted [post]
// @Security BearerAuth
func (a *API) MarkAllDoneAchievementsAsAccounted(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/mark_all_done_achievements_as_accounted")

	// Mark all done achievements as accounted
	count, err := a.sesc.MarkAllDoneAchievementsAsAccounted(ctx)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ServerError{
			Code:      "MARK_ALL_ACCOUNTED_ERROR",
			Message:   "Failed to mark all done achievements as accounted",
			RuMessage: "Не удалось отметить все выполненные достижения как учтенные",
			Details:   err.Error(),
		}.WithStatus(http.StatusInternalServerError))
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "All done achievements marked as accounted successfully",
		"count":   count,
	}

	a.writeJSON(ctx, w, response, http.StatusOK)

	rec.Set("success", true)
	rec.Set("marked_count", count)
}
