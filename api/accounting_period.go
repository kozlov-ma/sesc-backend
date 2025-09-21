package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	accountingperiod "github.com/kozlov-ma/sesc-backend/accounting_period"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type CreateAccountingPeriodRequest struct {
	Name                   string `json:"name"                      example:"Q1 2025"               description:"Name of the accounting period"`
	Description            string `json:"description"               example:"First quarter of 2025" description:"Description of the accounting period"`
	StartPlanningDate      string `json:"start_planning_date"       example:"2025-01-01"            description:"Start date for planning phase (YYYY-MM-DD format)"`
	StartAchCollectionDate string `json:"start_ach_collection_date" example:"2025-02-01"            description:"Start date for achievement collection phase (YYYY-MM-DD format)"`
	FinishDate             string `json:"finish_date"               example:"2025-03-01"            description:"End date of the accounting period (YYYY-MM-DD format)"`
}

type UpdateAccountingPeriodRequest struct {
	Name                   string `json:"name"                      example:"Q1 2025 Updated"               description:"Name of the accounting period"`
	Description            string `json:"description"               example:"Updated first quarter of 2025" description:"Description of the accounting period"`
	StartPlanningDate      string `json:"start_planning_date"       example:"2025-01-01"                    description:"Start date for planning phase (YYYY-MM-DD format)"`
	StartAchCollectionDate string `json:"start_ach_collection_date" example:"2025-02-01"                    description:"Start date for achievement collection phase (YYYY-MM-DD format)"`
	FinishDate             string `json:"finish_date"               example:"2025-03-01"                    description:"End date of the accounting period (YYYY-MM-DD format)"`
}

//

// CreateAccountingPeriod godoc
// @Summary Create new accounting period
// @Description Создать новый учетный период (всегда создается со статусом "planning")
// @Tags accounting_periods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param request body CreateAccountingPeriodRequest true "New period"
// @Success 201 {object} respond.AccountingPeriod
// @Failure 400 {object} respond.Error "Invalid data"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 409 {object} respond.Error "Planning period already exists"
// @Router /accounting_periods [post]
func (a *API) CreateAccountingPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateAccountingPeriodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	opt := accountingperiod.CreateOptions{
		Name:                   req.Name,
		Description:            req.Description,
		StartPlanningDate:      req.StartPlanningDate,
		StartAchCollectionDate: req.StartAchCollectionDate,
		FinishDate:             req.FinishDate,
	}

	accp, err := a.sesc.CreateAccountingPeriod(ctx, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := respond.WithAccountingPeriod(accp)
	a.writeJSON(ctx, w, respond.WithStatus(response, http.StatusCreated))
}

// UpdateAccountingPeriodInfo godoc
// @Summary Update accounting period
// @Description Редактировать учетный период
// @Tags accounting_periods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param periodID path string true "Period ID"
// @Param request body UpdateAccountingPeriodRequest true "Update period"
// @Success 200 {object} respond.AccountingPeriod
// @Failure 400 {object} respond.Error "Validation failed"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error "Not found"
// @Router /accounting_periods/{periodID} [put]
// === Update Accounting Period ===
func (a *API) UpdateAccountingPeriodInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("periodID")
	rec := event.Get(ctx)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	var req UpdateAccountingPeriodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	opt := accountingperiod.UpdateOptions{
		Name:                   req.Name,
		Description:            req.Description,
		StartPlanningDate:      req.StartPlanningDate,
		StartAchCollectionDate: req.StartAchCollectionDate,
		FinishDate:             req.FinishDate,
	}

	accp, err := a.sesc.UpdateAccountingPeriod(ctx, id, opt)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := respond.WithAccountingPeriod(accp)
	a.writeJSON(ctx, w, respond.WithStatus(response, http.StatusOK))
}

// === Begin Collection ===

// BeginCollection godoc
// @Summary Begin achievement collection
// @Description Перевести период в статус "achcollect"
// @Tags accounting_periods
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param periodID path string true "Period ID"
// @Success 200 {object} respond.AccountingPeriod
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error "Not found"
// @Router /accounting_periods/{periodID}/beginCollection [post]
func (a *API) BeginCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)
	idStr := r.PathValue("periodID")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	accp, err := a.sesc.BeginCollection(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := respond.WithAccountingPeriod(accp)
	a.writeJSON(ctx, w, response)
}

// === Finish Period ===

// FinishPeriod godoc
// @Summary Finish accounting period
// @Description Перевести период в статус "finished"
// @Tags accounting_periods
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param periodID path string true "Period ID"
// @Success 200 {object} respond.AccountingPeriod
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error "Not found"
// @Router /accounting_periods/{periodID}/finish [post]
func (a *API) FinishPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)
	idStr := r.PathValue("periodID")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	accp, err := a.sesc.FinishPeriod(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := respond.WithAccountingPeriod(accp)
	a.writeJSON(ctx, w, response)
}

// === Cancel Period ===

// CancelPeriod godoc
// @Summary Cancel accounting period
// @Description Отменить учетный период
// @Tags accounting_periods
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param periodID path string true "Period ID"
// @Success 200 {object} respond.AccountingPeriod
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error "Not found"
// @Router /accounting_periods/{periodID}/cancel [post]
func (a *API) CancelPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)
	idStr := r.PathValue("periodID")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	accp, err := a.sesc.CancelPeriod(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := respond.WithAccountingPeriod(accp)
	a.writeJSON(ctx, w, response)
}

// === Mark As Non Executed ===

// MarkAsNonExecuted godoc
// @Summary Mark accounting period as not executed
// @Description Перевести период в статус "nonexecuted"
// @Tags accounting_periods
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param periodID path string true "Period ID"
// @Success 200 {object} respond.AccountingPeriod
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error "Not found"
// @Router /accounting_periods/{periodID}/markAsNonExecuted [post]
func (a *API) MarkAsNonExecuted(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)
	idStr := r.PathValue("periodID")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	accp, err := a.sesc.MarkAsNonExecuted(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := respond.WithAccountingPeriod(accp)
	a.writeJSON(ctx, w, response)
}

// === Get Accounting Periods ===

// GetAccountingPeriods godoc
// @Summary Get all accounting periods
// @Description Получить список всех учетных периодов
// @Tags accounting_periods
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} respond.AccountingPeriods
// @Failure 401 {object} respond.Error "Unauthorized"
// @Router /accounting_periods [get]
func (a *API) GetAccountingPeriods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	accps, err := a.sesc.GetAccountingPeriods(ctx)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := respond.WithAccountingPeriods(accps, len(accps))
	a.writeJSON(ctx, w, response)
}
