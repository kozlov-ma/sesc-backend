package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
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
	PointsLimit int32     `json:"pointsLimit" example:"10"                                   validate:"required"`
	GroupID     uuid.UUID `json:"groupId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Active      bool      `json:"active"      example:"true"                                 validate:"required"`
	Kind        string    `json:"kind"        example:"scientific"                           validate:"required,oneof=olympiad development scientific"`
}

type CreateAchievementGroupRequest struct {
	Name        string `json:"name"        example:"Научная деятельность"              validate:"required"`
	Description string `json:"description" example:"Достижения в научной деятельности" validate:"required"`
}

type CreateAchievementTemplateRequest struct {
	Name        string    `json:"name"        example:"Публикация в журнале"                 validate:"required"`
	Description string    `json:"description" example:"Публикация статьи в научном журнале"  validate:"required"`
	PointsLimit int32     `json:"pointsLimit" example:"10"                                   validate:"required"`
	GroupID     uuid.UUID `json:"groupId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Kind        string    `json:"kind"        example:"scientific"                           validate:"required,oneof=olympiad development scientific"`
}

type PatchAchievementGroupRequest struct {
	Name        *string `json:"name,omitzero"        example:"Научная деятельность"`
	Description *string `json:"description,omitzero" example:"Достижения в научной деятельности"`
	Active      *bool   `json:"active,omitzero"      example:"true"`
}

