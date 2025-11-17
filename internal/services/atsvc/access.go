package atsvc

import (
	"github.com/kozlov-ma/sesc-backend/company"
)

// ViewAchievementTemplatesAction allows all authenticated users to view achievement templates and groups.
type ViewAchievementTemplatesAction struct{}

func NewViewAchievementTemplatesAction() ViewAchievementTemplatesAction {
	return ViewAchievementTemplatesAction{}
}

func (a ViewAchievementTemplatesAction) AllowsUser(_ company.User) bool {
	return true
}

// HandleAchievementTemplatesAction allows only admin to manage achievement templates and groups.
type HandleAchievementTemplatesAction struct{}

func NewHandleAchievementTemplatesAction() HandleAchievementTemplatesAction {
	return HandleAchievementTemplatesAction{}
}

func (a HandleAchievementTemplatesAction) AllowsUser(u company.User) bool {
	return u.HasRole(company.Admin)
}
