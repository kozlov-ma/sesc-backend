package api

import (
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
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

var groups = []AchievementGroupResponse{
	{
		ID:          uuid.Must(uuid.NewV4()),
		Name:        "Сопровождение (подготовка/организация, проведение) мероприятий программы развития, плана работы.",
		Description: "",
		Active:      true,
	},
	{
		ID:          uuid.Must(uuid.NewV4()),
		Name:        "Обеспечение участия в мероприятиях и сопровождение обучающихся СУНЦ УрФУ",
		Description: "",
		Active:      true,
	},
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
	// TODO: Use event recorder for logging
	_ = event.Get(ctx)

	// TODO: Implement show_inactive filter
	_ = r.URL.Query().Get("show_inactive")

	// TODO: Implement search functionality
	_ = r.URL.Query().Get("search")

	// TODO: Implement actual business logic
	// For now, return mock data

	a.writeJSON(ctx, w, groups, http.StatusOK)
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
	// TODO: Use event recorder for logging
	_ = event.Get(ctx)

	var req CreateAchievementGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// TODO: Implement actual business logic
	// For now, return mock data
	group := AchievementGroupResponse{
		ID:          uuid.Must(uuid.NewV4()),
		Name:        req.Name,
		Description: req.Description,
		Active:      true,
	}

	a.writeJSON(ctx, w, group, http.StatusCreated)
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
	// TODO: Use event recorder for logging
	_ = event.Get(ctx)

	// TODO: Implement show_inactive filter
	_ = r.URL.Query().Get("show_inactive")

	// TODO: Implement search functionality
	_ = r.URL.Query().Get("search")

	// TODO: Implement actual business logic
	// For now, return mock data
	templates := []AchievementTemplateResponse{
		{
			ID:          uuid.Must(uuid.NewV4()),
			Name:        "международный уровень",
			Description: "",
			PointsLimit: 10,
			GroupID:     groups[0].ID,
			Active:      true,
			Kind:        "scientific",
		},
	}

	a.writeJSON(ctx, w, templates, http.StatusOK)
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
	// TODO: Use event recorder for logging
	_ = event.Get(ctx)

	var req CreateAchievementTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// TODO: Implement actual business logic
	// For now, return mock data
	template := AchievementTemplateResponse{
		ID:          uuid.Must(uuid.NewV4()),
		Name:        req.Name,
		Description: req.Description,
		PointsLimit: req.PointsLimit,
		GroupID:     req.GroupID,
		Active:      true,
		Kind:        req.Kind,
	}

	a.writeJSON(ctx, w, template, http.StatusCreated)
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
	// TODO: Use event recorder for logging
	_ = event.Get(ctx)

	idStr := r.PathValue("id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(ctx, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid group ID format",
			RuMessage: "Некорректный формат ID группы",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	var req PatchAchievementGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// TODO: Implement actual business logic
	// For now, return mock data
	group := AchievementGroupResponse{
		ID:          groupID,
		Name:        "Updated Group",
		Description: "Updated Description",
		Active:      true,
	}

	a.writeJSON(ctx, w, group, http.StatusOK)
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
	// TODO: Use event recorder for logging
	_ = event.Get(ctx)

	idStr := r.PathValue("id")
	templateID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(ctx, w, InvalidUUIDError{
			Code:      "INVALID_UUID",
			Message:   "Invalid template ID format",
			RuMessage: "Некорректный формат ID шаблона",
		}.WithStatus(http.StatusBadRequest))
		return
	}

	var req PatchAchievementTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(ctx, w, ErrInvalidRequest.WithStatus(http.StatusBadRequest))
		return
	}

	// TODO: Implement actual business logic
	// For now, return mock data
	template := AchievementTemplateResponse{
		ID:          templateID,
		Name:        "Updated Template",
		Description: "Updated Description",
		PointsLimit: 15,
		GroupID:     uuid.Must(uuid.NewV4()),
		Active:      true,
		Kind:        "scientific",
	}

	a.writeJSON(ctx, w, template, http.StatusOK)
}
