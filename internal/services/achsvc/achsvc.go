package achsvc

import (
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
	case sesc.ScientificDeputy, sesc.OlympiadDeputy, sesc.DevelopmentDeputy, sesc.AcademicDirector:
		return entAchievement.And(
			entAchievement.StatusNotIn(
				achievement.StatusDraft,
				achievement.StatusAccounted,
				achievement.StatusDepheadReview,
			),
			entAchievement.HasTemplateWith(achievementtemplate.ReviewerRole(askingUser.Role)),
		)
	case sesc.ChiefEconomist:
		return entAchievement.Or(
			entAchievement.Status(achievement.StatusDone),
			entAchievement.Status(achievement.StatusAccounted),
		)
	}

	panic("invalid role")
}
