package respond

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
)

// AchievementGroup represents an achievement group response
type AchievementGroup struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name        string    `json:"name"        example:"Научная деятельность"                 validate:"required"`
	Description string    `json:"description" example:"Достижения в научной деятельности"    validate:"required"`
	Active      bool      `json:"active"      example:"true"                                 validate:"required"`
}

// AchievementTemplate represents an achievement template response
type AchievementTemplate struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name        string    `json:"name"        example:"Публикация в журнале"                 validate:"required"`
	Description string    `json:"description" example:"Публикация статьи в научном журнале"  validate:"required"`
	PointsLimit int       `json:"pointsLimit" example:"10"                                   validate:"required"`
	GroupID     uuid.UUID `json:"groupId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Active      bool      `json:"active"      example:"true"                                 validate:"required"`
	Kind        string    `json:"kind"        example:"scientific"                           validate:"required" enums:"olympiad,development,scientific"`
}

// WithAchievementGroup converts an ent.AchievementGroup to a response
func WithAchievementGroup(group *ent.AchievementGroup) AchievementGroup {
	return AchievementGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}
}

// WithAchievementGroups converts a slice of ent.AchievementGroup to responses
func WithAchievementGroups(groups ent.AchievementGroups) []AchievementGroup {
	response := make([]AchievementGroup, 0, len(groups))
	for _, group := range groups {
		response = append(response, WithAchievementGroup(group))
	}
	return response
}

// WithAchievementTemplate converts an ent.AchievementTemplate to a response
func WithAchievementTemplate(template *ent.AchievementTemplate) AchievementTemplate {
	return AchievementTemplate{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		PointsLimit: template.PointsLimit,
		GroupID:     template.GroupID,
		Active:      template.Active,
		Kind:        template.Kind.String(),
	}
}

// WithAchievementTemplates converts a slice of ent.AchievementTemplate to responses
func WithAchievementTemplates(templates ent.AchievementTemplates) []AchievementTemplate {
	response := make([]AchievementTemplate, 0, len(templates))
	for _, template := range templates {
		response = append(response, WithAchievementTemplate(template))
	}
	return response
}
