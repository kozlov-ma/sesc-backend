// Package sescsvc provides services for managing SESC employees, departments, and achievements.
package sescsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementreview"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// CreateAchievement creates a new achievement for a user based on a template.
// Returns achievement.ErrAchievementTemplateNotFound if the template does not exist.
func (s *SESC) CreateAchievement(
	ctx context.Context,
	opt achievement.CreateOptions,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/create_achievement")
	rec.Set("user_id", opt.ForUser.ID)
	rec.Set("template_id", opt.TemplateID)

	// Check if template exists
	template, err := s.AchievementTemplateByID(ctx, opt.TemplateID)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to get template: %w", err))
		return achievement.Achievement{}, err
	}

	// Generate new UUID for achievement
	id, err := s.newUUID()
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to generate ID: %w", err))
		return achievement.Achievement{}, err
	}
	rec.Add("generated_id", id)

	// Start a transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to start transaction: %w", err))
		return achievement.Achievement{}, err
	}

	// Create achievement
	achievementEntity, err := tx.Achievement.Create().
		SetID(id).
		SetOwnerID(opt.ForUser.ID).
		SetTemplateID(opt.TemplateID).
		SetStatus(string(achievement.StatusDraft)).
		SetPoints(template.PointsLimit).
		Save(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to create achievement: %w", err))
		return achievement.Achievement{}, rollback(tx, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to commit transaction: %w", err))
		return achievement.Achievement{}, err
	}

	result := achievement.Achievement{
		ID:       achievementEntity.ID,
		Owner:    opt.ForUser,
		Template: template,
		Status:   achievement.Status(achievementEntity.Status),
		Points:   achievementEntity.Points,
	}

	rec.Add("created_achievement", result)
	return result, nil
}

// DeleteAchievement deletes an achievement.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *SESC) DeleteAchievement(
	ctx context.Context,
	opt achievement.DeleteOptions,
) error {
	rec := event.Get(ctx).Sub("sesc/delete_achievement")
	rec.Add("user_id", opt.OwnerID)
	rec.Add("achievement_id", opt.AchievementID)

	// Get achievement to check status
	achievementEntity, err := s.client.Achievement.Query().
		Where(
			entAchievement.ID(opt.AchievementID),
			entAchievement.OwnerID(opt.OwnerID),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "achievement not found")
		return achievement.ErrAchievementNotFound
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
		return err
	}

	// Check if achievement is in draft status
	if achievementEntity.Status != string(achievement.StatusDraft) {
		rec.Add(events.Error, "achievement is not in draft status")
		return achievement.ErrWrongAchievementStatus
	}

	// Start a transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to start transaction: %w", err))
		return err
	}

	// Delete achievement documents
	_, err = tx.AchievementDocument.Delete().
		Where(achievementdocument.AchievementID(opt.AchievementID)).
		Exec(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to delete achievement documents: %w", err))
		return rollback(tx, err)
	}

	// Delete achievement reviews
	_, err = tx.AchievementReview.Delete().
		Where(achievementreview.AchievementID(opt.AchievementID)).
		Exec(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to delete achievement reviews: %w", err))
		return rollback(tx, err)
	}

	// Delete achievement
	err = tx.Achievement.DeleteOne(achievementEntity).Exec(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to delete achievement: %w", err))
		return rollback(tx, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to commit transaction: %w", err))
		return err
	}

	rec.Add("success", true)
	return nil
}

// AddDocument adds a document to an achievement.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *SESC) AddDocument(
	ctx context.Context,
	opt achievement.AddDocumentOptions,
) (achievement.Document, error) {
	rec := event.Get(ctx).Sub("sesc/add_document")
	rec.Add("user_id", opt.OwnerID)
	rec.Add("achievement_id", opt.AchievementID)
	rec.Add("file_id", opt.FileID)
	rec.Add("name", opt.Name)

	// Get achievement to check status
	achievementEntity, err := s.client.Achievement.Query().
		Where(
			entAchievement.ID(opt.AchievementID),
			entAchievement.OwnerID(opt.OwnerID),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "achievement not found")
		return achievement.Document{}, achievement.ErrAchievementNotFound
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
		return achievement.Document{}, err
	}

	// Check if achievement is in draft status
	if achievementEntity.Status != string(achievement.StatusDraft) {
		rec.Add(events.Error, "achievement is not in draft status")
		return achievement.Document{}, achievement.ErrWrongAchievementStatus
	}

	// Get file to ensure it exists
	fileEntity, err := s.client.File.Get(ctx, opt.FileID)
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "file not found")
		return achievement.Document{}, sesc.ErrFileNotFound
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to get file: %w", err))
		return achievement.Document{}, err
	}

	// Generate new UUID for document
	id, err := s.newUUID()
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to generate ID: %w", err))
		return achievement.Document{}, err
	}
	rec.Add("generated_document_id", id)

	// Create document
	documentEntity, err := s.client.AchievementDocument.Create().
		SetID(id).
		SetAchievementID(opt.AchievementID).
		SetName(opt.Name).
		SetFileID(opt.FileID).
		Save(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to create document: %w", err))
		return achievement.Document{}, err
	}

	file := sesc.File{
		ID: fileEntity.ID,
	}

	result := achievement.Document{
		ID:     documentEntity.ID,
		Name:   documentEntity.Name,
		FileID: file.ID,
	}

	rec.Add("created_document", result)
	return result, nil
}

