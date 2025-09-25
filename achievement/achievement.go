package achievement

import (
	"github.com/gofrs/uuid/v5"
)

type UUID = uuid.UUID

type Status string

const (
	// StatusDraft is the default achievement status. When Achievement has this status, it can be edited by the teacher
	// that this achievement belongs to. They can attach or remove necessary Documents to confirm their achievement.
	StatusDraft = "draft"

	// StatusDepheadReview is a status that is assigned to the Achievement when the teacher submits it.
	// Now, the achievement should be reviewed by the department head.
	// The department head can approve or request points change.
	StatusDepheadReview = "dephead_review"

	// StatusDepheadPointsChange is a status when department head requests points change.
	// The teacher can modify the points and resubmit.
	StatusDepheadPointsChange = "dephead_points_change"

	// StatusInspectorReview is a status for an Achievement that now should be reviewed by a designated inspector.
	// What inspector should review it is defined by the achievement template kind.
	StatusInspectorReview = "inspector_review"

	// StatusInspectorPointsChange is a status when inspector requests points change.
	// The teacher can modify the points and resubmit.
	StatusInspectorPointsChange = "inspector_points_change"

	// StatusDone could be assigned to the achievement in these cases:
	// 1) Achievement was assigned 0 points
	// 2) Achievement passed the inspector review
	// Then, the Achievement can be used to calculate total points for the user.
	StatusDone = "done"

	// StatusAccounted is assigned to achievements that are done and have been accounted for in reports.
	// This status indicates that the achievement points have been included in financial calculations.
	StatusAccounted = "accounted"
)

type CreateOptions struct {
	ForUserID  UUID
	TemplateID UUID
	Points     int
}

type AddDocumentOptions struct {
	OwnerID       UUID
	AchievementID UUID
	Name          string
	FileID        UUID
}

type RemoveDocumentOptions struct {
	OwnerID       UUID
	AchievementID UUID
	DocumentID    UUID
}

type SubmitOptions struct {
	OwnerID       UUID
	AchievementID UUID
}

type DeleteOptions struct {
	OwnerID       UUID
	AchievementID UUID
}

type ReviewOptions struct {
	AchievementOwnerID UUID
	AchievementID      UUID
	ReviewerID         UUID

	PointsAssigned int
	Comment        string
}

type UpdatePointsOptions struct {
	OwnerID       UUID
	AchievementID UUID
	Points        int
}
