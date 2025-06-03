package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type AchievementGroupResponse struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name        string    `json:"name"        example:"Научная деятельность"                 validate:"required"`
	Description string    `json:"description" example:"Достижения в научной деятельности"    validate:"required"`
	Active      bool      `json:"active"      example:"true"                                 validate:"required"`
}

type AchievementTemplateResponse struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name        string    `json:"name"        example:"Публикация в журнале"                 validate:"required"`
	Description string    `json:"description" example:"Публикация статьи в научном журнале"  validate:"required"`
	PointsLimit int       `json:"pointsLimit" example:"10"                                   validate:"required"`
	GroupID     uuid.UUID `json:"groupId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Active      bool      `json:"active"      example:"true"                                 validate:"required"`
	Kind        string    `json:"kind"        example:"scientific"                           validate:"required" enums:"olympiad,development,scientific"`
}

type CreateAchievementGroupRequest struct {
	Name        string `json:"name"        example:"Научная деятельность"              validate:"required"`
	Description string `json:"description" example:"Достижения в научной деятельности" validate:"required"`
}

type CreateAchievementTemplateRequest struct {
	Name        string    `json:"name"        example:"Публикация в журнале"                 validate:"required"`
	Description string    `json:"description" example:"Публикация статьи в научном журнале"  validate:"required"`
	PointsLimit int       `json:"pointsLimit" example:"10"                                   validate:"required"`
	GroupID     uuid.UUID `json:"groupId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Kind        string    `json:"kind"        example:"scientific"                           validate:"required,oneof=olympiad development scientific"`
}

type PatchAchievementGroupRequest struct {
	Name        *string `json:"name,omitzero"        example:"Научная деятельность"`
	Description *string `json:"description,omitzero" example:"Достижения в научной деятельности"`
	Active      *bool   `json:"active,omitzero"      example:"true"`
}

type PatchAchievementTemplateRequest struct {
	Name        *string `json:"name,omitzero"        example:"Публикация в журнале"`
	Description *string `json:"description,omitzero" example:"Публикация статьи в научном журнале"`
	PointsLimit *int    `json:"pointsLimit,omitzero" example:"10"`
	Active      *bool   `json:"active,omitzero"      example:"true"`
	Kind        *string `json:"kind,omitzero"                                                      validate:"omitempty,oneof=olympiad development scientific"`
}

// GetAchievementGroups godoc
// @Summary Get all achievement groups
// @Description Retrieves all achievement groups
// @Tags achievement-groups
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param show_inactive query bool false "Show inactive groups" default(false)
// @Param search query string false "Search by name"
// @Success 200 {array} AchievementGroupResponse
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-groups [get]
func (a *API) GetAchievementGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Parse query parameters
	showInactiveStr := r.URL.Query().Get("show_inactive")
	showInactive := false
	if showInactiveStr != "" {
		var err error
		showInactive, err = strconv.ParseBool(showInactiveStr)
		if err != nil {
			rec.Add(events.Error, "invalid show_inactive parameter")
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
	}

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
	response := make([]AchievementGroupResponse, 0, len(groups))
	for _, group := range groups {
		response = append(response, AchievementGroupResponse{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			Active:      group.Active,
		})
	}

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
// @Param request body CreateAchievementGroupRequest true "Group details"
// @Success 201 {object} AchievementGroupResponse
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-groups [post]
func (a *API) CreateAchievementGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateAchievementGroupRequest
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
	response := AchievementGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}

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
// @Success 200 {array} AchievementTemplateResponse
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-templates [get]
func (a *API) GetAchievementTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Parse query parameters
	showInactiveStr := r.URL.Query().Get("show_inactive")
	showInactive := false
	if showInactiveStr != "" {
		var err error
		showInactive, err = strconv.ParseBool(showInactiveStr)
		if err != nil {
			rec.Add(events.Error, "invalid show_inactive parameter")
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
	}

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
	response := make([]AchievementTemplateResponse, 0, len(templates))
	for _, template := range templates {
		response = append(response, AchievementTemplateResponse{
			ID:          template.ID,
			Name:        template.Name,
			Description: template.Description,
			PointsLimit: template.PointsLimit,
			GroupID:     template.GroupID,
			Active:      template.Active,
			Kind:        template.Kind.String(),
		})
	}

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
// @Param request body CreateAchievementTemplateRequest true "Template details"
// @Success 201 {object} AchievementTemplateResponse
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden - admin role required"
// @Failure 404 {object} respond.Error "Group not found"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /achievement-templates [post]
func (a *API) CreateAchievementTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateAchievementTemplateRequest
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
	if err := achievement.Kind(req.Kind).Validate(); err != nil {
		rec.Add(events.Error, "invalid kind value")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Create options
	options := achievement.TemplateCreateOptions{
		Name:        req.Name,
		Description: req.Description,
		PointsLimit: req.PointsLimit,
		GroupID:     req.GroupID,
		Kind:        achievement.Kind(req.Kind),
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
	response := AchievementTemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		PointsLimit: template.PointsLimit,
		GroupID:     template.GroupID,
		Active:      template.Active,
		Kind:        template.Kind.String(),
	}

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
// @Param request body PatchAchievementGroupRequest true "Group fields to update"
// @Success 200 {object} AchievementGroupResponse
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

	var req PatchAchievementGroupRequest
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
	response := AchievementGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}

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
// @Param request body PatchAchievementTemplateRequest true "Template fields to update"
// @Success 200 {object} AchievementTemplateResponse
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

	var req PatchAchievementTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Validate kind if provided
	if req.Kind != nil {
		if err := achievement.Kind(*req.Kind).Validate(); err != nil {
			rec.Add(events.Error, "invalid kind value")
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
		Name:        req.Name,
		Description: req.Description,
		PointsLimit: req.PointsLimit,
		Active:      req.Active,
		Kind:        (*achievement.Kind)(req.Kind),
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
	response := AchievementTemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		PointsLimit: template.PointsLimit,
		GroupID:     template.GroupID,
		Active:      template.Active,
		Kind:        template.Kind.String(),
	}

	a.writeJSON(ctx, w, response)
}