// RemoveDocument removes a document from an achievement.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrDocumentNotFound if the document does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *SESC) RemoveDocument(
	ctx context.Context,
	opt achievement.RemoveDocumentOptions,
) error {
	rec := event.Get(ctx).Sub("sesc/remove_document")
	rec.Add("user_id", opt.OwnerID)
	rec.Add("achievement_id", opt.AchievementID)
	rec.Add("document_id", opt.DocumentID)

	// Get achievement to check status
	achievementEntity, err := s.client.Achievement.Query().
		Where(
			entAchievement.ID(opt.AchievementID),
			entAchievement.OwnerID(opt.OwnerID),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "achievement not found")
		return achievement.ErrAchievementNotFound
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
		return err
	}

	// Check if achievement is in draft status
	if achievementEntity.Status != string(achievement.StatusDraft) {
		rec.Add(events.Error, "achievement is not in draft status")
		return achievement.ErrWrongAchievementStatus
	}

	// Check if document exists and belongs to the achievement
	documentEntity, err := s.client.AchievementDocument.Query().
		Where(
			achievementdocument.ID(opt.DocumentID),
			achievementdocument.AchievementID(opt.AchievementID),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "document not found")
		return achievement.ErrDocumentNotFound
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query document: %w", err))
		return err
	}

	// Delete document
	err = s.client.AchievementDocument.DeleteOne(documentEntity).Exec(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to delete document: %w", err))
		return err
	}

	rec.Add("success", true)
	return nil
}

// SubmitAchievement submits an achievement for review.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *SESC) SubmitAchievement(
	ctx context.Context,
	opt achievement.SubmitOptions,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/submit_achievement")
	rec.Add("user_id", opt.OwnerID)
	rec.Add("achievement_id", opt.AchievementID)

	// Get achievement to check status
	achievementEntity, err := s.client.Achievement.Query().
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
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "achievement not found")
		return achievement.Achievement{}, achievement.ErrAchievementNotFound
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
		return achievement.Achievement{}, err
	}

	// Check if achievement is in draft status
	if achievementEntity.Status != string(achievement.StatusDraft) {
		rec.Add(events.Error, "achievement is not in draft status")
		return achievement.Achievement{}, achievement.ErrWrongAchievementStatus
	}

	// Update achievement status to department head review
	updatedEntity, err := s.client.Achievement.UpdateOne(achievementEntity).
		SetStatus(string(achievement.StatusDepheadReview)).
		Save(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to update achievement status: %w", err))
		return achievement.Achievement{}, err
	}

	// Convert template
	template := achievement.Template{
		ID:          achievementEntity.Edges.Template.ID,
		Name:        achievementEntity.Edges.Template.Name,
		Description: achievementEntity.Edges.Template.Description,
		PointsLimit: achievementEntity.Edges.Template.PointsLimit,
		GroupID:     achievementEntity.Edges.Template.GroupID,
		Active:      achievementEntity.Edges.Template.Active,
		Kind:        achievement.Kind(achievementEntity.Edges.Template.Kind),
	}

	// Convert owner
	owner, err := convertUser(achievementEntity.Edges.Owner)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to convert owner: %w", err))
		return achievement.Achievement{}, err
	}

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

	// Convert reviews
	reviews := make([]achievement.Review, 0, len(achievementEntity.Edges.Reviews))
	for _, rev := range achievementEntity.Edges.Reviews {
		reviewer, err := convertUser(rev.Edges.Reviewer)
		if err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to convert reviewer: %w", err))
			return achievement.Achievement{}, err
		}

		reviews = append(reviews, achievement.Review{
			ID:             rev.ID,
			From:           reviewer,
			PointsAssigned: rev.PointsAssigned,
			Comment:        rev.Comment,
		})
	}

	result := achievement.Achievement{
		ID:        updatedEntity.ID,
		Owner:     owner,
		Template:  template,
		Status:    achievement.Status(updatedEntity.Status),
		Points:    updatedEntity.Points,
		Documents: documents,
		Reviews:   reviews,
	}

	rec.Add("updated_achievement", result)
	return result, nil
}

