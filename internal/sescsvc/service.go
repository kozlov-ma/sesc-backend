// Package sescsvc provides services for managing SESC employees and departments.
package sescsvc

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/internal/services/achsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/atsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/depsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/usersvc"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type SESC struct {
	*achsvc.ACS
	*atsvc.ATS
	*depsvc.DES
	*usersvc.USS
}

func New(client *ent.Client) *SESC {
	return &SESC{
		ACS: achsvc.New(client),
		ATS: atsvc.New(client),
		DES: depsvc.New(client),
		USS: usersvc.New(client),
	}
}

type (
	UUID                             = uuid.UUID
	User                             = sesc.User
	Department                       = sesc.Department
	Role                             = sesc.Role
	UserUpdateOptions                = sesc.UserUpdateOptions
	AchievementGroup                 = achievement.Group
	AchievementTemplate              = achievement.Template
	AchievementGroupCreateOptions    = achievement.GroupCreateOptions
	AchievementGroupUpdateOptions    = achievement.GroupUpdateOptions
	AchievementGroupSearchOptions    = achievement.GroupSearchOptions
	AchievementTemplateCreateOptions = achievement.TemplateCreateOptions
	AchievementTemplateUpdateOptions = achievement.TemplateUpdateOptions
	AchievementTemplateSearchOptions = achievement.TemplateSearchOptions
)
