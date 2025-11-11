package atsvc

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// AchievementGroups gets all achievement groups with access control.
func (s *ATS) AchievementGroups(
	ctx context.Context,
	actingUser company.User,
	options achievement.GroupSearchOptions,
) (ent.AchievementGroups, error) {
	rec := event.Get(ctx).Sub("atsvc/achievement_groups_with_access")

	action := NewViewAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.achievementGroups(ctx, options)
}

// AchievementGroupByID gets an achievement group by ID with access control.
func (s *ATS) AchievementGroupByID(
	ctx context.Context,
	actingUser company.User,
	id uuid.UUID,
) (*ent.AchievementGroup, error) {
	rec := event.Get(ctx).Sub("atsvc/achievement_group_by_id_with_access")

	action := NewViewAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.achievementGroupByID(ctx, id)
}

// CreateAchievementGroup creates a new achievement group with access control.
func (s *ATS) CreateAchievementGroup(
	ctx context.Context,
	actingUser company.User,
	options achievement.GroupCreateOptions,
) (*ent.AchievementGroup, error) {
	rec := event.Get(ctx).Sub("atsvc/create_achievement_group_with_access")

	action := NewHandleAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.createAchievementGroup(ctx, options)
}

// UpdateAchievementGroup updates an achievement group with access control.
func (s *ATS) UpdateAchievementGroup(
	ctx context.Context,
	actingUser company.User,
	id uuid.UUID,
	options achievement.GroupUpdateOptions,
) (*ent.AchievementGroup, error) {
	rec := event.Get(ctx).Sub("atsvc/update_achievement_group_with_access")

	action := NewHandleAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.updateAchievementGroup(ctx, id, options)
}

// AchievementTemplates gets all achievement templates with access control.
func (s *ATS) AchievementTemplates(
	ctx context.Context,
	actingUser company.User,
	options achievement.TemplateSearchOptions,
) (ent.AchievementTemplates, error) {
	rec := event.Get(ctx).Sub("atsvc/achievement_templates_with_access")

	action := NewViewAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.achievementTemplates(ctx, options)
}

// AchievementTemplateByID gets an achievement template by ID with access control.
func (s *ATS) AchievementTemplateByID(
	ctx context.Context,
	actingUser company.User,
	id uuid.UUID,
) (*ent.AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("atsvc/achievement_template_by_id_with_access")

	action := NewViewAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.achievementTemplateByID(ctx, id)
}

// CreateAchievementTemplate creates a new achievement template with access control.
func (s *ATS) CreateAchievementTemplate(
	ctx context.Context,
	actingUser company.User,
	options achievement.TemplateCreateOptions,
) (*ent.AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("atsvc/create_achievement_template_with_access")

	action := NewHandleAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.createAchievementTemplate(ctx, options)
}

// UpdateAchievementTemplate updates an achievement template with access control.
func (s *ATS) UpdateAchievementTemplate(
	ctx context.Context,
	actingUser company.User,
	id uuid.UUID,
	options achievement.TemplateUpdateOptions,
) (*ent.AchievementTemplate, error) {
	rec := event.Get(ctx).Sub("atsvc/update_achievement_template_with_access")

	action := NewHandleAchievementTemplatesAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.updateAchievementTemplate(ctx, id, options)
}