// determineNewStatus determines the new status for an achievement based on the current status,
// reviewer role, and points assigned.
func determineNewStatus(
	currentStatus achievement.Status,
	reviewerRole sesc.Role,
	templateKind achievement.Kind,
	pointsAssigned int,
) (achievement.Status, bool) {
	switch currentStatus {
	case achievement.StatusDepheadReview:
		// Department head review
		if reviewerRole.ID == sesc.Dephead.ID {
			if pointsAssigned > 0 {
				// If points > 0, move to inspector review
				return achievement.StatusInspectorReview, true
			}
			// If points = 0, mark as done
			return achievement.StatusDone, true
		}
	case achievement.StatusInspectorReview:
		// Inspector review - check if reviewer has the correct role based on template kind
		expectedRole := templateKind.InspectorRole()
		if reviewerRole.ID == expectedRole.ID {
			// After inspector review, mark as done
			return achievement.StatusDone, true
		}
	}

	// Not a valid reviewer or status
	return "", false
}

// convertAchievementToModel converts an achievement entity to a domain model
func convertAchievementToModel(
	achievementEntity *ent.Achievement,
	reviewerEntity *ent.User,
	reviewID UUID,
	opt achievement.ReviewOptions,
) (achievement.Achievement, error) {
	// Convert template
	template := achievement.Template{
		ID:          achievementEntity.Edges.Template.ID,
		Name:        achievementEntity.Edges.Template.Name,
		Description: achievementEntity.Edges.Template.Description,
		PointsLimit: achievementEntity.Edges.Template.PointsLimit,
		GroupID:     achievementEntity.Edges.Template.GroupID,
		Active:      achievementEntity.Edges.Template.Active,
		Kind:        achievement.Kind(achievementEntity.Edges.Template.Kind),
	}

	// Convert owner
	owner, err := convertUser(achievementEntity.Edges.Owner)
	if err != nil {
		return achievement.Achievement{}, fmt.Errorf("failed to convert owner: %w", err)
	}

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

	// Convert existing reviews
	reviews := make([]achievement.Review, 0, len(achievementEntity.Edges.Reviews)+1)
	for _, rev := range achievementEntity.Edges.Reviews {
		revUser, err := convertUser(rev.Edges.Reviewer)
		if err != nil {
			return achievement.Achievement{}, fmt.Errorf("failed to convert reviewer: %w", err)
		}

		reviews = append(reviews, achievement.Review{
			ID:             rev.ID,
			From:           revUser,
			PointsAssigned: rev.PointsAssigned,
			Comment:        rev.Comment,
		})
	}

	// Add the new review
	reviewerUser, err := convertUser(reviewerEntity)
	if err != nil {
		return achievement.Achievement{}, fmt.Errorf("failed to convert reviewer: %w", err)
	}

	reviews = append(reviews, achievement.Review{
		ID:             reviewID,
		From:           reviewerUser,
		PointsAssigned: opt.PointsAssigned,
		Comment:        opt.Comment,
	})

	return achievement.Achievement{
		ID:        achievementEntity.ID,
		Owner:     owner,
		Template:  template,
		Status:    achievement.Status(achievementEntity.Status),
		Points:    achievementEntity.Points,
		Documents: documents,
		Reviews:   reviews,
	}, nil
}

