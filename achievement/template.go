package achievement

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type Kind string

const (
	Olympiad    Kind = "olympiad"
	Development Kind = "development"
	Scientific  Kind = "scientific"
)

func (k Kind) InspectorRole() sesc.Role {
	switch k {
	case Olympiad:
		return sesc.OlympiadDeputy
	case Development:
		return sesc.DevelopmentDeputy
	case Scientific:
		return sesc.ScientificDeputy
	}
	panic("wrong achievement kind")
}

func (k Kind) String() string {
	return string(k)
}

func (k Kind) IsValid() bool {
	switch k {
	case Olympiad, Development, Scientific:
		return true
	default:
		return false
	}
}

func (k Kind) Validate() error {
	if !k.IsValid() {
		return ErrInvalidAchievementKind
	}
	return nil
}

// GroupCreateOptions contains options for creating an achievement group
type GroupCreateOptions struct {
	Name        string
	Description string
}

// GroupUpdateOptions contains options for updating an achievement group
type GroupUpdateOptions struct {
	Name        *string
	Description *string
	Active      *bool
}

// TemplateCreateOptions contains options for creating an achievement template
type TemplateCreateOptions struct {
	Name        string
	Description string
	PointsLimit int
	GroupID     uuid.UUID
	Kind        Kind
}

// TemplateUpdateOptions contains options for updating an achievement template
type TemplateUpdateOptions struct {
	Name        *string
	Description *string
	PointsLimit *int
	Active      *bool
	Kind        *Kind
}

// GroupSearchOptions contains options for searching achievement groups
type GroupSearchOptions struct {
	ShowInactive bool
	Search       string
}

// TemplateSearchOptions contains options for searching achievement templates
type TemplateSearchOptions struct {
	ShowInactive bool
	Search       string
}
