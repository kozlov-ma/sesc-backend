package sesc

import (
	"github.com/gofrs/uuid/v5"
)

// AchievementGroup represents a group of related achievement templates
type AchievementGroup struct {
	ID          uuid.UUID
	Name        string
	Description string
	Active      bool
}

// AchievementTemplate represents a template for creating achievements
type AchievementTemplate struct {
	ID          uuid.UUID
	Name        string
	Description string
	PointsLimit int
	GroupID     uuid.UUID
	Active      bool
	Kind        string // olympiad, development, scientific
}

// AchievementGroupCreateOptions contains options for creating an achievement group
type AchievementGroupCreateOptions struct {
	Name        string
	Description string
}

// AchievementGroupUpdateOptions contains options for updating an achievement group
type AchievementGroupUpdateOptions struct {
	Name        *string
	Description *string
	Active      *bool
}

// AchievementTemplateCreateOptions contains options for creating an achievement template
type AchievementTemplateCreateOptions struct {
	Name        string
	Description string
	PointsLimit int
	GroupID     uuid.UUID
	Kind        string
}

// AchievementTemplateUpdateOptions contains options for updating an achievement template
type AchievementTemplateUpdateOptions struct {
	Name        *string
	Description *string
	PointsLimit *int
	Active      *bool
	Kind        *string
}

// AchievementGroupSearchOptions contains options for searching achievement groups
type AchievementGroupSearchOptions struct {
	ShowInactive bool
	Search       string
}

// AchievementTemplateSearchOptions contains options for searching achievement templates
type AchievementTemplateSearchOptions struct {
	ShowInactive bool
	Search       string
}