// GetAchievement retrieves an achievement by ID and owner ID.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
func (s *SESC) GetAchievement(
	ctx context.Context,
	achievementID UUID,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/get_achievement")
	rec.Add("achievement_id", achievementID)

	// Get achievement with all related data
	achievementEntity, err := s.client.Achievement.Query().
		Where(
			entAchievement.ID(achievementID),
		).
		WithTemplate().
		WithOwner(func(q *ent.UserQuery) {
			q.WithDepartment()
		}).
		WithDocuments(func(q *ent.AchievementDocumentQuery) {
			q.WithFile()
		}).
		WithReviews(func(q *ent.AchievementReviewQuery) {
			q.WithReviewer()
		}).
		Only(ctx)

	if ent.IsNotFound(err) {
		rec.Add(events.Error, "achievement not found")
		return achievement.Achievement{}, achievement.ErrAchievementNotFound
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
		return achievement.Achievement{}, err
	}

	// Convert to domain model
	result, err := convertAchievementEntityToDomain(achievementEntity)
	if err != nil {
		rec.Add(events.Error, err)
		return achievement.Achievement{}, err
	}

	rec.Add("achievement", result)
	return result, nil
}

func (s *SESC) GetAchievementsForUser(
	ctx context.Context,
	userID UUID,
	offset, limit int,
) ([]achievement.Achievement, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_achievements_for_user")
	rec.Add("user_id", userID)
	rec.Add("offset", offset)
	rec.Add("limit", limit)

	// Count total achievements for the user
	total, err := s.client.Achievement.Query().
		Where(entAchievement.OwnerID(userID)).
		Count(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to count achievements: %w", err))
		return nil, 0, err
	}

	// Get achievements with pagination
	achievementEntities, err := s.client.Achievement.Query().
		Where(entAchievement.OwnerID(userID)).
		WithTemplate().
		WithOwner(func(q *ent.UserQuery) {
			q.WithDepartment()
		}).
		WithDocuments(func(q *ent.AchievementDocumentQuery) {
			q.WithFile()
		}).
		WithReviews(func(q *ent.AchievementReviewQuery) {
			q.WithReviewer()
		}).
		Order(ent.Desc(entAchievement.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievements: %w", err))
		return nil, 0, err
	}

	// Convert to domain models
	results := make([]achievement.Achievement, 0, len(achievementEntities))
	for _, entity := range achievementEntities {
		ach, err := convertAchievementEntityToDomain(entity)
		if err != nil {
			rec.Add(events.Error, err)
			return nil, 0, err
		}
		results = append(results, ach)
	}

	rec.Add("achievements_count", len(results))
	rec.Add("total_count", total)
	return results, total, nil
}

// GetUserAchievements retrieves all achievements for the current user with pagination.
func (s *SESC) GetUserAchievements(
	ctx context.Context,
	userID UUID,
	offset, limit int,
) ([]achievement.Achievement, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_user_achievements")
	rec.Add("user_id", userID)
	rec.Add("offset", offset)
	rec.Add("limit", limit)

	// Count total achievements for the user
	total, err := s.client.Achievement.Query().
		Where(entAchievement.OwnerID(userID)).
		Count(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to count achievements: %w", err))
		return nil, 0, err
	}

	// Get achievements for the user with pagination
	achievementEntities, err := s.client.Achievement.Query().
		Where(entAchievement.OwnerID(userID)).
		WithTemplate(func(q *ent.AchievementTemplateQuery) {
			q.WithGroup()
		}).
		WithOwner(func(q *ent.UserQuery) {
			q.WithDepartment()
		}).
		WithDocuments(func(q *ent.AchievementDocumentQuery) {
			q.WithFile()
		}).
		WithReviews(func(q *ent.AchievementReviewQuery) {
			q.WithReviewer()
		}).
		Order(ent.Desc(entAchievement.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievements: %w", err))
		return nil, 0, err
	}

	// Convert to domain models
	results := make([]achievement.Achievement, 0, len(achievementEntities))
	for _, entity := range achievementEntities {
		ach, err := convertAchievementEntityToDomain(entity)
		if err != nil {
			rec.Add(events.Error, err)
			return nil, 0, err
		}
		results = append(results, ach)
	}

	rec.Add("achievements_count", len(results))
	rec.Add("total_count", total)
	return results, total, nil
}

// GetGroupedAchievements retrieves all achievements grouped by user with pagination.
func (s *SESC) GetGroupedAchievements(
	ctx context.Context,
	offset, limit int,
) (map[UUID][]achievement.Achievement, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_grouped_achievements")
	rec.Add("offset", offset)
	rec.Add("limit", limit)

	// Count total unique users with achievements
	totalUsers, err := s.client.User.Query().
		Where(user.HasAchievements()).
		Count(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to count users with achievements: %w", err))
		return nil, 0, err
	}

	// Get users with achievements with pagination
	users, err := s.client.User.Query().
		Where(user.HasAchievements()).
		WithDepartment().
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query users with achievements: %w", err))
		return nil, 0, err
	}

	// Create a map to store achievements grouped by user ID
	groupedAchievements := make(map[UUID][]achievement.Achievement)

	// For each user, get their achievements (excluding drafts)
	for _, user := range users {
		achievementEntities, err := s.client.Achievement.Query().
			Where(
				entAchievement.OwnerID(user.ID),
				entAchievement.StatusNEQ(string(achievement.StatusDraft)), // Exclude drafts
			).
			WithTemplate().
			WithOwner(func(q *ent.UserQuery) {
				q.WithDepartment()
			}).
			WithDocuments(func(q *ent.AchievementDocumentQuery) {
				q.WithFile()
			}).
			WithReviews(func(q *ent.AchievementReviewQuery) {
				q.WithReviewer()
			}).
			Order(ent.Desc(entAchievement.FieldID)).
			All(ctx)
		if err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to query achievements for user %s: %w", user.ID, err))
			return nil, 0, err
		}

		// Convert to domain models
		userAchievements := make([]achievement.Achievement, 0, len(achievementEntities))
		for _, entity := range achievementEntities {
			ach, err := convertAchievementEntityToDomain(entity)
			if err != nil {
				rec.Add(events.Error, err)
				return nil, 0, err
			}
			userAchievements = append(userAchievements, ach)
		}

		// Add to the grouped map
		groupedAchievements[user.ID] = userAchievements
	}

	rec.Add("users_count", len(users))
	rec.Add("total_users", totalUsers)
	return groupedAchievements, totalUsers, nil
}

