package achsvc

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
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

func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return fmt.Errorf("%w: rolling back transaction: %w", err, rerr)
	}
	return err
}

// buildRoleBasedFilters creates appropriate filters based on the asking user's role
func (s *ACS) buildRoleBasedFilters(askingUser *ent.User) predicate.Achievement {
	switch askingUser.Role {
	case sesc.Dephead:
		// Department head: filter achievements from their department with DepheadReview status
		if askingUser.Edges.Department != nil {
			return entAchievement.And(
				entAchievement.HasOwnerWith(user.DepartmentID(askingUser.Edges.Department.ID)),
				entAchievement.Status(string(achievement.StatusDepheadReview)),
			)
		}
	case sesc.OlympiadDeputy:
		// Olympiad deputy: filter achievements with InspectorReview status and Olympiad kind
		return entAchievement.And(
			entAchievement.Status(string(achievement.StatusInspectorReview)),
			entAchievement.HasTemplateWith(func(tq *ent.AchievementTemplateQuery) {
				tq.Where(func(tq *ent.AchievementTemplateQuery) {
					tq.HasGroupWith(func(gq *ent.AchievementGroupQuery) {
						gq.Kind(string(achievement.KindOlympiad))
					})
				})
			}),
		)
	case sesc.AcademicDirector:
		// Academic director: filter achievements with InspectorReview status and Development kind
		return entAchievement.And(
			entAchievement.Status(string(achievement.StatusInspectorReview)),
			entAchievement.HasTemplateWith(func(tq *ent.AchievementTemplateQuery) {
				tq.Where(func(tq *ent.AchievementTemplateQuery) {
					tq.HasGroupWith(func(gq *ent.AchievementGroupQuery) {
						gq.Kind(string(achievement.KindDevelopment))
					})
				})
			}),
		)
	default:
		// For other roles, don't apply any special filters (show all achievements)
		// Return a predicate that matches all achievements
		return entAchievement.IDNotNil()
	}

	// Fallback for cases where conditions are not met
	return entAchievement.IDNotNil()
}
