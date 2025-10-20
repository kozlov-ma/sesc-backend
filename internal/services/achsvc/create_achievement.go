package achsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// ACCESSTODO
// CreateAchievement creates a new achievement for a user based on a template.
// Returns achievement.ErrAchievementTemplateNotFound if the template does not exist.
func (s *ACS) CreateAchievement(
	ctx context.Context,
	opt achievement.CreateOptions,
) (*ent.Achievement, error) {
	rec := event.Get(ctx).Sub("achsvc/create_achievement")
	rec.Sub("params").Set(
		"user_id", opt.ForUserID,
		"template_id", opt.TemplateID,
	)

	statsRec := event.Root(ctx).Sub("stats")

	var deptID string
	err := rec.Operation("get_author_department", func(_ *event.Record) error {
		author, err := s.company.User(ctx, companyquery.User{ID: opt.ForUserID})
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if id := author.DepartmentID; id == "" {
			return fmt.Errorf(
				"user must be assigned to a department to create achievements: %w",
				achievement.ErrCannotCreateAchievement,
			)
		}

		deptID = author.DepartmentID

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement author: %w", err)
	}

	var ach *ent.Achievement
	err = txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		var template *ent.AchievementTemplate
		err := rec.Operation("get_template", func(rec *event.Record) error {
			rec.Sub("params").Set("template_id", opt.TemplateID)

			start := time.Now()
			tmpl, err := tx.AchievementTemplate.Get(ctx, opt.TemplateID)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if ent.IsNotFound(err) {
				return achievement.ErrAchievementTemplateNotFound
			} else if err != nil {
				return fmt.Errorf("failed to get template: %w", err)
			}

			template = tmpl
			rec.Set("template", template)
			return nil
		})
		if err != nil {
			return err
		}

		err = rec.Operation("create_achievement", func(_ *event.Record) error {
			start := time.Now()
			achievement, err := tx.Achievement.Create().
				SetOwnerID(opt.ForUserID).
				SetDepartmentID(deptID).
				SetTemplateID(opt.TemplateID).
				SetStatus(string(achievement.StatusDraft)).
				SetPoints(template.PointsLimit).
				Save(ctx)
			statsRec.Add(events.PostgresQueries, 1)
			statsRec.Add(events.PostgresTime, time.Since(start))

			if err != nil {
				return fmt.Errorf("couldn't save achievement: %w", err)
			}

			ach = achievement
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return ach, err
}
