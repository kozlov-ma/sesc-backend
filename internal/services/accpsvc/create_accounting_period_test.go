package accpsvc_test

import (
	"testing"

	accountingperiod "github.com/kozlov-ma/sesc-backend/accounting_period"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/services/accpsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/stretchr/testify/require"
)

func TestCreateAccountingPeriod(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Test accounting period data
		opt := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		// Call the method being tested
		period, err := svc.CreateAccountingPeriod(ctx, opt)

		// Verify the results
		require.NoError(t, err)
		require.NotEmpty(t, period.ID)
		require.Equal(t, opt.Name, period.Name)
		require.Equal(t, opt.Description, *period.Description)
		require.Equal(t, accountingperiod.StatusPlanning, period.Status)

		// Verify period was actually created in the database
		createdPeriod, err := svc.GetAccountingPeriodByID(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, period.ID, createdPeriod.ID)
		require.Equal(t, opt.Name, createdPeriod.Name)
		require.Equal(t, opt.Description, *createdPeriod.Description)
		require.Equal(t, accountingperiod.StatusPlanning, createdPeriod.Status)
	})

	t.Run("duplicate_planning_period", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create the first period
		opt1 := accountingperiod.CreateOptions{
			Name:                   "Q1 2025",
			Description:            "First quarter",
			StartPlanningDate:      "2025-02-01",
			StartAchCollectionDate: "2025-05-12",
			FinishDate:             "2025-07-08",
		}

		firstPeriod, err := svc.CreateAccountingPeriod(ctx, opt1)
		require.NoError(t, err)
		require.NotEmpty(t, firstPeriod.ID)

		// Try to create another planning period
		opt2 := accountingperiod.CreateOptions{
			Name:                   "Q2 2025",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		_, err = svc.CreateAccountingPeriod(ctx, opt2)

		// Verify the results
		require.Error(t, err)
		require.Equal(t, accountingperiod.ErrPlanningPeriodAlreadyExists, err)
	})

	t.Run("invalid_name", func(t *testing.T) {
		// Setup test context with database
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := accpsvc.New(client, nil, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Test accounting period data with empty name
		opt := accountingperiod.CreateOptions{
			Name:                   "",
			Description:            "Second quarter",
			StartPlanningDate:      "2025-08-01",
			StartAchCollectionDate: "2025-11-12",
			FinishDate:             "2026-01-08",
		}

		// Call the method being tested
		_, err := svc.CreateAccountingPeriod(ctx, opt)

		// Verify the results
		require.Error(t, err)
		require.Equal(t, accountingperiod.ErrInvalidAccountingPeriodName, err)
	})
}
