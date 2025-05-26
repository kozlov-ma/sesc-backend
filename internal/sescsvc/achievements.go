// Package sescsvc provides services for managing SESC employees, departments, and achievements.
package sescsvc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
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
	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.ForUser.ID,
		"template_id", opt.TemplateID,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Check if template exists
	var template achievement.Template
	err := rec.Operation("get_template", func(opRec *event.Record) error {
		opRec.Sub("params").Set("template_id", opt.TemplateID)

		queryStart := time.Now()
		tmpl, err := s.AchievementTemplateByID(ctx, opt.TemplateID)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to get template: %w", err))
			return err
		}

		template = tmpl
		opRec.Set("template", template)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Generate new UUID for achievement
	var id UUID
	err = rec.Operation("generate_uuid", func(opRec *event.Record) error {
		uuid, err := s.newUUID()
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to generate ID: %w", err))
			return err
		}
		id = uuid
		opRec.Set("generated_id", id)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Create achievement in database
	var achievementEntity *ent.Achievement
	err = rec.Operation("create_achievement", func(opRec *event.Record) error {
		// Start a transaction
		queryStart := time.Now()
		tx, err := s.client.Tx(ctx)
		queryCount++
		opRec.Add("tx_start_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to start transaction: %w", err))
			return err
		}

		// Create achievement
		queryStart = time.Now()
		entity, err := tx.Achievement.Create().
			SetID(id).
			SetOwnerID(opt.ForUser.ID).
			SetTemplateID(opt.TemplateID).
			SetStatus(string(achievement.StatusDraft)).
			SetPoints(template.PointsLimit).
			Save(ctx)
		queryCount++
		opRec.Add("create_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to create achievement: %w", err))
			return rollback(tx, err)
		}

		// Commit transaction
		queryStart = time.Now()
		if err := tx.Commit(); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to commit transaction: %w", err))
			return err
		}
		queryCount++
		opRec.Add("commit_time_ms", time.Since(queryStart).Milliseconds())

		achievementEntity = entity
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	result := achievement.Achievement{
		ID:       achievementEntity.ID,
		Owner:    opt.ForUser,
		Template: template,
		Status:   achievement.Status(achievementEntity.Status),
		Points:   achievementEntity.Points,
	}

	rec.Set("created_achievement", result)
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

	// Get achievement to check status
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
		opRec.Set("status", achievementEntity.Status)
		return nil
	})
	if err != nil {
		return err
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
		return err
	}

	// Delete achievement and its documents
	var tx *ent.Tx
	err = rec.Operation("delete_achievement", func(opRec *event.Record) error {
		// Start a transaction
		queryStart := time.Now()
		txn, err := s.client.Tx(ctx)
		queryCount++
		opRec.Add("tx_start_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to start transaction: %w", err))
			return err
		}
		tx = txn

		// Delete achievement documents
		queryStart = time.Now()
		result, err := tx.AchievementDocument.Delete().
			Where(achievementdocument.AchievementID(opt.AchievementID)).
			Exec(ctx)
		queryCount++
		opRec.Add("delete_documents_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to delete achievement documents: %w", err))
			return rollback(tx, err)
		}
		opRec.Set("documents_deleted", result)

		// Delete achievement
		queryStart = time.Now()
		result, err = tx.Achievement.Delete().
			Where(
				entAchievement.ID(opt.AchievementID),
				entAchievement.OwnerID(opt.OwnerID),
			).
			Exec(ctx)
		queryCount++
		opRec.Add("delete_achievement_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to delete achievement: %w", err))
			return rollback(tx, err)
		}
		opRec.Set("achievements_deleted", result)

		return nil
	})
	if err != nil {
		return err
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
		return err
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"success", true,
		"achievement_id", opt.AchievementID,
	)

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

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.OwnerID,
		"achievement_id", opt.AchievementID,
		"file_id", opt.FileID,
		"name", opt.Name,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Get achievement to check status
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
		opRec.Set("status", achievementEntity.Status)
		return nil
	})
	if err != nil {
		return achievement.Document{}, err
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
		return achievement.Document{}, err
	}

	// Verify file exists
	var fileEntity *ent.File
	err = rec.Operation("verify_file", func(opRec *event.Record) error {
		opRec.Sub("params").Set("file_id", opt.FileID)

		queryStart := time.Now()
		file, err := s.client.File.Get(ctx, opt.FileID)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if ent.IsNotFound(err) {
			opRec.Add(events.Error, "file not found")
			return sesc.ErrFileNotFound
		}
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to get file: %w", err))
			return err
		}

		fileEntity = file
		opRec.Set("file_name", fileEntity.Name)
		return nil
	})
	if err != nil {
		return achievement.Document{}, err
	}

	// Generate ID and create document
	var documentEntity *ent.AchievementDocument
	var documentID UUID
	err = rec.Operation("create_document", func(opRec *event.Record) error {
		// Generate new UUID for document
		id, err := s.newUUID()
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to generate ID: %w", err))
			return err
		}
		documentID = id
		opRec.Set("document_id", documentID)

		// Create document
		queryStart := time.Now()
		docEntity, err := s.client.AchievementDocument.Create().
			SetID(documentID).
			SetAchievementID(opt.AchievementID).
			SetName(opt.Name).
			SetFileID(opt.FileID).
			Save(ctx)
		queryCount++
		opRec.Add("create_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to create document: %w", err))
			return err
		}

		documentEntity = docEntity
		return nil
	})
	if err != nil {
		return achievement.Document{}, err
	}

	// Prepare result
	result := achievement.Document{
		ID:     documentEntity.ID,
		Name:   documentEntity.Name,
		FileID: fileEntity.ID,
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"document_id", result.ID,
		"document_name", result.Name,
		"file_id", result.FileID,
	)

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

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.OwnerID,
		"achievement_id", opt.AchievementID,
		"document_id", opt.DocumentID,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Get achievement to check status
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
		opRec.Set("status", achievementEntity.Status)
		return nil
	})
	if err != nil {
		return err
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
		return err
	}

	// Verify document exists and belongs to the achievement
	var documentEntity *ent.AchievementDocument
	err = rec.Operation("verify_document", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"document_id", opt.DocumentID,
			"achievement_id", opt.AchievementID,
		)

		queryStart := time.Now()
		document, err := s.client.AchievementDocument.Query().
			Where(
				achievementdocument.ID(opt.DocumentID),
				achievementdocument.AchievementID(opt.AchievementID),
			).
			Only(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if ent.IsNotFound(err) {
			opRec.Add(events.Error, "document not found")
			return achievement.ErrDocumentNotFound
		}
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query document: %w", err))
			return err
		}

		documentEntity = document
		opRec.Set("document_name", documentEntity.Name)
		return nil
	})
	if err != nil {
		return err
	}

	// Delete the document
	err = rec.Operation("delete_document", func(opRec *event.Record) error {
		opRec.Sub("params").Set("document_id", documentEntity.ID)

		queryStart := time.Now()
		err := s.client.AchievementDocument.DeleteOne(documentEntity).Exec(ctx)
		queryCount++
		opRec.Add("delete_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to delete document: %w", err))
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"success", true,
		"document_id", opt.DocumentID,
		"achievement_id", opt.AchievementID,
	)

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
			Kind:        achievement.Kind(tmpl.Kind),
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert owner
	err = rec.Operation("convert_owner", func(opRec *event.Record) error {
		usr, err := convertUser(achievementEntity.Edges.Owner)
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to convert owner: %w", err))
			return err
		}
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
			reviewer, err := convertUser(rev.Edges.Reviewer)
			if err != nil {
				opRec.Add(events.Error, fmt.Errorf("failed to convert reviewer: %w", err))
				return err
			}

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
		"reviewer_role_id", reviewerRole.ID,
		"reviewer_role_name", reviewerRole.Name,
		"template_kind", templateKind,
		"points_assigned", pointsAssigned,
	)

	// We'll return the results directly instead of using a struct

	switch currentStatus {
	case achievement.StatusDepheadReview:
		// Department head review
		if reviewerRole.ID == sesc.Dephead.ID {
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
			"expected_role", sesc.Dephead.ID,
			"is_valid_reviewer", false,
		)

	case achievement.StatusInspectorReview:
		// Inspector review - check if reviewer has the correct role based on template kind
		expectedRole := templateKind.InspectorRole()
		rec.Sub("inspector_check").Set(
			"expected_role_id", expectedRole.ID,
			"expected_role_name", expectedRole.Name,
		)

		if reviewerRole.ID == expectedRole.ID {
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
) (achievement.Achievement, error) {
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
		Kind:        achievement.Kind(achievementEntity.Edges.Template.Kind),
	}

	// Convert owner
	owner, err := convertUser(achievementEntity.Edges.Owner)
	if err != nil {
		subRec.Add(events.Error, fmt.Errorf("failed to convert owner: %w", err))
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
	subRec.Set("documents_count", len(documents))

	// Convert existing reviews
	reviews := make([]achievement.Review, 0, len(achievementEntity.Edges.Reviews)+1)
	for _, rev := range achievementEntity.Edges.Reviews {
		revUser, err := convertUser(rev.Edges.Reviewer)
		if err != nil {
			subRec.Add(events.Error, fmt.Errorf("failed to convert reviewer: %w", err))
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
		subRec.Add(events.Error, fmt.Errorf("failed to convert reviewer: %w", err))
		return achievement.Achievement{}, fmt.Errorf("failed to convert reviewer: %w", err)
	}

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
	return result, nil
}

// GetAchievement retrieves an achievement by ID and owner ID.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
func (s *SESC) GetAchievement(
	ctx context.Context,
	achievementID UUID,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("sesc/get_achievement")
	// Group parameters together
	rec.Sub("params").Set("achievement_id", achievementID)

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
		opRec.Sub("params").Set("achievement_id", achievementID)

		queryStart := time.Now()
		entity, err := s.client.Achievement.Query().
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
		opRec.Set("found", true)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert to domain model
	var result achievement.Achievement
	err = rec.Operation("convert_to_domain", func(opRec *event.Record) error {
		domainModel, err := convertAchievementEntityToDomain(achievementEntity, opRec)
		if err != nil {
			opRec.Add(events.Error, err)
			return err
		}
		result = domainModel
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	rec.Set("achievement", result)
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
		ach, err := convertAchievementEntityToDomain(entity, rec)
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

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", userID,
		"offset", offset,
		"limit", limit,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Count total achievements for the user
	var totalAchievements int
	err := rec.Operation("count_achievements", func(opRec *event.Record) error {
		opRec.Sub("params").Set("user_id", userID)

		queryStart := time.Now()
		count, err := s.client.Achievement.Query().
			Where(entAchievement.OwnerID(userID)).
			Count(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to count achievements: %w", err))
			return err
		}

		totalAchievements = count
		opRec.Set("total_count", totalAchievements)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Get achievements with pagination
	var achievementEntities []*ent.Achievement
	err = rec.Operation("query_achievements", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"user_id", userID,
			"offset", offset,
			"limit", limit,
		)

		queryStart := time.Now()
		entities, err := s.client.Achievement.Query().
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
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query achievements: %w", err))
			return err
		}

		achievementEntities = entities
		opRec.Set("achievements_count", len(achievementEntities))
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Convert to domain models
	result := make([]achievement.Achievement, 0, len(achievementEntities))
	err = rec.Operation("convert_achievements", func(opRec *event.Record) error {
		opRec.Set("entities_count", len(achievementEntities))

		for i, entity := range achievementEntities {
			achRec := opRec.Sub(fmt.Sprintf("achievement_%d", i))
			achRec.Set("id", entity.ID)

			ach, err := convertAchievementEntityToDomain(entity, achRec)
			if err != nil {
				achRec.Add(events.Error, err)
				return err
			}
			result = append(result, ach)
		}

		opRec.Set("converted_count", len(result))
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"achievements_count", len(result),
		"total_achievements", totalAchievements,
		"user_id", userID,
	)

	return result, totalAchievements, nil
}

// GetGroupedAchievements retrieves all achievements grouped by user with pagination.
func (s *SESC) GetGroupedAchievements(
	ctx context.Context,
	offset, limit int,
) (map[UUID][]achievement.Achievement, int, error) {
	rec := event.Get(ctx).Sub("sesc/get_grouped_achievements")

	// Group parameters together
	rec.Sub("params").Set(
		"offset", offset,
		"limit", limit,
	)

	// Track stats in root record
	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	// Count total unique users with achievements
	var totalUsers int
	err := rec.Operation("count_users", func(opRec *event.Record) error {
		queryStart := time.Now()
		count, err := s.client.User.Query().
			Where(user.HasAchievements()).
			Count(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to count users with achievements: %w", err))
			return err
		}

		totalUsers = count
		opRec.Set("total_users", totalUsers)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Get users with achievements with pagination
	var users []*ent.User
	err = rec.Operation("query_users", func(opRec *event.Record) error {
		opRec.Sub("params").Set(
			"offset", offset,
			"limit", limit,
		)

		queryStart := time.Now()
		userList, err := s.client.User.Query().
			Where(user.HasAchievements()).
			WithDepartment().
			Offset(offset).
			Limit(limit).
			All(ctx)
		queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query users with achievements: %w", err))
			return err
		}

		users = userList
		opRec.Set("users_count", len(users))
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Create a map to store achievements grouped by user ID
	groupedAchievements := make(map[UUID][]achievement.Achievement)

	// For each user, get their achievements (excluding drafts)
	err = rec.Operation("get_user_achievements", func(opRec *event.Record) error {
		opRec.Set("users_count", len(users))

		for i, usr := range users {
			userRec := opRec.Sub(fmt.Sprintf("user_%d", i))
			userRec.Set("user_id", usr.ID)

			// Query achievements for this user
			queryStart := time.Now()
			achievementEntities, err := s.client.Achievement.Query().
				Where(
					entAchievement.OwnerID(usr.ID),
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
			queryCount++
			userRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

			if err != nil {
				userRec.Add(events.Error, fmt.Errorf("failed to query achievements for user %s: %w", usr.ID, err))
				return err
			}
			userRec.Set("achievements_count", len(achievementEntities))

			// Convert to domain models
			userAchievements := make([]achievement.Achievement, 0, len(achievementEntities))
			for j, entity := range achievementEntities {
				achRec := userRec.Sub(fmt.Sprintf("achievement_%d", j))
				achRec.Set("id", entity.ID)

				ach, err := convertAchievementEntityToDomain(entity, achRec)
				if err != nil {
					achRec.Add(events.Error, err)
					return err
				}
				userAchievements = append(userAchievements, ach)
			}

			// Add to the grouped map
			groupedAchievements[usr.ID] = userAchievements
			userRec.Set("converted_count", len(userAchievements))
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Record successful outcome
	rec.Sub("result").Set(
		"users_count", len(users),
		"total_users", totalUsers,
		"total_groups", len(groupedAchievements),
	)

	return groupedAchievements, totalUsers, nil
}

// Helper function to convert achievement entity to domain model
func convertAchievementEntityToDomain(
	achievementEntity *ent.Achievement,
	rec *event.Record,
) (achievement.Achievement, error) {
	// Group parameters
	rec.Sub("params").Set("achievement_id", achievementEntity.ID)

	// Validate required relationships
	if achievementEntity.Edges.Owner == nil {
		rec.Add(events.Error, "achievement owner not loaded")
		return achievement.Achievement{}, errors.New("achievement owner not loaded")
	}
	if achievementEntity.Edges.Template == nil {
		rec.Add(events.Error, "achievement template not loaded")
		return achievement.Achievement{}, errors.New("achievement template not loaded")
	}

	// Result to be populated by operations
	var (
		owner     sesc.User
		template  achievement.Template
		documents []achievement.Document
		reviews   []achievement.Review
	)

	// Convert owner
	err := rec.Operation("convert_owner", func(opRec *event.Record) error {
		opRec.Sub("params").Set("owner_id", achievementEntity.Edges.Owner.ID)

		owner = sesc.User{
			ID:         achievementEntity.Edges.Owner.ID,
			FirstName:  achievementEntity.Edges.Owner.FirstName,
			LastName:   achievementEntity.Edges.Owner.LastName,
			MiddleName: achievementEntity.Edges.Owner.MiddleName,
		}

		// Add department if available
		if achievementEntity.Edges.Owner.Edges.Department != nil {
			dept := achievementEntity.Edges.Owner.Edges.Department
			opRec.Sub("department").Set(
				"id", dept.ID,
				"name", dept.Name,
			)

			owner.Department = sesc.Department{
				ID:          dept.ID,
				Name:        dept.Name,
				Description: dept.Description,
			}
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert template
	err = rec.Operation("convert_template", func(opRec *event.Record) error {
		tmpl := achievementEntity.Edges.Template
		opRec.Sub("params").Set(
			"template_id", tmpl.ID,
			"kind", tmpl.Kind,
			"points_limit", tmpl.PointsLimit,
		)

		template = achievement.Template{
			ID:          tmpl.ID,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			PointsLimit: tmpl.PointsLimit,
			GroupID:     tmpl.GroupID,
			Active:      tmpl.Active,
			Kind:        achievement.Kind(tmpl.Kind),
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert documents
	err = rec.Operation("convert_documents", func(opRec *event.Record) error {
		documents = make([]achievement.Document, 0)
		if achievementEntity.Edges.Documents != nil {
			opRec.Set("total_documents", len(achievementEntity.Edges.Documents))

			validDocs := 0
			for _, doc := range achievementEntity.Edges.Documents {
				if doc.Edges.File == nil {
					opRec.Add(events.Error, fmt.Sprintf("file not loaded for document %s", doc.ID))
					continue
				}
				documents = append(documents, achievement.Document{
					ID:     doc.ID,
					Name:   doc.Name,
					FileID: doc.Edges.File.ID,
				})
				validDocs++
			}
			opRec.Set("valid_documents", validDocs)
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert reviews
	err = rec.Operation("convert_reviews", func(opRec *event.Record) error {
		reviews = make([]achievement.Review, 0)
		if achievementEntity.Edges.Reviews != nil {
			opRec.Set("total_reviews", len(achievementEntity.Edges.Reviews))

			validReviews := 0
			for _, rev := range achievementEntity.Edges.Reviews {
				if rev.Edges.Reviewer == nil {
					opRec.Add(events.Error, fmt.Sprintf("reviewer not loaded for review %s", rev.ID))
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
				validReviews++
			}
			opRec.Set("valid_reviews", validReviews)
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
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

	// Record final result stats
	rec.Sub("result").Set(
		"status", result.Status,
		"points", result.Points,
		"documents_count", len(documents),
		"reviews_count", len(reviews),
	)

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
		opRec.Set("reviewer_role_id", reviewer.RoleID)
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Validate review parameters
	err = rec.Operation("validate_review", func(opRec *event.Record) error {
		// Check if achievement is in the correct status for review
		currentStatus := achievement.Status(achievementEntity.Status)
		reviewerRole, _ := sesc.RoleByID(reviewer.RoleID)
		templateKind := achievement.Kind(achievementEntity.Edges.Template.Kind)

		opRec.Sub("params").Set(
			"current_status", currentStatus,
			"reviewer_role_id", reviewerRole.ID,
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
		id, err := s.newUUID()
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
		domainModel, err := convertAchievementToModel(achievementEntity, reviewer, reviewID, opt, opRec)
		if err != nil {
			opRec.Add(events.Error, err)
			return err
		}

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
