package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// RemoveDocument removes a document from an achievement.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrDocumentNotFound if the document does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *ACS) RemoveDocument(
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
