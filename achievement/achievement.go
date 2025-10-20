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
	// The department head can approve, disapprove, or request changes.
	StatusDepheadReview = "dephead_review"

	// StatusDepheadRequestedChanges is assigned when the department head requests changes to the achievement.
	// The teacher must update the achievement points according to the review comment.
	StatusDepheadRequestedChanges = "dephead_requested_changes"

	// StatusInspectorReview is a status for an Achievement that now should be reviewed by a designated inspector.
	// What inspector should review it is defined by the achievement template kind.
	StatusInspectorReview = "inspector_review"

	// StatusInspectorRequestedChanges is assigned when the inspector requests changes to the achievement.
	// The teacher must update the achievement points according to the review comment.
	StatusInspectorRequestedChanges = "inspector_requested_changes"

	// StatusDone could be assigned to the achievement in these cases:
	// 1) Achievement was disapproved by reviewer
	// 2) Achievement passed all reviews successfully
	// Then, the Achievement can be used to calculate total points for the user.
	StatusDone = "done"

	// StatusAccounted is assigned to achievements that are done and have been accounted for in reports.
	// This status indicates that the achievement points have been included in financial calculations.
	StatusAccounted = "accounted"
)

type CreateOptions struct {
	ForUserID  string
	TemplateID UUID
}

type AddDocumentOptions struct {
	OwnerID       string
	AchievementID UUID
	Name          string
	FileID        UUID
}

type RemoveDocumentOptions struct {
	OwnerID       string
	AchievementID UUID
	DocumentID    UUID
}

type SubmitOptions struct {
	OwnerID       string
	AchievementID UUID
}

type DeleteOptions struct {
	OwnerID       string
	AchievementID UUID
}

type ReviewOptions struct {
	AchievementOwnerID string
	AchievementID      UUID
	ReviewerID         string

	Action  ReviewAction
	Comment string
}

type ReviewAction string

const (
	ReviewActionApprove        ReviewAction = "approve"
	ReviewActionDisapprove     ReviewAction = "disapprove"
	ReviewActionRequestChanges ReviewAction = "request_changes"
)

type UpdatePointsOptions struct {
	OwnerID       string
	AchievementID UUID
	Points        int
	Comment       string
}
