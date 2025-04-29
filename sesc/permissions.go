package sesc

import "github.com/gofrs/uuid/v5"

// Permission represents an ability of a user to perform a specific action,
// possibly as a part of their Role.
//
// Permissions are predefined in this file.
type Permission struct {
	ID          UUID
	Name        string
	Description string
}

var (
	PermissionDraftAchievementList = Permission{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "draft_achievement_list",
		Description: "Создание и заполнение листа достижений, отправка на проверку",
	}
	PermissionDepheadReview = Permission{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "dephead_review",
		Description: "Проверка и одобрение листа достижений как заведующий кафедрой",
	}
	PermissionContestReview = Permission{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "contest_review",
		Description: "Проверка достижений, связанных с олимпиадной деятельностью",
	}
	PermissionDevelopmentReview = Permission{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "development_review",
		Description: "Проверка достижений, связанных с развитием",
	}
	PermissionScientificReview = Permission{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "scientific_review",
		Description: "Проверка достижений, связанных с научной деятельностью",
	}
)
