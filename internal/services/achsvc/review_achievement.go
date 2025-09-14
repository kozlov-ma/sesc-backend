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
	"github.com/kozlov-ma/sesc-backend/internal/services/txhelper"
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
	err := txhelper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
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

			currentStatus := achievement.Status(ach.Status)
			if currentStatus != achievement.StatusDepheadReview && currentStatus != achievement.StatusInspectorReview {
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

			pointsLimit := ach.Edges.Template.PointsLimit
			if opt.PointsAssigned > pointsLimit {
				return achievement.ErrPointsLimitExceeded
			}

			return nil
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

		err = rec.Operation("update_achievement", func(_ *event.Record) error {
			currentStatus := achievement.Status(ach.Status)

			var newStatus achievement.Status
			switch {
			case opt.PointsAssigned == 0:
				newStatus = achievement.StatusDone
			case currentStatus == achievement.StatusDepheadReview:
				newStatus = achievement.StatusInspectorReview
			case currentStatus == achievement.StatusInspectorReview:
				newStatus = achievement.StatusDone
			}

			start := time.Now()
			_, err := tx.Achievement.UpdateOne(ach).
				SetStatus(string(newStatus)).
				SetPoints(opt.PointsAssigned).
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
