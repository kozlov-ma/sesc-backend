package sesc

import "slices"

// AchievementDocument deletion status constants
const (
	AchievementDocumentStatusActive            = "active"
	AchievementDocumentStatusScheduledDeletion = "scheduled_deletion"
	AchievementDocumentStatusDeleted           = "deleted"
)

// ValidAchievementDocumentStatuses returns all valid status values
func ValidAchievementDocumentStatuses() []string {
	return []string{
		AchievementDocumentStatusActive,
		AchievementDocumentStatusScheduledDeletion,
		AchievementDocumentStatusDeleted,
	}
}

// IsValidAchievementDocumentStatus checks if the given status is valid
func IsValidAchievementDocumentStatus(status string) bool {
	return slices.Contains(ValidAchievementDocumentStatuses(), status)
}
