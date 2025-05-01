package sesc

// Permission represents an ability of a user to perform a specific action,
// possibly as a part of their Role.
//
// Permissions are predefined in this file.
type Permission struct {
	ID          int32
	Name        string
	Description string
}

var (
	PermissionDraftAchievementList = Permission{
		ID:          1,
		Name:        "draft_achievement_list",
		Description: "Создание и заполнение листа достижений, отправка на проверку, просмотр результатов проверки",
	}
	PermissionDepheadReview = Permission{
		ID:          2,
		Name:        "dephead_review",
		Description: "Проверка и одобрение листа достижений как заведующий кафедрой",
	}
	PermissionContestReview = Permission{
		ID:          3,
		Name:        "contest_review",
		Description: "Проверка достижений, связанных с олимпиадной деятельностью",
	}
	PermissionDevelopmentReview = Permission{
		ID:          4,
		Name:        "development_review",
		Description: "Проверка достижений, связанных с развитием",
	}
	PermissionScientificReview = Permission{
		ID:          5,
		Name:        "scientific_review",
		Description: "Проверка достижений, связанных с научной деятельностью",
	}
)

var Permissions []Permission = []Permission{
	PermissionDraftAchievementList,
	PermissionDepheadReview,
	PermissionContestReview,
	PermissionDevelopmentReview,
	PermissionScientificReview,
}
