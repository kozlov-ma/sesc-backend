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
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/review_achievement")

	// Group parameters together
	rec.Sub("params").Set(
		"achievement_id", opt.AchievementID,
		"achievement_owner_id", opt.AchievementOwnerID,
		"reviewer_id", opt.ReviewerID,
		"points_assigned", opt.PointsAssigned,
		"comment_length", len(opt.Comment),
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Start a transaction and get achievement data
	var (
		tx                *ent.Tx
		achievementEntity *ent.Achievement
		reviewer          *ent.User
		reviewID          UUID
		updatedEntity     *ent.Achievement
		newStatus         achievement.Status
		isValidReviewer   bool
	)

	// Initialize transaction
	err := rec.Operation("start_transaction", func(opRec *event.Record) error {
		queryStart := time.Now()
		txn, err := s.client.Tx(ctx)
		queryCount++
		opRec.Add("tx_start_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to start transaction: %w", err))
			return err
		}
		tx = txn
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Get achievement with all related data
	err = rec.Operation("query_achievement", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"achievement_id", opt.AchievementID,
			"owner_id", opt.AchievementOwnerID,
		)

		queryStart := time.Now()
		entity, err := tx.Achievement.Query().
			Where(
				entAchievement.ID(opt.AchievementID),
				entAchievement.OwnerID(opt.AchievementOwnerID),
			).
			WithTemplate().
			WithOwner().
			WithDocuments(func(q *ent.AchievementDocumentQuery) {
				q.WithFile()
			}).
			WithReviews(func(q *ent.AchievementReviewQuery) {
				q.WithReviewer()
			}).
			Only(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if ent.IsNotFound(err) {
			opRec.Add(events.Error, "achievement not found")
			return rollback(tx, achievement.ErrAchievementNotFound)
		}
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
			return rollback(tx, err)
		}

		achievementEntity = entity
		opRec.Set("current_status", achievementEntity.Status)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Get reviewer to ensure they exist
	err = rec.Operation("get_reviewer", func(opRec *event.Record) error {
		opRec.Sub("params").Set("reviewer_id", opt.ReviewerID)

		queryStart := time.Now()
		user, err := tx.User.Get(ctx, opt.ReviewerID)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if ent.IsNotFound(err) {
			opRec.Add(events.Error, "reviewer not found")
			return rollback(tx, sesc.ErrUserNotFound)
		}
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to get reviewer: %w", err))
			return rollback(tx, err)
		}

		reviewer = user
		opRec.Set("reviewer_role", reviewer.Role)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Validate review parameters
	err = rec.Operation("validate_review", func(opRec *event.Record) error {
		// Check if achievement is in the correct status for review
		currentStatus := achievement.Status(achievementEntity.Status)
		reviewerRole := reviewer.Role
		templateKind := achievementEntity.Edges.Template.Kind

		opRec.Sub("params").Set(
			"current_status", currentStatus,
			"reviewer_role", reviewerRole,
			"template_kind", templateKind,
			"points_assigned", opt.PointsAssigned,
		)

		// Check if the assigned points exceed the template's limit
		pointsLimit := achievementEntity.Edges.Template.PointsLimit
		opRec.Set("points_limit", pointsLimit)

		if opt.PointsAssigned > pointsLimit {
			opRec.Add(
				events.Error,
				fmt.Sprintf("points assigned (%d) exceed template limit (%d)", opt.PointsAssigned, pointsLimit),
			)
			return rollback(tx, achievement.ErrPointsLimitExceeded)
		}

		// Determine the new status based on current status, reviewer role, and points
		status, validReviewer := determineNewStatus(
			currentStatus,
			reviewerRole,
			templateKind,
			opt.PointsAssigned,
			opRec,
		)

		newStatus = status
		isValidReviewer = validReviewer

		// If not in a reviewable status
		if currentStatus != achievement.StatusDepheadReview && currentStatus != achievement.StatusInspectorReview {
			opRec.Add(events.Error, "achievement is not in a reviewable status")
			return rollback(tx, achievement.ErrWrongAchievementStatus)
		}

		// Check if the reviewer has the required role
		if !isValidReviewer {
			opRec.Add(events.Error, "reviewer does not have the required role")
			return rollback(tx, sesc.ErrInvalidRole)
		}

		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Create the review and update the achievement
	err = rec.Operation("create_review", func(opRec *event.Record) error {
		// Generate new UUID for review
		id, err := uuid.NewV7()
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to generate review ID: %w", err))
			return rollback(tx, err)
		}
		reviewID = id
		opRec.Set("review_id", reviewID)

		// Create review
		queryStart := time.Now()
		_, err = tx.AchievementReview.Create().
			SetID(reviewID).
			SetAchievementID(opt.AchievementID).
			SetReviewerID(opt.ReviewerID).
			SetPointsAssigned(opt.PointsAssigned).
			SetComment(opt.Comment).
			Save(ctx)
		queryCount++
		opRec.Add("create_review_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to create review: %w", err))
			return rollback(tx, err)
		}

		// Update achievement status and points
		queryStart = time.Now()
		updated, err := tx.Achievement.UpdateOne(achievementEntity).
			SetStatus(string(newStatus)).
			SetPoints(opt.PointsAssigned).
			Save(ctx)
		queryCount++
		opRec.Add("update_achievement_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to update achievement: %w", err))
			return rollback(tx, err)
		}

		updatedEntity = updated
		opRec.Set("new_status", updatedEntity.Status)
		opRec.Set("new_points", updatedEntity.Points)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Commit transaction
	err = rec.Operation("commit_transaction", func(opRec *event.Record) error {
		queryStart := time.Now()
		err := tx.Commit()
		queryCount++
		opRec.Add("commit_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to commit transaction: %w", err))
			return err
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert to domain model
	var result achievement.Achievement
	err = rec.Operation("convert_to_domain", func(opRec *event.Record) error {
		domainModel := convertAchievementToModel(achievementEntity, reviewer, reviewID, opt, opRec)

		// Update the status and points to match the updated entity
		domainModel.Status = achievement.Status(updatedEntity.Status)
		domainModel.Points = updatedEntity.Points

		result = domainModel
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Record final result
	rec.Sub("result").Set(
		"achievement_id", result.ID,
		"new_status", result.Status,
		"points", result.Points,
		"review_id", reviewID,
	)

	return result, nil
}
