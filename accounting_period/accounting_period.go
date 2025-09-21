package accountingperiod

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Status string

const (
	// Accounting Period behaviour is FSM with following states and transitions:
	// Planning:
	// - To AchievementCollection
	// - To Cancelled
	// AchievementCollection:
	// - To Finished
	// - To NotExecuted

	// StatusPlanning -- period that can be started by Administrator. In bounds of
	// each accounting period some roles can add period-specific content to this one
	// such as AchievementTemplate or Document, but "teacher" roles can't post achievements
	// to review before the next period.
	// - There can be only one planning period.
	// - Files that were loaded n periods ago(n is a number from config) become deleted,
	// but db entries remaining for history consistency
	StatusPlanning = "planning"

	// StatusAchievementCollection -- period where teachers can post achievements to review.
	//
	// - There can be only one active period
	// - After transition from this state period becomes invariant
	StatusAchievementCollection = "achcollect"

	// StatusCancelled -- only Administrator can cancel the period. After cancellation period becomes immutable
	StatusCancelled = "cancelled"

	// StatusNotExecuted -- only Administrator can set the period as not executed. After setting the period as not executed period becomes immutable
	StatusNotExecuted = "nonexecuted"

	// StatusFinished means that the period is finished
	StatusFinished = "finished"
)

// CreateOptions represents options for creating a new accounting period
type CreateOptions struct {
	Name                   string
	Description            string
	StartPlanningDate      string
	StartAchCollectionDate string
	FinishDate             string
}

// Validate validates the create options
func (o CreateOptions) Validate() error {
	if o.Name == "" {
		return ErrInvalidAccountingPeriodName
	}
	return nil
}

// UpdateOptions represents options for updating an accounting period
type UpdateOptions struct {
	Name                   string
	Description            string
	StartPlanningDate      string
	StartAchCollectionDate string
	FinishDate             string
}

// Validate validates the update options
func (o UpdateOptions) Validate() error {
	if o.Name == "" {
		return ErrInvalidAccountingPeriodName
	}
	return nil
}

// StatusTransition represents a status transition
type StatusTransition struct {
	From Status
	To   Status
}

// ValidTransitions defines the allowed status transitions
var ValidTransitions = []StatusTransition{
	{From: StatusPlanning, To: StatusAchievementCollection},
	{From: StatusPlanning, To: StatusCancelled},
	{From: StatusAchievementCollection, To: StatusFinished},
	{From: StatusAchievementCollection, To: StatusNotExecuted},
}

// IsValidTransition checks if a transition from one status to another is valid
func IsValidTransition(from, to Status) bool {
	for _, transition := range ValidTransitions {
		if transition.From == from && transition.To == to {
			return true
		}
	}
	return false
}

// GetStatusDisplayName returns the display name for a status
func GetStatusDisplayName(status Status) string {
	switch status {
	case StatusPlanning:
		return "Планирование"
	case StatusAchievementCollection:
		return "Сбор достижений"
	case StatusCancelled:
		return "Отменен"
	case StatusNotExecuted:
		return "Не исполнен"
	case StatusFinished:
		return "Завершен"
	default:
		return "Неизвестный статус"
	}
}

// PeriodInfo represents basic period information
type PeriodInfo struct {
	ID                     uuid.UUID  `json:"id"`
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	StartPlanningDate      *time.Time `json:"start_planning_date,omitempty"`
	StartAchCollectionDate *time.Time `json:"start_ach_collection_date,omitempty"`
	FinishDate             *time.Time `json:"finish_date,omitempty"`
	CancelDate             *time.Time `json:"cancel_date,omitempty"`
	BecameNonExecutedDate  *time.Time `json:"became_non_executed_date,omitempty"`
	Status                 Status     `json:"status"`
}
