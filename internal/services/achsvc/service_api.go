package achsvc

import (
	"bytes"
	"context"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// GetAchievement retrieves an achievement by ID with access control.
func (s *ACS) GetAchievement(
	ctx context.Context,
	actingUser company.User,
	id UUID,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/get_achievement_with_access")

	ach, err := s.getAchievement(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, err
	}

	// Check access using ViewAchievementAction
	action := NewViewAchievementAction(
		ach.OwnerID,
		ach.DepartmentID,
		achievement.Status(ach.Status),
		ach.Edges.Template.ReviewerRole,
	)
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
	return ach, nil
}

// CreateAchievement creates a new achievement with access control.
func (s *ACS) CreateAchievement(
	ctx context.Context,
	actingUser company.User,
	opt achievement.CreateOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/create_achievement_with_access")

	action := NewCreateAchievementAction(opt.ForUserID)
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
	return s.createAchievement(ctx, opt)
}

// DeleteAchievement deletes an achievement with access control.
func (s *ACS) DeleteAchievement(
	ctx context.Context,
	actingUser company.User,
	opt achievement.DeleteOptions,
) error {
	rec := event.Get(ctx).Sub("achsvc/delete_achievement_with_access")

	// Get achievement to verify owner
	ach, err := s.getAchievement(ctx, opt.AchievementID)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return err
	}

	action := NewModifyAchievementAction(ach.OwnerID, opt.AchievementID)
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.deleteAchievement(ctx, opt)
}

// AddDocument adds a document to an achievement with access control.
func (s *ACS) AddDocument(
	ctx context.Context,
	actingUser company.User,
	opt achievement.AddDocumentOptions,
) (*ent.AchievementDocument, error) {
	rec := event.Get(ctx).Sub("achsvc/add_document_with_access")

	action := NewModifyAchievementAction(opt.OwnerID, opt.AchievementID)
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
	return s.addDocument(ctx, opt)
}

// RemoveDocument removes a document from an achievement with access control.
func (s *ACS) RemoveDocument(
	ctx context.Context,
	actingUser company.User,
	opt achievement.RemoveDocumentOptions,
) error {
	rec := event.Get(ctx).Sub("achsvc/remove_document_with_access")

	// Get achievement to get owner ID
	ach, err := s.getAchievement(ctx, opt.AchievementID)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return err
	}

	action := NewModifyAchievementAction(ach.OwnerID, opt.AchievementID)
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.removeDocument(ctx, opt)
}

// SubmitAchievement submits an achievement for review with access control.
func (s *ACS) SubmitAchievement(
	ctx context.Context,
	actingUser company.User,
	opt achievement.SubmitOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/submit_achievement_with_access")

	// Get achievement to check status
	ach, err := s.getAchievement(ctx, opt.AchievementID)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, err
	}

	action := NewSubmitAchievementAction(opt.OwnerID, achievement.Status(ach.Status))
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
	return s.submitAchievement(ctx, opt)
}

// ReviewAchievement reviews an achievement with access control.
func (s *ACS) ReviewAchievement(
	ctx context.Context,
	actingUser company.User,
	opt achievement.ReviewOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/review_achievement_with_access")

	// Get achievement to check its properties for access control
	ach, err := s.getAchievement(ctx, opt.AchievementID)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, err
	}

	// Check access using ReviewAchievementAction
	action := NewReviewAchievementAction(
		achievement.Status(ach.Status),
		ach.Edges.Template.ReviewerRole,
		ach.DepartmentID,
	)
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
	return s.reviewAchievement(ctx, actingUser.ID, opt)
}

// UpdateAchievementPoints updates achievement points with access control.
func (s *ACS) UpdateAchievementPoints(
	ctx context.Context,
	actingUser company.User,
	opt achievement.UpdatePointsOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/update_achievement_points_with_access")

	action := NewModifyAchievementAction(opt.OwnerID, opt.AchievementID)
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
	return s.updateAchievementPoints(ctx, opt)
}

// GetUserAchievements retrieves achievements for a user with access control.
func (s *ACS) GetUserAchievements(
	ctx context.Context,
	actingUser company.User,
	offset, limit int,
	requireChanges bool,
) (ent.Achievements, int, error) {
	rec := event.Get(ctx).Sub("achsvc/get_user_achievements_with_access")

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)

	return s.getUserAchievements(ctx, actingUser.ID, actingUser.ID, offset, limit, requireChanges)
}

// GetUsersWithAchievements retrieves users with achievements with access control.
func (s *ACS) GetUsersWithAchievements(
	ctx context.Context,
	actingUser company.User,
	offset, limit int,
	search string,
) ([]company.User, int, error) {
	rec := event.Get(ctx).Sub("achsvc/get_users_with_achievements_with_access")

	// Get user's department for access check
	action := NewListUsersWithAchievementsAction(actingUser.ID, actingUser.DepartmentID)
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return nil, 0, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.getUsersWithAchievements(ctx, actingUser.ID, offset, limit, search)
}

// GenerateUserPointsReport generates a user points report with access control.
func (s *ACS) GenerateUserPointsReport(
	ctx context.Context,
	actingUser company.User,
) (*bytes.Buffer, error) {
	rec := event.Get(ctx).Sub("achsvc/generate_user_points_report_with_access")

	action := NewGenerateUserPointsReportAction()
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
	return s.generateUserPointsReport(ctx)
}

// MarkAllDoneAchievementsAsAccounted marks achievements as accounted with access control.
func (s *ACS) MarkAllDoneAchievementsAsAccounted(
	ctx context.Context,
	actingUser company.User,
) (int, error) {
	rec := event.Get(ctx).Sub("achsvc/mark_all_done_achievements_as_accounted_with_access")

	action := NewAccountingAction()
	if !actingUser.Can(action) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", actingUser)
		return 0, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", actingUser)
	return s.markAllDoneAchievementsAsAccounted(ctx)
}
