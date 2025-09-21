package accpsvc_test

import (
	"testing"

	accountingperiod "github.com/kozlov-ma/sesc-backend/accounting_period"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/services/accpsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/stretchr/testify/require"
)

func TestAccountingPeriodStatusTransitions(t *testing.T) {
	t.Run("planning_to_achievement_collection", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create a planning period
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusPlanning, period.Status)

		// Transition to achievement collection
		updatedPeriod, err := svc.BeginCollection(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusAchievementCollection, updatedPeriod.Status)
	})

	t.Run("planning_to_cancelled", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create a planning period
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusPlanning, period.Status)

		// Cancel the period
		updatedPeriod, err := svc.CancelPeriod(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusCancelled, updatedPeriod.Status)
		require.NotNil(t, updatedPeriod.CancelDate)
	})

	t.Run("achievement_collection_to_finished", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create a planning period
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)

		// Transition to achievement collection
		period, err = svc.BeginCollection(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusAchievementCollection, period.Status)

		// Finish the period
		updatedPeriod, err := svc.FinishPeriod(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusFinished, updatedPeriod.Status)
	})

	t.Run("achievement_collection_to_not_executed", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create a planning period
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)

		// Transition to achievement collection
		period, err = svc.BeginCollection(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusAchievementCollection, period.Status)

		// Mark as not executed
		updatedPeriod, err := svc.MarkAsNonExecuted(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusNotExecuted, updatedPeriod.Status)
		require.NotNil(t, updatedPeriod.BecameNonExecutedDate)
	})

	t.Run("invalid_transition_from_planning_to_finished", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create a planning period
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusPlanning, period.Status)

		// Try to finish the period directly (invalid transition)
		_, err = svc.FinishPeriod(ctx, period.ID)
		require.Error(t, err)
		require.Equal(t, accountingperiod.ErrInvalidStatusTransition, err)
	})

	t.Run("cannot_have_multiple_active_periods", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create first period and transition to achievement collection
		opt1 := accountingperiod.CreateOptions{
			Name:                   "Q1 2025",
			Description:            "First quarter",
			StartPlanningDate:      "2025-02-01",
			StartAchCollectionDate: "2025-05-12",
			FinishDate:             "2025-07-08",
		}

		period1, err := svc.CreateAccountingPeriod(ctx, opt1)
		require.NoError(t, err)

		period1, err = svc.BeginCollection(ctx, period1.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusAchievementCollection, period1.Status)

		// Create second period
		opt2 := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period2, err := svc.CreateAccountingPeriod(ctx, opt2)
		require.NoError(t, err)

		// Try to transition second period to achievement collection (should fail)
		_, err = svc.BeginCollection(ctx, period2.ID)
		require.Error(t, err)
		require.Equal(t, accountingperiod.ErrActivePeriodAlreadyExists, err)
	})
}
