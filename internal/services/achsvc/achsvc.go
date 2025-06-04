package achsvc

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementtemplate"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/predicate"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	UUID                             = uuid.UUID
	Role                             = sesc.Role
	UserUpdateOptions                = sesc.UserUpdateOptions
	AchievementGroupCreateOptions    = achievement.GroupCreateOptions
	AchievementGroupUpdateOptions    = achievement.GroupUpdateOptions
	AchievementGroupSearchOptions    = achievement.GroupSearchOptions
	AchievementTemplateCreateOptions = achievement.TemplateCreateOptions
	AchievementTemplateUpdateOptions = achievement.TemplateUpdateOptions
	AchievementTemplateSearchOptions = achievement.TemplateSearchOptions
)

type ACS struct {
	client *ent.Client
}

func New(client *ent.Client) *ACS {
	return &ACS{
		client: client,
	}
}

func withTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %w", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// buildRoleBasedFilters creates appropriate filters based on the asking user's role
func (s *ACS) buildRoleBasedFilters(askingUser *ent.User) predicate.Achievement {
	switch askingUser.Role {
	case sesc.Teacher:
		return entAchievement.OwnerID(askingUser.ID)
	case sesc.Dephead:
		if askingUser.DepartmentID != nil {
			return entAchievement.And(
				entAchievement.HasOwnerWith(user.DepartmentID(askingUser.Edges.Department.ID)),
				entAchievement.StatusNotIn(
					achievement.StatusDraft,
					achievement.StatusAccounted,
				),
			)
		}
	case sesc.ScientificDeputy:
		return entAchievement.And(
			entAchievement.StatusNotIn(
				achievement.StatusDraft,
				achievement.StatusAccounted,
				achievement.StatusDepheadReview,
			),
			entAchievement.HasTemplateWith(achievementtemplate.Kind(achievement.Scientific)),
		)
	case sesc.OlympiadDeputy:
		return entAchievement.And(
			entAchievement.StatusNotIn(
				achievement.StatusDraft,
				achievement.StatusAccounted,
				achievement.StatusDepheadReview,
			),
			entAchievement.HasTemplateWith(achievementtemplate.Kind(achievement.Olympiad)),
		)
	case sesc.DevelopmentDeputy:
		return entAchievement.And(
			entAchievement.StatusNotIn(
				achievement.StatusDraft,
				achievement.StatusAccounted,
				achievement.StatusDepheadReview,
			),
			entAchievement.HasTemplateWith(achievementtemplate.Kind(achievement.Development)),
		)
	case sesc.AcademicDirector:
		return entAchievement.And(
			entAchievement.StatusNotIn(
				achievement.StatusDraft,
				achievement.StatusAccounted,
				achievement.StatusDepheadReview,
			),
			entAchievement.HasTemplateWith(achievementtemplate.Kind(achievement.Development)),
		)
	case sesc.ChiefEconomist:
		return entAchievement.Or(
			entAchievement.Status(achievement.StatusDone),
			entAchievement.Status(achievement.StatusAccounted),
		)
	}

	panic("invalid role")
}
