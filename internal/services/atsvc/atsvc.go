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
// Returns achievement.ErrAchievementGroupNotFound if the group does not exist.
func (s *ATS) AchievementGroupByID(ctx context.Context, id uuid.UUID) (*ent.AchievementGroup, error) {
	rec := event.Get(ctx).Sub("sesc/achievement_group_by_id")
	rec.Add("group_id", id)

	group, err := s.client.AchievementGroup.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			rec.Add(events.Error, "achievement group not found")
			return nil, achievement.ErrAchievementGroupNotFound
		}
		rec.Add(events.Error, fmt.Errorf("failed to get achievement group: %w", err))
		return nil, err
	}

	return group, nil
}

// CreateAchievementGroup creates a new achievement group with auto-generated ID.
func (s *ATS) CreateAchievementGroup(
	ctx context.Context,
	options achievement.GroupCreateOptions,
) (*ent.AchievementGroup, error) {
	rec := event.Get(ctx).Sub("sesc/create_achievement_group")
	rec.Add("group_name", options.Name)

	group, err := s.client.AchievementGroup.Create().
		SetName(options.Name).
		SetDescription(options.Description).
		Save(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to create achievement group: %w", err))
		return nil, err
	}

	rec.Add("created_group_id", group.ID)
	return group, nil
}

// UpdateAchievementGroup updates an existing achievement group.
// Returns achievement.ErrAchievementGroupNotFound if the group does not exist.
func (s *ATS) UpdateAchievementGroup(
	ctx context.Context,
	id uuid.UUID,
	options achievement.GroupUpdateOptions,
) (*ent.AchievementGroup, error) {
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
		return nil, achievement.ErrAchievementGroupNotFound
	case err != nil:
		rec.Add(events.Error, fmt.Errorf("failed to update achievement group: %w", err))
		return nil, err
	}

	rec.Add("updated_group_id", group.ID)
	return group, nil
}

// AchievementTemplates gets all achievement templates with optional filtering.
func (s *ATS) AchievementTemplates(
	ctx context.Context,
	options achievement.TemplateSearchOptions,
) (ent.AchievementTemplates, error) {
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

	rec.Add("templates_count", len(templates))
	return templates, nil
}

// AchievementTemplateByID gets an achievement template by its ID.
// Returns achievement.ErrAchievementTemplateNotFound if the template does not exist.
func (s *ATS) AchievementTemplateByID(ctx context.Context, id uuid.UUID) (*ent.AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("sesc/achievement_template_by_id")
	rec.Add("template_id", id)

	template, err := s.client.AchievementTemplate.Get(ctx, id)
	switch {
	case ent.IsNotFound(err):
		rec.Add(events.Error, "achievement template not found")
		return nil, achievement.ErrAchievementTemplateNotFound
	case err != nil:
		rec.Add(events.Error, fmt.Errorf("failed to get achievement template: %w", err))
		return nil, err
	}

	return template, nil
}

// CreateAchievementTemplate creates a new achievement template with auto-generated ID.
// Returns achievement.ErrAchievementGroupNotFound if the specified group does not exist.
func (s *ATS) CreateAchievementTemplate(
	ctx context.Context,
	options achievement.TemplateCreateOptions,
) (*ent.AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("sesc/create_achievement_template")
	rec.Add("template_name", options.Name)
	rec.Add("group_id", options.GroupID)

	// Validate the kind
	if err := options.ReviewerRole.ValidateReviewer(); err != nil {
		rec.Add(events.Error, fmt.Errorf("invalid reviewer role: %w", err))
		return nil, err
	}

	template, err := s.client.AchievementTemplate.Create().
		SetName(options.Name).
		SetDescription(options.Description).
		SetPointsLimit(options.PointsLimit).
		SetGroupID(options.GroupID).
		SetReviewerRole(options.ReviewerRole).
		Save(ctx)
	switch {
	case ent.IsConstraintError(err):
		rec.Add(events.Error, "achievement group not found")
		return nil, achievement.ErrAchievementGroupNotFound
	case err != nil:
		rec.Add(events.Error, fmt.Errorf("failed to create achievement template: %w", err))
		return nil, err
	}

	rec.Add("created_template_id", template.ID)
	return template, nil
}

// UpdateAchievementTemplate updates an existing achievement template.
// Returns achievement.ErrAchievementTemplateNotFound if the template does not exist.
// Note: GroupID cannot be changed after creation.
func (s *ATS) UpdateAchievementTemplate(
	ctx context.Context,
	id uuid.UUID,
	options achievement.TemplateUpdateOptions,
) (*ent.AchievementTemplate, error) {
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
	if options.ReviewerRole != nil {
		// Validate the kind
		if err := options.ReviewerRole.ValidateReviewer(); err != nil {
			rec.Add(events.Error, fmt.Errorf("invalid achievement kind: %w", err))
			return nil, err
		}
		update = update.SetReviewerRole(*options.ReviewerRole)
	}

	template, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			rec.Add(events.Error, "achievement template not found")
			return nil, achievement.ErrAchievementTemplateNotFound
		}
		rec.Add(events.Error, fmt.Errorf("failed to update achievement template: %w", err))
		return nil, err
	}

	rec.Add("updated_template_id", template.ID)
	return template, nil
}
