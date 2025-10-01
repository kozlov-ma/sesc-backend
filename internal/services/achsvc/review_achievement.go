package achsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// validateReviewRequest validates that the review request is valid
func (s *ACS) validateReviewRequest(
	ach *ent.Achievement,
	reviewer *ent.User,
	action achievement.ReviewAction,
) error {
	currentStatus := achievement.Status(ach.Status)
	if currentStatus != achievement.StatusDepheadReview &&
		currentStatus != achievement.StatusInspectorReview {
		return achievement.ErrWrongAchievementStatus
	}

	reviewerRole := reviewer.Role
	needReviewerRole := ach.Edges.Template.ReviewerRole

	var validReviewer bool
	switch currentStatus {
	case achievement.StatusDepheadReview:
		validReviewer = reviewerRole == sesc.Dephead
	case achievement.StatusInspectorReview:
		validReviewer = reviewerRole == needReviewerRole
	}

	if !validReviewer {
		return sesc.ErrInvalidRole
	}

	if action != achievement.ReviewActionApprove &&
		action != achievement.ReviewActionDisapprove &&
		action != achievement.ReviewActionRequestChanges {
		return fmt.Errorf("invalid review action: %s", action)
	}

	return nil
}

// calculateReviewPoints calculates the points to assign in the review
func calculateReviewPoints(action achievement.ReviewAction, currentPoints int) int {
	switch action {
	case achievement.ReviewActionApprove:
		return currentPoints
	case achievement.ReviewActionDisapprove, achievement.ReviewActionRequestChanges:
		return 0
	default:
		return 0
	}
}

// calculateNewStatusAndPoints calculates the new status and points after review
func calculateNewStatusAndPoints(
	action achievement.ReviewAction,
	currentStatus achievement.Status,
	currentPoints int,
) (achievement.Status, int) {
	switch action {
	case achievement.ReviewActionApprove:
		switch currentStatus {
		case achievement.StatusDepheadReview:
			return achievement.StatusInspectorReview, currentPoints
		case achievement.StatusInspectorReview:
			return achievement.StatusDone, currentPoints
		}
	case achievement.ReviewActionDisapprove:
		return achievement.StatusDone, 0
	case achievement.ReviewActionRequestChanges:
		switch currentStatus {
		case achievement.StatusDepheadReview:
			return achievement.StatusDepheadRequestedChanges, currentPoints
		case achievement.StatusInspectorReview:
			return achievement.StatusInspectorRequestedChanges, currentPoints
		}
	}
	return currentStatus, currentPoints
}

// ReviewAchievement reviews an achievement with approve, disapprove, or request changes action.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in the correct status for review.
func (s *ACS) ReviewAchievement(
	ctx context.Context,
	opt achievement.ReviewOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/review_achievement")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"achievement_id", opt.AchievementID,
		"achievement_owner_id", opt.AchievementOwnerID,
		"reviewer_id", opt.ReviewerID,
		"action", string(opt.Action),
		"comment_length", len(opt.Comment),
	)

	var updatedAch *ent.Achievement
	err := txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		var ach *ent.Achievement
		err := rec.Operation("query_achievement", func(_ *event.Record) error {
			start := time.Now()
			entity, err := tx.Achievement.Query().
				Where(
					entAchievement.ID(opt.AchievementID),
					entAchievement.OwnerID(opt.AchievementOwnerID),
				).
				WithTemplate().
				Only(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return achievement.ErrAchievementNotFound
			}
			if err != nil {
				return fmt.Errorf("failed to query achievement: %w", err)
			}

			ach = entity
			return nil
		})
		if err != nil {
			return err
		}

		var reviewer *ent.User
		err = rec.Operation("get_reviewer", func(_ *event.Record) error {
			start := time.Now()
			user, err := tx.User.Get(ctx, opt.ReviewerID)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return sesc.ErrUserNotFound
			}
			if err != nil {
				return fmt.Errorf("failed to get reviewer: %w", err)
			}

			reviewer = user
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("validate_review", func(rec *event.Record) error {
			rec.Sub("params").Set(
				"reviewer_role", reviewer.Role.String(),
				"ach_status", ach.Status,
			)

			return s.validateReviewRequest(ach, reviewer, opt.Action)
		})
		if err != nil {
			return err
		}

		// Create the review
		err = rec.Operation("create_review", func(_ *event.Record) error {
			reviewID, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate review ID: %w", err)
			}

			pointsForReview := calculateReviewPoints(opt.Action, ach.Points)

			start := time.Now()
			_, err = tx.AchievementReview.Create().
				SetID(reviewID).
				SetAchievementID(opt.AchievementID).
				SetReviewerID(opt.ReviewerID).
				SetPointsAssigned(pointsForReview).
				SetComment(opt.Comment).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to create review: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("update_achievement", func(_ *event.Record) error {
			currentStatus := achievement.Status(ach.Status)
			newStatus, newPoints := calculateNewStatusAndPoints(
				opt.Action,
				currentStatus,
				ach.Points,
			)

			start := time.Now()
			_, err := tx.Achievement.UpdateOne(ach).
				SetStatus(string(newStatus)).
				SetPoints(newPoints).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to update achievement: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("query_updated", func(_ *event.Record) error {
			var err error
			updatedAch, err = tx.Achievement.Query().
				Where(entAchievement.ID(ach.ID)).
				WithDocuments().
				WithOwner().
				WithReviews().
				WithTemplate().
				Only(ctx)
			if err != nil {
				return fmt.Errorf("failed to query updated achievement: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return updatedAch, err
}
