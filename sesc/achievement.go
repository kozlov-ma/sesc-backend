package sesc

import (
	"github.com/gofrs/uuid/v5"
)

// AchievementKind represents the type of achievement
type AchievementKind string

const (
	Olympiad    AchievementKind = "olympiad"
	Development AchievementKind = "development"
	Scientific  AchievementKind = "scientific"
)

// String implements the Stringer interface
func (k AchievementKind) String() string {
	return string(k)
}

// IsValid checks if the AchievementKind is one of the valid values
func (k AchievementKind) IsValid() bool {
	switch k {
	case Olympiad, Development, Scientific:
		return true
	default:
		return false
	}
}

// Validate returns an error if the AchievementKind is not valid
func (k AchievementKind) Validate() error {
	if !k.IsValid() {
		return ErrInvalidAchievementKind
	}
	return nil
}

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
	Kind        AchievementKind
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
	Kind        AchievementKind
}

// AchievementTemplateUpdateOptions contains options for updating an achievement template
type AchievementTemplateUpdateOptions struct {
	Name        *string
	Description *string
	PointsLimit *int
	Active      *bool
	Kind        *AchievementKind
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
