package respond

import (
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
)

type AccountingPeriod struct {
	ID                     int        `json:"id"                        example:"2599"                 description:"Unique identifier of the accounting period"                                                        validate:"required"`
	Name                   string     `json:"name"                      example:"Q2 2025"              description:"Name of the accounting period"                                                                     validate:"required"`
	Description            string     `json:"description"               example:"Second quarter"       description:"Description of the accounting period"                                                              validate:"required"`
	StartPlanningDate      *time.Time `json:"start_planning_date"       example:"2025-08-01T00:00:00Z" description:"Start date for planning phase"`
	StartAchCollectionDate *time.Time `json:"start_ach_collection_date" example:"2025-11-12T00:00:00Z" description:"Start date for achievement collection phase"`
	FinishDate             *time.Time `json:"finish_date"               example:"2026-01-08T00:00:00Z" description:"End date of the accounting period"`
	CancelDate             *time.Time `json:"cancel_date"               example:"2025-12-11T00:00:00Z" description:"Date when the period was cancelled (if applicable)"`
	BecameNonExecutedDate  *time.Time `json:"became_non_executed_date"  example:"2025-12-15T00:00:00Z" description:"Date when the period was marked as not executed (if applicable)"`
	Status                 string     `json:"status"                    example:"planning"             description:"Current status of the accounting period (planning, achcollect, finished, cancelled, not_executed)" validate:"required"`
}

type AccountingPeriods struct {
	AccountingPeriods []*AccountingPeriod `json:"accounting_periods" description:"List of accounting periods"         validate:"required"`
	Total             int                 `json:"total"              description:"Total number of accounting periods" validate:"required"`
}

func WithAccountingPeriod(ap *ent.AccountingPeriod) *AccountingPeriod {
	if ap == nil {
		return nil
	}
	var description string
	if ap.Description != nil {
		description = *ap.Description
	}

	return &AccountingPeriod{
		ID:                     ap.ID,
		Name:                   ap.Name,
		Description:            description,
		StartPlanningDate:      ap.StartPlanningDate,
		StartAchCollectionDate: ap.StartAchievementCollectionDate,
		FinishDate:             ap.FinishDate,
		CancelDate:             ap.CancelDate,
		BecameNonExecutedDate:  ap.BecameNonExecutedDate,
		Status:                 ap.Status,
	}
}

func WithAccountingPeriods(aps []*ent.AccountingPeriod, total int) AccountingPeriods {
	periods := make([]*AccountingPeriod, len(aps))
	for i, ap := range aps {
		periods[i] = WithAccountingPeriod(ap)
	}

	return AccountingPeriods{
		AccountingPeriods: periods,
		Total:             total,
	}
}