type PatchAchievementTemplateRequest struct {
	Name        *string    `json:"name,omitzero"        example:"Публикация в журнале"`
	Description *string    `json:"description,omitzero" example:"Публикация статьи в научном журнале"`
	PointsLimit *int32     `json:"pointsLimit,omitzero" example:"10"`
	GroupID     *uuid.UUID `json:"groupId,omitzero"     example:"550e8400-e29b-41d4-a716-446655440000"`
	Active      *bool      `json:"active,omitzero"      example:"true"`
	Kind        *string    `json:"kind,omitzero"        example:"scientific"                           validate:"omitempty,oneof=olympiad development scientific"`
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
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 500 {object} ServerError "Internal server error"
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
			writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
			return
		}
	}

	search := r.URL.Query().Get("search")

	// Create search options
	options := sesc.AchievementGroupSearchOptions{
		ShowInactive: showInactive,
		Search:       search,
	}

	// Call service
	groups, err := a.sesc.AchievementGroups(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
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

	a.writeJSON(ctx, w, response, http.StatusOK)
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
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /achievement-groups [post]
func (a *API) CreateAchievementGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateAchievementGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Validate required fields
	if req.Name == "" {
		rec.Add(events.Error, "name is required")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Create options
	options := sesc.AchievementGroupCreateOptions{
		Name:        req.Name,
		Description: req.Description,
	}

	// Call service
	group, err := a.sesc.CreateAchievementGroup(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
		return
	}

	// Convert to response format
	response := AchievementGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}

	a.writeJSON(ctx, w, response, http.StatusCreated)
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
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 500 {object} ServerError "Internal server error"
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
			writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
			return
		}
	}

	search := r.URL.Query().Get("search")

	// Create search options
	options := sesc.AchievementTemplateSearchOptions{
		ShowInactive: showInactive,
		Search:       search,
	}

	// Call service
	templates, err := a.sesc.AchievementTemplates(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
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
			Kind:        template.Kind,
		})
	}

	a.writeJSON(ctx, w, response, http.StatusOK)
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
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} GroupNotFoundError "Group not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /achievement-templates [post]
func (a *API) CreateAchievementTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var req CreateAchievementTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Validate required fields
	if req.Name == "" {
		rec.Add(events.Error, "name is required")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}
	if req.PointsLimit <= 0 {
		rec.Add(events.Error, "pointsLimit must be positive")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}
	if req.Kind != "olympiad" && req.Kind != "development" && req.Kind != "scientific" {
		rec.Add(events.Error, "invalid kind value")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Create options
	options := sesc.AchievementTemplateCreateOptions{
		Name:        req.Name,
		Description: req.Description,
		PointsLimit: req.PointsLimit,
		GroupID:     req.GroupID,
		Kind:        req.Kind,
	}

	// Call service
	template, err := a.sesc.CreateAchievementTemplate(ctx, options)
	if err != nil {
		rec.Add(events.Error, err)
		// Check if it's a group not found error
		if errors.Is(err, sesc.ErrAchievementGroupNotFound) {
			writeError(ctx, w, GroupNotFoundError{
				Code:      "GROUP_NOT_FOUND",
				Message:   "Achievement group not found",
				RuMessage: "Группа достижений не найдена",
			}.WithStatus(http.StatusNotFound))
			return
		}
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
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
		Kind:        template.Kind,
	}

	a.writeJSON(ctx, w, response, http.StatusCreated)
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
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} GroupNotFoundError "Group not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /achievement-groups/{id} [patch]
func (a *API) PatchAchievementGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		rec.Add(events.Error, "invalid group ID format")
		writeError(ctx, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid group ID format",
			RuMessage: "Некорректный формат ID группы",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	var req PatchAchievementGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Create update options
	options := sesc.AchievementGroupUpdateOptions{
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
	}

	// Call service
	group, err := a.sesc.UpdateAchievementGroup(ctx, groupID, options)
	if err != nil {
		rec.Add(events.Error, err)
		if errors.Is(err, sesc.ErrAchievementGroupNotFound) {
			writeError(ctx, w, GroupNotFoundError{
				Code:      "GROUP_NOT_FOUND",
				Message:   "Achievement group not found",
				RuMessage: "Группа достижений не найдена",
			}.WithStatus(http.StatusNotFound))
			return
		}
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
		return
	}

	// Convert to response format
	response := AchievementGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}

	a.writeJSON(ctx, w, response, http.StatusOK)
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
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} AchievementTemplateNotFoundError "Template not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /achievement-templates/{id} [patch]
func (a *API) PatchAchievementTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	idStr := r.PathValue("id")
	templateID, err := uuid.FromString(idStr)
	if err != nil {
		rec.Add(events.Error, "invalid template ID format")
		writeError(ctx, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid template ID format",
			RuMessage: "Некорректный формат ID шаблона",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	var req PatchAchievementTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rec.Add(events.Error, "invalid request body")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Validate kind if provided
	if req.Kind != nil && *req.Kind != "olympiad" && *req.Kind != "development" && *req.Kind != "scientific" {
		rec.Add(events.Error, "invalid kind value")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Validate points limit if provided
	if req.PointsLimit != nil && *req.PointsLimit <= 0 {
		rec.Add(events.Error, "pointsLimit must be positive")
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// Create update options
	options := sesc.AchievementTemplateUpdateOptions{
		Name:        req.Name,
		Description: req.Description,
		PointsLimit: req.PointsLimit,
		GroupID:     req.GroupID,
		Active:      req.Active,
		Kind:        req.Kind,
	}

	// Call service
	template, err := a.sesc.UpdateAchievementTemplate(ctx, templateID, options)
	if err != nil {
		rec.Add(events.Error, err)
		if errors.Is(err, sesc.ErrAchievementTemplateNotFound) {
			writeError(ctx, w, AchievementTemplateNotFoundError{
				Code:      "ACHIEVEMENT_TEMPLATE_NOT_FOUND",
				Message:   "Achievement template not found",
				RuMessage: "Шаблон достижения не найден",
			}.WithStatus(http.StatusNotFound))
			return
		}
		if errors.Is(err, sesc.ErrAchievementGroupNotFound) {
			writeError(ctx, w, GroupNotFoundError{
				Code:      "GROUP_NOT_FOUND",
				Message:   "Achievement group not found",
				RuMessage: "Группа достижений не найдена",
			}.WithStatus(http.StatusNotFound))
			return
		}
		writeError(ctx, w, ErrServerError.WithStatus(http.StatusInternalServerError))
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
		Kind:        template.Kind,
	}

	a.writeJSON(ctx, w, response, http.StatusOK)
}
