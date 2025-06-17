package achievement

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type ReviewerRole sesc.Role

func (r ReviewerRole) String() string {
	return sesc.Role(r).String()
}

func (r ReviewerRole) Name() string {
	return sesc.Role(r).Name()
}

func (r ReviewerRole) IsValid() bool {
	err := sesc.ValidateRole(int(r))
	return err == nil
}

func (r ReviewerRole) Validate() error {
	if !r.IsValid() {
		return ErrInvalidReviewerRole
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
	Name         string
	Description  string
	PointsLimit  int
	GroupID      uuid.UUID
	ReviewerRole ReviewerRole
}

// TemplateUpdateOptions contains options for updating an achievement template
type TemplateUpdateOptions struct {
	Name         *string
	Description  *string
	PointsLimit  *int
	Active       *bool
	ReviewerRole *ReviewerRole
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
