package achsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// CreateAchievement creates a new achievement for a user based on a template.
// Returns achievement.ErrAchievementTemplateNotFound if the template does not exist.
func (s *ACS) CreateAchievement(
	ctx context.Context,
	opt achievement.CreateOptions,
) (achievement.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/create_achievement")
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
	var template *ent.AchievementTemplate
	err := rec.Operation("get_template", func(opRec *event.Record) error {
		opRec.Sub("params").Set("template_id", opt.TemplateID)

		queryStart := time.Now()
		tmpl, err := s.client.AchievementTemplate.Get(ctx, opt.TemplateID)
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
		ID:    achievementEntity.ID,
		Owner: opt.ForUser,
		Template: achievement.Template{
			ID:          template.ID,
			Name:        template.Name,
			Description: template.Description,
			PointsLimit: template.PointsLimit,
			GroupID:     template.GroupID,
			Active:      template.Active,
			Kind:        template.Kind,
		},
		Status: achievement.Status(achievementEntity.Status),
		Points: achievementEntity.Points,
	}

	rec.Set("created_achievement", result)
	return result, nil
}
