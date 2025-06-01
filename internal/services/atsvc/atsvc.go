package atsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementtemplate"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	UUID                             = uuid.UUID
	User                             = sesc.User
	Department                       = sesc.Department
	Role                             = sesc.Role
	UserUpdateOptions                = sesc.UserUpdateOptions
	AchievementGroup                 = achievement.Group
	AchievementTemplate              = achievement.Template
	AchievementGroupCreateOptions    = achievement.GroupCreateOptions
	AchievementGroupUpdateOptions    = achievement.GroupUpdateOptions
	AchievementGroupSearchOptions    = achievement.GroupSearchOptions
	AchievementTemplateCreateOptions = achievement.TemplateCreateOptions
	AchievementTemplateUpdateOptions = achievement.TemplateUpdateOptions
	AchievementTemplateSearchOptions = achievement.TemplateSearchOptions
)

type ATS struct {
	client *ent.Client
}

func New(client *ent.Client) *ATS {
	return &ATS{
		client: client,
	}
}

// AchievementGroupByID gets an achievement group by its ID.
// Returns sesc.ErrAchievementGroupNotFound if the group does not exist.
func (s *ATS) AchievementGroupByID(ctx context.Context, id UUID) (AchievementGroup, error) {
	rec := event.Get(ctx).Sub("sesc/achievement_group_by_id")
	rec.Add("group_id", id)

	group, err := s.client.AchievementGroup.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			rec.Add(events.Error, "achievement group not found")
			return AchievementGroup{}, achievement.ErrAchievementGroupNotFound
		}
		rec.Add(events.Error, fmt.Errorf("failed to get achievement group: %w", err))
		return AchievementGroup{}, err
	}

	return AchievementGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}, nil
}

// CreateAchievementGroup creates a new achievement group with auto-generated ID.
func (s *ATS) CreateAchievementGroup(
	ctx context.Context,
	options AchievementGroupCreateOptions,
) (AchievementGroup, error) {
	rec := event.Get(ctx).Sub("sesc/create_achievement_group")
	rec.Add("group_name", options.Name)

	group, err := s.client.AchievementGroup.Create().
		SetName(options.Name).
		SetDescription(options.Description).
		Save(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to create achievement group: %w", err))
		return AchievementGroup{}, err
	}

	result := AchievementGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}

	rec.Add("created_group", result)
	return result, nil
}

// UpdateAchievementGroup updates an existing achievement group.
// Returns sesc.ErrAchievementGroupNotFound if the group does not exist.
func (s *ATS) UpdateAchievementGroup(
	ctx context.Context,
	id UUID,
	options AchievementGroupUpdateOptions,
) (AchievementGroup, error) {
	rec := event.Get(ctx).Sub("sesc/update_achievement_group")
	rec.Add("group_id", id)

	update := s.client.AchievementGroup.UpdateOneID(id)

	if options.Name != nil {
		update = update.SetName(*options.Name)
	}
	if options.Description != nil {
		update = update.SetDescription(*options.Description)
	}
	if options.Active != nil {
		update = update.SetActive(*options.Active)
	}

	group, err := update.Save(ctx)
	switch {
	case ent.IsNotFound(err):
		rec.Add(events.Error, "achievement group not found")
		return AchievementGroup{}, achievement.ErrAchievementGroupNotFound
	case err != nil:
		rec.Add(events.Error, fmt.Errorf("failed to update achievement group: %w", err))
		return AchievementGroup{}, err
	}

	result := AchievementGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Active:      group.Active,
	}

	rec.Add("updated_group", result)
	return result, nil
}

// AchievementTemplates gets all achievement templates with optional filtering.
func (s *ATS) AchievementTemplates(
	ctx context.Context,
	options AchievementTemplateSearchOptions,
) ([]AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("sesc/achievement_templates")

	query := s.client.AchievementTemplate.Query()

	// Apply filters
	if !options.ShowInactive {
		query = query.Where(achievementtemplate.Active(true))
	}

	if options.Search != "" {
		searchTerm := strings.ToLower(options.Search)
		query = query.Where(achievementtemplate.Or(
			achievementtemplate.NameContainsFold(searchTerm),
			achievementtemplate.DescriptionContainsFold(searchTerm),
		))
	}

	templates, err := query.All(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement templates: %w", err))
		return nil, err
	}

	result := make([]AchievementTemplate, 0, len(templates))
	for _, t := range templates {
		result = append(result, AchievementTemplate{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			PointsLimit: t.PointsLimit,
			GroupID:     t.GroupID,
			Active:      t.Active,
			Kind:        t.Kind,
		})
	}

	rec.Add("templates_count", len(result))
	return result, nil
}

