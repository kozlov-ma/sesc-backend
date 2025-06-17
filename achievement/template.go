package achievement

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

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
	Name         string
	Description  string
	PointsLimit  int
	GroupID      uuid.UUID
	ReviewerRole sesc.Role
}

// TemplateUpdateOptions contains options for updating an achievement template
type TemplateUpdateOptions struct {
	Name         *string
	Description  *string
	PointsLimit  *int
	Active       *bool
	ReviewerRole *sesc.Role
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
