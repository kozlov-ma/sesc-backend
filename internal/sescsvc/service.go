// Package sescsvc provides services for managing SESC employees and departments.
package sescsvc

import (
	"github.com/gofrs/uuid/v5"
	accountingperiod "github.com/kozlov-ma/sesc-backend/accounting_period"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/filesvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/accpsvc"
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
	*accpsvc.ACCPS
}

func New(client *ent.Client, fileService *filesvc.FileService, cfg *config.Config) *SESC {
	return &SESC{
		ACS:   achsvc.New(client),
		ATS:   atsvc.New(client),
		DES:   depsvc.New(client),
		USS:   usersvc.New(client),
		ACCPS: accpsvc.New(client, fileService, &cfg.AccountingPeriod),
	}
}

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
	AccountingPeriodCreateOptions    = accountingperiod.CreateOptions
	AccountingPeriodUpdateOptions    = accountingperiod.UpdateOptions
)