// AchievementTemplateByID gets an achievement template by its ID.
// Returns sesc.ErrAchievementTemplateNotFound if the template does not exist.
func (s *ATS) AchievementTemplateByID(ctx context.Context, id UUID) (AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("sesc/achievement_template_by_id")
	rec.Add("template_id", id)

	template, err := s.client.AchievementTemplate.Get(ctx, id)
	switch {
	case ent.IsNotFound(err):
		rec.Add(events.Error, "achievement template not found")
		return AchievementTemplate{}, achievement.ErrAchievementTemplateNotFound
	case err != nil:
		rec.Add(events.Error, fmt.Errorf("failed to get achievement template: %w", err))
		return AchievementTemplate{}, err
	}

	return AchievementTemplate{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		PointsLimit: template.PointsLimit,
		GroupID:     template.GroupID,
		Active:      template.Active,
		Kind:        template.Kind,
	}, nil
}

// CreateAchievementTemplate creates a new achievement template with auto-generated ID.
// Returns sesc.ErrAchievementGroupNotFound if the specified group does not exist.
func (s *ATS) CreateAchievementTemplate(
	ctx context.Context,
	options AchievementTemplateCreateOptions,
) (AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("sesc/create_achievement_template")
	rec.Add("template_name", options.Name)
	rec.Add("group_id", options.GroupID)

	// Validate the kind
	if err := options.Kind.Validate(); err != nil {
		rec.Add(events.Error, fmt.Errorf("invalid achievement kind: %w", err))
		return AchievementTemplate{}, err
	}

	template, err := s.client.AchievementTemplate.Create().
		SetName(options.Name).
		SetDescription(options.Description).
		SetPointsLimit(options.PointsLimit).
		SetGroupID(options.GroupID).
		SetKind(options.Kind).
		Save(ctx)
	switch {
	case ent.IsConstraintError(err):
		rec.Add(events.Error, "achievement group not found")
		return AchievementTemplate{}, achievement.ErrAchievementGroupNotFound
	case err != nil:
		rec.Add(events.Error, fmt.Errorf("failed to create achievement template: %w", err))
		return AchievementTemplate{}, err
	}

	result := AchievementTemplate{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		PointsLimit: template.PointsLimit,
		GroupID:     template.GroupID,
		Active:      template.Active,
		Kind:        template.Kind,
	}

	rec.Add("created_template", result)
	return result, nil
}

// UpdateAchievementTemplate updates an existing achievement template.
// Returns sesc.ErrAchievementTemplateNotFound if the template does not exist.
// Note: GroupID cannot be changed after creation.
func (s *ATS) UpdateAchievementTemplate(
	ctx context.Context,
	id UUID,
	options AchievementTemplateUpdateOptions,
) (AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("sesc/update_achievement_template")
	rec.Add("template_id", id)

	update := s.client.AchievementTemplate.UpdateOneID(id)

	if options.Name != nil {
		update = update.SetName(*options.Name)
	}
	if options.Description != nil {
		update = update.SetDescription(*options.Description)
	}
	if options.PointsLimit != nil {
		update = update.SetPointsLimit(*options.PointsLimit)
	}
	if options.Active != nil {
		update = update.SetActive(*options.Active)
	}
	if options.Kind != nil {
		// Validate the kind
		if err := options.Kind.Validate(); err != nil {
			rec.Add(events.Error, fmt.Errorf("invalid achievement kind: %w", err))
			return AchievementTemplate{}, err
		}
		update = update.SetKind(*options.Kind)
	}

	template, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			rec.Add(events.Error, "achievement template not found")
			return AchievementTemplate{}, achievement.ErrAchievementTemplateNotFound
		}
		rec.Add(events.Error, fmt.Errorf("failed to update achievement template: %w", err))
		return AchievementTemplate{}, err
	}

	result := AchievementTemplate{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		PointsLimit: template.PointsLimit,
		GroupID:     template.GroupID,
		Active:      template.Active,
		Kind:        template.Kind,
	}

	rec.Add("updated_template", result)
	return result, nil
}