// Helper function to convert achievement entity to domain model
func convertAchievementEntityToDomain(achievementEntity *ent.Achievement) (achievement.Achievement, error) {
	if achievementEntity.Edges.Owner == nil {
		return achievement.Achievement{}, errors.New("achievement owner not loaded")
	}
	if achievementEntity.Edges.Template == nil {
		return achievement.Achievement{}, errors.New("achievement template not loaded")
	}

	// Convert owner
	owner := sesc.User{
		ID:         achievementEntity.Edges.Owner.ID,
		FirstName:  achievementEntity.Edges.Owner.FirstName,
		LastName:   achievementEntity.Edges.Owner.LastName,
		MiddleName: achievementEntity.Edges.Owner.MiddleName,
	}

	// Add department if available
	if achievementEntity.Edges.Owner.Edges.Department != nil {
		owner.Department = sesc.Department{
			ID:          achievementEntity.Edges.Owner.Edges.Department.ID,
			Name:        achievementEntity.Edges.Owner.Edges.Department.Name,
			Description: achievementEntity.Edges.Owner.Edges.Department.Description,
		}
	}

	// Convert template
	template := achievement.Template{
		ID:          achievementEntity.Edges.Template.ID,
		Name:        achievementEntity.Edges.Template.Name,
		Description: achievementEntity.Edges.Template.Description,
		PointsLimit: achievementEntity.Edges.Template.PointsLimit,
		GroupID:     achievementEntity.Edges.Template.GroupID,
		Active:      achievementEntity.Edges.Template.Active,
		Kind:        achievement.Kind(achievementEntity.Edges.Template.Kind),
	}

	// Convert documents
	documents := make([]achievement.Document, 0)
	if achievementEntity.Edges.Documents != nil {
		for _, doc := range achievementEntity.Edges.Documents {
			if doc.Edges.File == nil {
				continue
			}
			documents = append(documents, achievement.Document{
				ID:     doc.ID,
				Name:   doc.Name,
				FileID: doc.Edges.File.ID,
			})
		}
	}

	// Convert reviews
	reviews := make([]achievement.Review, 0)
	if achievementEntity.Edges.Reviews != nil {
		for _, rev := range achievementEntity.Edges.Reviews {
			if rev.Edges.Reviewer == nil {
				continue
			}
			reviewer := sesc.User{
				ID:         rev.Edges.Reviewer.ID,
				FirstName:  rev.Edges.Reviewer.FirstName,
				LastName:   rev.Edges.Reviewer.LastName,
				MiddleName: rev.Edges.Reviewer.MiddleName,
			}
			reviews = append(reviews, achievement.Review{
				ID:             rev.ID,
				From:           reviewer,
				PointsAssigned: rev.PointsAssigned,
				Comment:        rev.Comment,
			})
		}
	}

	// Create the achievement domain model
	result := achievement.Achievement{
		ID:        achievementEntity.ID,
		Owner:     owner,
		Template:  template,
		Status:    achievement.Status(achievementEntity.Status),
		Points:    achievementEntity.Points,
		Documents: documents,
		Reviews:   reviews,
	}

	return result, nil
}

