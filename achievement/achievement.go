package achievement

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type UUID = uuid.UUID

// A Document is a file attached to some Achievement as confirmation.
type Document struct {
	ID     UUID
	Name   string
	FileID UUID
}

// A Review is a comment from reviewer setting the points to the Achievement.
// If PointsAssigned will decrease the current achievement points, the Comment field
// must be added.
type Review struct {
	ID             UUID
	From           sesc.User
	PointsAssigned int
	Comment        string
}

type Status string

const (
	// StatusDraft is the default achievement status. When Achievement has this status, it can be edited by the teacher
	// that this achievement belongs to. They can attach or remove necessary Documents to confirm their achievement.
	StatusDraft = "draft"

	// StatusDepheadReview is a status that is assigned to the Achievement when the teacher submits it.
	// Now, the achievement should be reviewed by the department head.
	// If the department head assigns >0 points to the achievement, it is then passed to the inspector.
	StatusDepheadReview = "dephead_review"

	// StatusInspectorReview is a status for an Achievement that now should be reviewed by a designated inspector.
	// What inspector should review it is defined by the achievement template kind.
	StatusInspectorReview = "inspector_review"

	// StatusDone could be assigned to the achievement in these cases:
	// 1) Achievement was assigned 0 points
	// 2) Achievement passed the inspector review
	// Then, the Achievement can be used to calculate total points for the user.
	StatusDone = "done"

	// StatusAccounted is assigned to achievements that are done and have been accounted for in reports.
	// This status indicates that the achievement points have been included in financial calculations.
	StatusAccounted = "accounted"
)

// An Achievement is an report of achievement that belongs to a SESC teacher.
// Achievements need to be reviewed by multiple people to confirm the points added to the teacher.
type Achievement struct {
	ID       UUID
	Owner    sesc.User
	Template Template

	Status    Status
	Points    int
	Documents []Document
	Reviews   []Review
}

type CreateOptions struct {
	ForUser    sesc.User
	TemplateID UUID
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
