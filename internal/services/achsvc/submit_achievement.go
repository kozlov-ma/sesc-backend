package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// SubmitAchievement submits an achievement for review.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *ACS) SubmitAchievement(
	ctx context.Context,
	opt achievement.SubmitOptions,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/submit_achievement")

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.OwnerID,
		"achievement_id", opt.AchievementID,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Get achievement with all related data
	var achievementEntity *ent.Achievement
	err := rec.Operation("query_achievement", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"achievement_id", opt.AchievementID,
			"owner_id", opt.OwnerID,
		)

		queryStart := time.Now()
		entity, err := s.client.Achievement.Query().
			Where(
				entAchievement.ID(opt.AchievementID),
				entAchievement.OwnerID(opt.OwnerID),
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
			return achievement.ErrAchievementNotFound
		}
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
			return err
		}

		achievementEntity = entity
		opRec.Set("current_status", achievementEntity.Status)
		opRec.Set("documents_count", len(achievementEntity.Edges.Documents))
		opRec.Set("reviews_count", len(achievementEntity.Edges.Reviews))
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Validate achievement status
	err = rec.Operation("validate_status", func(opRec *event.Record) error {
		opRec.Set("current_status", achievementEntity.Status)
		opRec.Set("required_status", string(achievement.StatusDraft))

		// Check if achievement is in draft status
		if achievementEntity.Status != string(achievement.StatusDraft) {
			opRec.Add(events.Error, "achievement is not in draft status")
			return achievement.ErrWrongAchievementStatus
		}

		opRec.Set("valid", true)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Update achievement status
	var updatedEntity *ent.Achievement
	err = rec.Operation("update_status", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"achievement_id", achievementEntity.ID,
			"new_status", string(achievement.StatusDepheadReview),
		)

		queryStart := time.Now()
		entity, err := s.client.Achievement.UpdateOne(achievementEntity).
			SetStatus(string(achievement.StatusDepheadReview)).
			Save(ctx)
		queryCount++
		opRec.Add("update_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to update achievement status: %w", err))
			return err
		}

		updatedEntity = entity
		opRec.Set("updated_status", updatedEntity.Status)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert to domain model
	var (
		template  achievement.Template
		owner     sesc.User
		documents []achievement.Document
		reviews   []achievement.Review
	)

	// Convert template
	err = rec.Operation("convert_template", func(opRec *event.Record) error {
		tmpl := achievementEntity.Edges.Template
		opRec.Sub("template").Set(
			"id", tmpl.ID,
			"name", tmpl.Name,
			"kind", tmpl.Kind,
		)

		template = achievement.Template{
			ID:          tmpl.ID,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			PointsLimit: tmpl.PointsLimit,
			GroupID:     tmpl.GroupID,
			Active:      tmpl.Active,
			Kind:        tmpl.Kind,
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert owner
	err = rec.Operation("convert_owner", func(opRec *event.Record) error {
		usr := convertUser(achievementEntity.Edges.Owner)
		owner = usr
		opRec.Set("owner_id", owner.ID)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert documents and reviews
	err = rec.Operation("convert_documents_and_reviews", func(opRec *event.Record) error {
		// Convert documents
		documents = make([]achievement.Document, 0, len(achievementEntity.Edges.Documents))
		for _, doc := range achievementEntity.Edges.Documents {
			file := sesc.File{
				ID: doc.Edges.File.ID,
			}

			documents = append(documents, achievement.Document{
				ID:     doc.ID,
				Name:   doc.Name,
				FileID: file.ID,
			})
		}
		opRec.Set("documents_count", len(documents))

		// Convert reviews
		reviews = make([]achievement.Review, 0, len(achievementEntity.Edges.Reviews))
		for _, rev := range achievementEntity.Edges.Reviews {
			reviewer := convertUser(rev.Edges.Reviewer)

			reviews = append(reviews, achievement.Review{
				ID:             rev.ID,
				From:           reviewer,
				PointsAssigned: rev.PointsAssigned,
				Comment:        rev.Comment,
			})
		}
		opRec.Set("reviews_count", len(reviews))
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Create the final result
	result := achievement.Achievement{
		ID:        updatedEntity.ID,
		Owner:     owner,
		Template:  template,
		Status:    achievement.Status(updatedEntity.Status),
		Points:    updatedEntity.Points,
		Documents: documents,
		Reviews:   reviews,
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"achievement_id", result.ID,
		"new_status", result.Status,
		"documents_count", len(documents),
		"reviews_count", len(reviews),
	)

	return result, nil
}

// determineNewStatus determines the new status for an achievement based on the current status,
// reviewer role, and points assigned.
func determineNewStatus(
	currentStatus achievement.Status,
	reviewerRole sesc.Role,
	templateKind achievement.Kind,
	pointsAssigned int,
	rec *event.Record,
) (achievement.Status, bool) {
	// Group parameters together
	rec.Sub("params").Set(
		"current_status", currentStatus,
		"reviewer_role_id", reviewerRole,
		"reviewer_role_name", reviewerRole.Name,
		"template_kind", templateKind,
		"points_assigned", pointsAssigned,
	)

	// We'll return the results directly instead of using a struct

	switch currentStatus {
	case achievement.StatusDepheadReview:
		// Department head review
		if reviewerRole == sesc.Dephead {
			if pointsAssigned > 0 {
				// If points > 0, move to inspector review
				rec.Sub("decision").Set(
					"reason", "department_head_assigned_points",
					"new_status", achievement.StatusInspectorReview,
					"is_valid_reviewer", true,
				)
				return achievement.StatusInspectorReview, true
			}
			// If points = 0, mark as done
			rec.Sub("decision").Set(
				"reason", "department_head_assigned_zero_points",
				"new_status", achievement.StatusDone,
				"is_valid_reviewer", true,
			)
			return achievement.StatusDone, true
		}
		// Not a department head
		rec.Sub("decision").Set(
			"reason", "not_department_head",
			"expected_role", sesc.Dephead,
			"is_valid_reviewer", false,
		)

	case achievement.StatusInspectorReview:
		// Inspector review - check if reviewer has the correct role based on template kind
		expectedRole := templateKind.InspectorRole()
		rec.Sub("inspector_check").Set(
			"expected_role_id", expectedRole,
			"expected_role_name", expectedRole.Name,
		)

		if reviewerRole == expectedRole {
			// After inspector review, mark as done
			rec.Sub("decision").Set(
				"reason", "inspector_reviewed",
				"new_status", achievement.StatusDone,
				"is_valid_reviewer", true,
			)
			return achievement.StatusDone, true
		}
		// Not the correct inspector
		rec.Sub("decision").Set(
			"reason", "wrong_inspector_role",
			"is_valid_reviewer", false,
		)

	default:
		// Status not eligible for review
		rec.Sub("decision").Set(
			"reason", "status_not_reviewable",
			"is_valid_reviewer", false,
		)
	}

	// Not a valid reviewer or status
	rec.Add(events.Error, "invalid reviewer or status")
	return "", false
}

// convertAchievementToModel converts an achievement entity to a domain model
func convertAchievementToModel(
	achievementEntity *ent.Achievement,
	reviewerEntity *ent.User,
	reviewID UUID,
	opt achievement.ReviewOptions,
	rec *event.Record,
) achievement.Achievement {
	subRec := rec.Sub("convert_achievement_to_model")
	subRec.Set("achievement_id", achievementEntity.ID)
	subRec.Set("reviewer_id", reviewerEntity.ID)
	subRec.Set("review_id", reviewID)

	// Convert template
	template := achievement.Template{
		ID:          achievementEntity.Edges.Template.ID,
		Name:        achievementEntity.Edges.Template.Name,
		Description: achievementEntity.Edges.Template.Description,
		PointsLimit: achievementEntity.Edges.Template.PointsLimit,
		GroupID:     achievementEntity.Edges.Template.GroupID,
		Active:      achievementEntity.Edges.Template.Active,
		Kind:        achievementEntity.Edges.Template.Kind,
	}

	// Convert owner
	owner := convertUser(achievementEntity.Edges.Owner)

	// Convert documents
	documents := make([]achievement.Document, 0, len(achievementEntity.Edges.Documents))
	for _, doc := range achievementEntity.Edges.Documents {
		file := sesc.File{
			ID: doc.Edges.File.ID,
		}

		documents = append(documents, achievement.Document{
			ID:     doc.ID,
			Name:   doc.Name,
			FileID: file.ID,
		})
	}
	subRec.Set("documents_count", len(documents))

	// Convert existing reviews
	reviews := make([]achievement.Review, 0, len(achievementEntity.Edges.Reviews)+1)
	for _, rev := range achievementEntity.Edges.Reviews {
		revUser := convertUser(rev.Edges.Reviewer)

		reviews = append(reviews, achievement.Review{
			ID:             rev.ID,
			From:           revUser,
			PointsAssigned: rev.PointsAssigned,
			Comment:        rev.Comment,
		})
	}

	// Add the new review
	reviewerUser := convertUser(reviewerEntity)

	reviews = append(reviews, achievement.Review{
		ID:             reviewID,
		From:           reviewerUser,
		PointsAssigned: opt.PointsAssigned,
		Comment:        opt.Comment,
	})
	subRec.Set("reviews_count", len(reviews))

	result := achievement.Achievement{
		ID:        achievementEntity.ID,
		Owner:     owner,
		Template:  template,
		Status:    achievement.Status(achievementEntity.Status),
		Points:    achievementEntity.Points,
		Documents: documents,
		Reviews:   reviews,
	}

	subRec.Set("result_status", result.Status)
	subRec.Set("result_points", result.Points)
	return result
}

func convertUser(u *ent.User) User {
	var dept Department
	dep := u.Edges.Department
	if dep != nil {
		dept = Department{
			ID:          dep.ID,
			Name:        dep.Name,
			Description: dep.Description,
		}
	}

	return User{
		ID:                u.ID,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		MiddleName:        u.MiddleName,
		PictureURL:        u.PictureURL,
		Suspended:         u.Suspended,
		Department:        dept,
		Role:              u.Role,
		Subdivision:       u.Subdivision,
		JobTitle:          u.JobTitle,
		EmploymentRate:    u.EmploymentRate,
		AcademicDegree:    u.AcademicDegree,
		PersonnelCategory: u.PersonnelCategory,
		EmploymentType:    u.EmploymentType,
		AcademicTitle:     u.AcademicTitle,
		Honors:            u.Honors,
		Category:          u.Category,
		DateOfEmployment:  u.DateOfEmployment,
		UnemploymentDate:  u.UnemploymentDate,
	}
}