// ReviewAchievement reviews an achievement, setting points and optionally a comment.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in the correct status for review.
func (s *SESC) ReviewAchievement(
	ctx context.Context,
	opt achievement.ReviewOptions,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/review_achievement")
	rec.Set(
		"achievement_owner_id", opt.AchievementOwnerID,
		"achievement_id", opt.AchievementID,
		"reviewer_id", opt.ReviewerID,
		"points_assigned", opt.PointsAssigned,
	)
	if opt.Comment != "" {
		rec.Set("comment", opt.Comment)
	}

	// Start a transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		rec.Set(events.Error, fmt.Errorf("failed to start transaction: %w", err))
		return achievement.Achievement{}, err
	}

	// Get achievement with all related data
	achievementEntity, err := tx.Achievement.Query().
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
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "achievement not found")
		return achievement.Achievement{}, rollback(tx, achievement.ErrAchievementNotFound)
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement: %w", err))
		return achievement.Achievement{}, rollback(tx, err)
	}

	// Get reviewer to ensure they exist
	reviewer, err := tx.User.Get(ctx, opt.ReviewerID)
	if ent.IsNotFound(err) {
		rec.Add(events.Error, "reviewer not found")
		return achievement.Achievement{}, rollback(tx, sesc.ErrUserNotFound)
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to get reviewer: %w", err))
		return achievement.Achievement{}, rollback(tx, err)
	}

	// Check if achievement is in the correct status for review
	currentStatus := achievement.Status(achievementEntity.Status)
	reviewerRole, _ := sesc.RoleByID(reviewer.RoleID)
	templateKind := achievement.Kind(achievementEntity.Edges.Template.Kind)

	// Check if the assigned points exceed the template's limit
	pointsLimit := achievementEntity.Edges.Template.PointsLimit
	if opt.PointsAssigned > pointsLimit {
		rec.Add(
			events.Error,
			fmt.Sprintf("points assigned (%d) exceed template limit (%d)", opt.PointsAssigned, pointsLimit),
		)
		return achievement.Achievement{}, rollback(tx, achievement.ErrPointsLimitExceeded)
	}

	// Determine the new status based on current status, reviewer role, and points
	newStatus, isValidReviewer := determineNewStatus(
		currentStatus,
		reviewerRole,
		templateKind,
		opt.PointsAssigned,
	)

	// If not in a reviewable status
	if currentStatus != achievement.StatusDepheadReview && currentStatus != achievement.StatusInspectorReview {
		rec.Add(events.Error, "achievement is not in a reviewable status")
		return achievement.Achievement{}, rollback(tx, achievement.ErrWrongAchievementStatus)
	}

	// Check if the reviewer has the required role
	if !isValidReviewer {
		rec.Add(events.Error, "reviewer does not have the required role")
		return achievement.Achievement{}, rollback(tx, sesc.ErrInvalidRole)
	}

	// Generate new UUID for review
	reviewID, err := s.newUUID()
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to generate review ID: %w", err))
		return achievement.Achievement{}, rollback(tx, err)
	}

	// Create review
	_, err = tx.AchievementReview.Create().
		SetID(reviewID).
		SetAchievementID(opt.AchievementID).
		SetReviewerID(opt.ReviewerID).
		SetPointsAssigned(opt.PointsAssigned).
		SetComment(opt.Comment).
		Save(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to create review: %w", err))
		return achievement.Achievement{}, rollback(tx, err)
	}

	// Update achievement status and points
	updatedEntity, err := tx.Achievement.UpdateOne(achievementEntity).
		SetStatus(string(newStatus)).
		SetPoints(opt.PointsAssigned).
		Save(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to update achievement: %w", err))
		return achievement.Achievement{}, rollback(tx, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to commit transaction: %w", err))
		return achievement.Achievement{}, err
	}

	// Convert to domain model
	result, err := convertAchievementToModel(achievementEntity, reviewer, reviewID, opt)
	if err != nil {
		rec.Add(events.Error, err)
		return achievement.Achievement{}, err
	}

	// Update the status and points to match the updated entity
	result.Status = achievement.Status(updatedEntity.Status)
	result.Points = updatedEntity.Points

	rec.Add("updated_achievement", result)
	return result, nil
}
