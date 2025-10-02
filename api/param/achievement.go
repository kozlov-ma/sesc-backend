package param

import (
	"github.com/gofrs/uuid/v5"
)

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
	Action  string `json:"action"  example:"approve"                       validate:"required,oneof=approve disapprove request_changes"`
	Comment string `json:"comment" example:"Good job, but could be better" validate:"omitempty"`
}

// UpdateAchievementPointsRequest represents the request to update achievement points
type UpdateAchievementPointsRequest struct {
	Points  int    `json:"points"  example:"8"                         validate:"required,min=0"`
	Comment string `json:"comment" example:"Updated based on feedback" validate:"omitempty"`
}

// CreateAchievementGroupRequest represents the request to create an achievement group
type CreateAchievementGroupRequest struct {
	Name        string `json:"name"        example:"Научная деятельность"              validate:"required"`
	Description string `json:"description" example:"Достижения в научной деятельности" validate:"required"`
}

// CreateAchievementTemplateRequest represents the request to create an achievement template
type CreateAchievementTemplateRequest struct {
	Name         string    `json:"name"         example:"Публикация в журнале"                 validate:"required"`
	Description  string    `json:"description"  example:"Публикация статьи в научном журнале"  validate:"required"`
	PointsLimit  int       `json:"pointsLimit"  example:"10"                                   validate:"required"`
	GroupID      uuid.UUID `json:"groupId"      example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	ReviewerRole int       `json:"reviewerRole" example:"3"                                    validate:"required"`
}

// PatchAchievementGroupRequest represents the request to update an achievement group
type PatchAchievementGroupRequest struct {
	Name        *string `json:"name,omitzero"        example:"Научная деятельность"`
	Description *string `json:"description,omitzero" example:"Достижения в научной деятельности"`
	Active      *bool   `json:"active,omitzero"      example:"true"`
}

// PatchAchievementTemplateRequest represents the request to update an achievement template
type PatchAchievementTemplateRequest struct {
	Name         *string `json:"name,omitzero"        example:"Публикация в журнале"`
	Description  *string `json:"description,omitzero" example:"Публикация статьи в научном журнале"`
	PointsLimit  *int    `json:"pointsLimit,omitzero" example:"10"`
	Active       *bool   `json:"active,omitzero"      example:"true"`
	ReviewerRole *int    `json:"reviewerRole"         example:"3"`
}
