package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// GetAchievementGroups godoc
// @Summary Get all achievement groups
// @Description Retrieves all achievement groups with filtering options
// @Tags achievement-groups
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param show_inactive query bool false "Show inactive groups" default(false)
// @Param search query string false "Search by name"
// @Success 200 {array} respond.AchievementGroup
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-groups [get]
func (a *API) GetAchievementGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Parse query parameters
	showInactive := param.QueryBoolOrFalse(r, "show_inactive")
	search := r.URL.Query().Get("search")

	// Create search options
	options := achievement.GroupSearchOptions{
		ShowInactive: showInactive,
		Search:       search,
	}

	// Call service
	groups, err := a.sesc.AchievementGroups(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievementGroups(groups)
	a.writeJSON(ctx, w, response)
}

// CreateAchievementGroup godoc
// @Summary Create new achievement group
// @Description Creates a new achievement group
// @Tags achievement-groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param request body param.CreateAchievementGroupRequest true "Group details"
// @Success 201 {object} respond.AchievementGroup
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-groups [post]
func (a *API) CreateAchievementGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req param.CreateAchievementGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Validate required fields
	if req.Name == "" {
		rec.Add(events.Error, achievement.ErrInvalidName)
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrInvalidName))
		return
	}

	// Create options
	options := achievement.GroupCreateOptions{
		Name:        req.Name,
		Description: req.Description,
	}

	// Call service
	group, err := a.sesc.CreateAchievementGroup(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievementGroup(group)
	a.writeJSON(ctx, w, respond.WithStatus(response, http.StatusCreated))
}

// GetAchievementTemplates godoc
// @Summary Get all achievement templates
// @Description Retrieves all achievement templates
// @Tags achievement-templates
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param show_inactive query bool false "Show inactive templates" default(false)
// @Param search query string false "Search by name"
// @Success 200 {array} respond.AchievementTemplate
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-templates [get]
func (a *API) GetAchievementTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Parse query parameters
	showInactive := param.QueryBoolOrFalse(r, "show_inactive")
	search := r.URL.Query().Get("search")

	// Create search options
	options := achievement.TemplateSearchOptions{
		ShowInactive: showInactive,
		Search:       search,
	}

	// Call service
	templates, err := a.sesc.AchievementTemplates(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievementTemplates(templates)
	a.writeJSON(ctx, w, response)
}

// CreateAchievementTemplate godoc
// @Summary Create new achievement template
// @Description Creates a new achievement template
// @Tags achievement-templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param request body param.CreateAchievementTemplateRequest true "Template details"
// @Success 201 {object} respond.AchievementTemplate
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 404 {object} respond.Error "Group not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-templates [post]
func (a *API) CreateAchievementTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req param.CreateAchievementTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Validate required fields
	if req.Name == "" {
		rec.Add(events.Error, "name is required")
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrInvalidName))
		return
	}
	if req.PointsLimit <= 0 {
		rec.Add(events.Error, "pointsLimit must be positive")
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrInvalidPointsLimit))
		return
	}
	if err := achievement.ReviewerRole(req.ReviewerRole).Validate(); err != nil {
		rec.Add(events.Error, "invalid reviewer role value")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Create options
	options := achievement.TemplateCreateOptions{
		Name:         req.Name,
		Description:  req.Description,
		PointsLimit:  req.PointsLimit,
		GroupID:      req.GroupID,
		ReviewerRole: achievement.ReviewerRole(req.ReviewerRole),
	}

	// Call service
	template, err := a.sesc.CreateAchievementTemplate(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		// Check if it's a group not found error
		if errors.Is(err, achievement.ErrAchievementGroupNotFound) {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievementTemplate(template)
	a.writeJSON(ctx, w, respond.WithStatus(response, http.StatusCreated))
}

// PatchAchievementGroup godoc
// @Summary Update achievement group
// @Description Updates an achievement group
// @Tags achievement-groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Group UUID"
// @Param request body param.PatchAchievementGroupRequest true "Group fields to update"
// @Success 200 {object} respond.AchievementGroup
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 404 {object} respond.Error "Group not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-groups/{id} [patch]
func (a *API) PatchAchievementGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		rec.Add(events.Error, "invalid group ID format")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	var req param.PatchAchievementGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Create update options
	options := achievement.GroupUpdateOptions{
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
	}

	// Call service
	group, err := a.sesc.UpdateAchievementGroup(ctx, groupID, options)
	if err != nil {
		rec.Add(events.Error, err)
		if errors.Is(err, achievement.ErrAchievementGroupNotFound) {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievementGroup(group)
	a.writeJSON(ctx, w, response)
}

// PatchAchievementTemplate godoc
// @Summary Update achievement template
// @Description Updates an achievement template
// @Tags achievement-templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "Template UUID"
// @Param request body param.PatchAchievementTemplateRequest true "Template fields to update"
// @Success 200 {object} respond.AchievementTemplate
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 404 {object} respond.Error "Template not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-templates/{id} [patch]
func (a *API) PatchAchievementTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")
	templateID, err := uuid.FromString(idStr)
	if err != nil {
		rec.Add(events.Error, "invalid template ID format")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	var req param.PatchAchievementTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Validate reviewer role if provided
	if req.ReviewerRole != nil {
		if err := achievement.ReviewerRole(*req.ReviewerRole).Validate(); err != nil {
			rec.Add(events.Error, "invalid reviewer role value")
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
	}

	// Validate points limit if provided
	if req.PointsLimit != nil && *req.PointsLimit <= 0 {
		rec.Add(events.Error, "pointsLimit must be positive")
		a.writeJSON(ctx, w, respond.WithError(ctx, achievement.ErrInvalidPointsLimit))
		return
	}

	// Create update options
	options := achievement.TemplateUpdateOptions{
		Name:         req.Name,
		Description:  req.Description,
		PointsLimit:  req.PointsLimit,
		Active:       req.Active,
		ReviewerRole: (*achievement.ReviewerRole)(req.ReviewerRole),
	}

	// Call service
	template, err := a.sesc.UpdateAchievementTemplate(ctx, templateID, options)
	if err != nil {
		rec.Add(events.Error, err)
		if errors.Is(err, achievement.ErrAchievementTemplateNotFound) {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		if errors.Is(err, achievement.ErrAchievementGroupNotFound) {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Convert to response format
	response := respond.WithAchievementTemplate(template)
	a.writeJSON(ctx, w, response)
}
