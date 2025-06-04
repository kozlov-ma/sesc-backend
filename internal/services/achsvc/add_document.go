package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
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
) (*ent.AchievementDocument, error) {
	rec := event.Get(ctx).Sub("sesc/add_document")
	statsRec := event.Root(ctx).Sub("stats")

	// Group parameters together
	rec.Sub("params").Set(
		"user_id", opt.OwnerID,
		"achievement_id", opt.AchievementID,
		"file_id", opt.FileID,
		"name", opt.Name,
	)

	var doc *ent.AchievementDocument
	err := withTx(ctx, s.client, func(tx *ent.Tx) error {
		var ach *ent.Achievement
		err := rec.Operation("query_achievement", func(rec *event.Record) (err error) {
			rec.Sub("params").Set(
				"achievement_id", opt.AchievementID,
			)

			start := time.Now()
			ach, err = tx.Achievement.Get(ctx, opt.AchievementID)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return achievement.ErrAchievementNotFound
			}

			if err != nil {
				return fmt.Errorf("failed to query achievement: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("validate_achievement_status", func(rec *event.Record) (err error) {
			rec.Set("current_status", ach.Status)
			rec.Set("required_status", string(achievement.StatusDraft))

			if ach.Status != achievement.StatusDraft {
				return achievement.ErrWrongAchievementStatus
			}
			return nil
		})
		if err != nil {
			return err
		}

		var fileEntity *ent.File
		err = rec.Operation("query_file", func(rec *event.Record) error {
			rec.Sub("params").Set("file_id", opt.FileID)

			start := time.Now()
			file, err := tx.File.Get(ctx, opt.FileID)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return sesc.ErrFileNotFound
			}
			if err != nil {
				return fmt.Errorf("failed to get file: %w", err)
			}

			fileEntity = file
			rec.Set("file_name", fileEntity.Name)
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("create_document", func(_ *event.Record) (err error) {
			rec.Sub("params").Set(
				"achievement_id", opt.AchievementID,
				"name", opt.Name,
				"file_id", opt.FileID,
			)

			start := time.Now()
			doc, err = tx.AchievementDocument.Create().
				SetAchievementID(opt.AchievementID).
				SetName(opt.Name).
				SetFileID(opt.FileID).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("failed to create document: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return doc, err
}
