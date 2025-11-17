package achsvc

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyservice"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementtemplate"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/predicate"
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
	client  *ent.Client
	company companyservice.S
}

func New(client *ent.Client, company companyservice.S) *ACS {
	return &ACS{
		client:  client,
		company: company,
	}
}

// buildRoleBasedFilters creates appropriate filters based on the asking user's role
func (s *ACS) buildRoleBasedFilters(
	askingUser company.User,
	requireChanges bool,
) predicate.Achievement {
	baseFilter := s.buildBaseRoleFilter(askingUser)

	if requireChanges {
		changesFilter := entAchievement.Or(
			entAchievement.Status(achievement.StatusDepheadRequestedChanges),
			entAchievement.Status(achievement.StatusInspectorRequestedChanges),
		)
		return entAchievement.And(baseFilter, changesFilter)
	}

	return baseFilter
}

// buildBaseRoleFilter creates base filters based on the asking user's role
func (s *ACS) buildBaseRoleFilter(askingUser company.User) predicate.Achievement {
	var filters []predicate.Achievement
	if askingUser.HasRole(company.Teacher) {
		filters = append(filters, entAchievement.OwnerID(askingUser.ID))
	}

	if askingUser.HasRole(company.Dephead) && askingUser.DepartmentID != "" {
		filters = append(filters, entAchievement.And(
			entAchievement.DepartmentID(askingUser.DepartmentID),
			entAchievement.StatusNotIn(
				achievement.StatusDraft,
				achievement.StatusAccounted,
			),
		))
	}

	if rr := askingUser.RolesIn(company.ScientificDeputy, company.OlympiadDeputy, company.DevelopmentDeputy, company.AcademicDirector); len(
		rr,
	) > 0 {
		for _, r := range rr {
			filters = append(filters,
				entAchievement.And(
					entAchievement.StatusNotIn(
						achievement.StatusDraft,
						achievement.StatusAccounted,
						achievement.StatusDepheadReview,
					),
					entAchievement.HasTemplateWith(achievementtemplate.ReviewerRole(r)),
				),
			)
		}
	}

	if askingUser.HasRole(company.ChiefEconomist) {
		filters = append(filters, entAchievement.Or(
			entAchievement.Status(achievement.StatusDone),
			entAchievement.Status(achievement.StatusAccounted),
		),
		)
	}

	return entAchievement.Or(filters...)
}
