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

// AddDocument adds a document to an achievement.
// Returns achievement.ErrAchievementNotFound if the achievement does not exist.
// Returns achievement.ErrWrongAchievementStatus if the achievement is not in draft status.
func (s *ACS) AddDocument(
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
	err = rec.Operation("create_document", func(opRec *event.Record) error {
		queryStart := time.Now()
		docEntity, err := s.client.AchievementDocument.Create().
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

	result := achievement.Document{
		ID:     documentEntity.ID,
		Name:   documentEntity.Name,
		FileID: fileEntity.ID,
	}

	rec.Sub("result").Set(
		"document_id", result.ID,
		"document_name", result.Name,
		"file_id", result.FileID,
	)

	return result, nil
}
