package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// ReviewAchievement reviews an achievement, setting points and optionally a comment.
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
		"points_assigned", opt.PointsAssigned,
		"comment_length", len(opt.Comment),
	)

	var updatedAch *ent.Achievement
	err := withTx(ctx, s.client, func(tx *ent.Tx) error {
		var ach *ent.Achievement
		err := rec.Operation("query_achievement", func(opRec *event.Record) error {
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
		err = rec.Operation("get_reviewer", func(opRec *event.Record) error {
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

		err = rec.Operation("validate_review", func(opRec *event.Record) error {
			currentStatus := achievement.Status(ach.Status)
			reviewerRole := reviewer.Role
			templateKind := ach.Edges.Template.Kind

			// Check if the assigned points exceed the template's limit
			pointsLimit := ach.Edges.Template.PointsLimit
			if opt.PointsAssigned > pointsLimit {
				return achievement.ErrPointsLimitExceeded
			}

			// If not in a reviewable status
			if currentStatus != achievement.StatusDepheadReview && currentStatus != achievement.StatusInspectorReview {
				return achievement.ErrWrongAchievementStatus
			}

			// Validate reviewer role based on current status
			var validReviewer bool
			switch currentStatus {
			case achievement.StatusDepheadReview:
				// Department head review
				validReviewer = reviewerRole == sesc.Dephead
			case achievement.StatusInspectorReview:
				// Inspector review - check if reviewer has the correct role based on template kind
				expectedRole := templateKind.InspectorRole()
				validReviewer = reviewerRole == expectedRole
			default:
				validReviewer = false
			}

			// Check if the reviewer has the required role
			if !validReviewer {
				return sesc.ErrInvalidRole
			}

			return nil
		})
		if err != nil {
			return err
		}

		// Create the review
		err = rec.Operation("create_review", func(opRec *event.Record) error {
			reviewID, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate review ID: %w", err)
			}

			start := time.Now()
			_, err = tx.AchievementReview.Create().
				SetID(reviewID).
				SetAchievementID(opt.AchievementID).
				SetReviewerID(opt.ReviewerID).
				SetPointsAssigned(opt.PointsAssigned).
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

		// Update achievement status and points
		err = rec.Operation("update_achievement", func(opRec *event.Record) error {
			currentStatus := achievement.Status(ach.Status)
			reviewerRole := reviewer.Role
			templateKind := ach.Edges.Template.Kind

			// Determine new status based on current status, reviewer role, and points
			var newStatus achievement.Status
			switch currentStatus {
			case achievement.StatusDepheadReview:
				// Department head review
				if reviewerRole == sesc.Dephead {
					if opt.PointsAssigned > 0 {
						// If points > 0, move to inspector review
						newStatus = achievement.StatusInspectorReview
					} else {
						// If points = 0, mark as done
						newStatus = achievement.StatusDone
					}
				}
			case achievement.StatusInspectorReview:
				// Inspector review - check if reviewer has the correct role based on template kind
				expectedRole := templateKind.InspectorRole()
				if reviewerRole == expectedRole {
					// After inspector review, mark as done
					newStatus = achievement.StatusDone
				}
			}

			start := time.Now()
			updated, err := tx.Achievement.UpdateOne(ach).
				SetStatus(string(newStatus)).
				SetPoints(opt.PointsAssigned).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to update achievement: %w", err)
			}

			updatedAch = updated
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return updatedAch, err
}
