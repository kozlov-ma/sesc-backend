package accpsvc_test

import (
	"testing"

	accountingperiod "github.com/kozlov-ma/sesc-backend/accounting_period"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/services/accpsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/stretchr/testify/require"
)

func TestGetAccountingPeriods(t *testing.T) {
	t.Run("get_all_periods", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create multiple periods
		opt1 := accountingperiod.CreateOptions{
			Name:                   "Q1 2025",
			Description:            "First quarter",
			StartPlanningDate:      "2025-02-01",
			StartAchCollectionDate: "2025-05-12",
			FinishDate:             "2025-07-08",
		}

		opt2 := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period1, err := svc.CreateAccountingPeriod(ctx, opt1)
		require.NoError(t, err)

		// Cancel the first period to create another planning period
		_, err = svc.CancelPeriod(ctx, period1.ID)
		require.NoError(t, err)

		period2, err := svc.CreateAccountingPeriod(ctx, opt2)
		require.NoError(t, err)

		// Get all periods
		periods, err := svc.GetAccountingPeriods(ctx)
		require.NoError(t, err)
		require.Len(t, periods, 2)

		// Verify periods are ordered by start planning date (descending)
		require.Equal(t, period2.ID, periods[0].ID)
		require.Equal(t, period1.ID, periods[1].ID)
	})

	t.Run("get_current_period", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create a period and transition to achievement collection
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)

		period, err = svc.BeginCollection(ctx, period.ID)
		require.NoError(t, err)

		// Get current period
		currentPeriod, err := svc.GetCurrentAccountingPeriod(ctx)
		require.NoError(t, err)
		require.Equal(t, period.ID, currentPeriod.ID)
		require.Equal(t, accountingperiod.StatusAchievementCollection, currentPeriod.Status)
	})

	t.Run("get_planning_period", func(t *testing.T) {
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

		// Get planning period
		planningPeriod, err := svc.GetPlanningAccountingPeriod(ctx)
		require.NoError(t, err)
		require.Equal(t, period.ID, planningPeriod.ID)
		require.Equal(t, accountingperiod.StatusPlanning, planningPeriod.Status)
	})

	t.Run("get_period_by_id", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create a period
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)

		// Get period by ID
		retrievedPeriod, err := svc.GetAccountingPeriodByID(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, period.ID, retrievedPeriod.ID)
		require.Equal(t, opt.Name, retrievedPeriod.Name)
		require.Equal(t, opt.Description, *retrievedPeriod.Description)
	})

	t.Run("period_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Try to get a non-existent period
		_, err := svc.GetCurrentAccountingPeriod(ctx)
		require.Error(t, err)
		require.Equal(t, accountingperiod.ErrAccountingPeriodNotFound, err)

		_, err = svc.GetPlanningAccountingPeriod(ctx)
		require.Error(t, err)
		require.Equal(t, accountingperiod.ErrAccountingPeriodNotFound, err)
	})
}
