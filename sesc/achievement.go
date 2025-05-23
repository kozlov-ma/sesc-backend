package sesc

import (
	"github.com/gofrs/uuid/v5"
)

// AchievementGroup represents a group of related achievement templates
type AchievementGroup struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
}

// AchievementTemplate represents a template for creating achievements
type AchievementTemplate struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PointsLimit int32     `json:"pointsLimit"`
	GroupID     uuid.UUID `json:"groupId"`
	Active      bool      `json:"active"`
	Kind        string    `json:"kind"` // olympiad, development, scientific
}

// AchievementGroupCreateOptions contains options for creating an achievement group
type AchievementGroupCreateOptions struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AchievementGroupUpdateOptions contains options for updating an achievement group
type AchievementGroupUpdateOptions struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

// AchievementTemplateCreateOptions contains options for creating an achievement template
type AchievementTemplateCreateOptions struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PointsLimit int32     `json:"pointsLimit"`
	GroupID     uuid.UUID `json:"groupId"`
	Kind        string    `json:"kind"`
}

// AchievementTemplateUpdateOptions contains options for updating an achievement template
type AchievementTemplateUpdateOptions struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	PointsLimit *int32     `json:"pointsLimit,omitempty"`
	GroupID     *uuid.UUID `json:"groupId,omitempty"`
	Active      *bool      `json:"active,omitempty"`
	Kind        *string    `json:"kind,omitempty"`
}

// AchievementGroupSearchOptions contains options for searching achievement groups
type AchievementGroupSearchOptions struct {
	ShowInactive bool   `json:"showInactive"`
	Search       string `json:"search"`
}

// AchievementTemplateSearchOptions contains options for searching achievement templates
type AchievementTemplateSearchOptions struct {
	ShowInactive bool   `json:"showInactive"`
	Search       string `json:"search"`
}
