package accpsvc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	accountingperiod "github.com/kozlov-ma/sesc-backend/accounting_period"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entFile "github.com/kozlov-ma/sesc-backend/db/entdb/ent/file"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/services/accpsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/stretchr/testify/require"
)

func TestFileCleanupOnPeriodFinish(t *testing.T) {
	t.Run("current_period_files_are_NOT_deleted_when_period_finishes", func(t *testing.T) {
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		mockFileService := NewMockFileService()

		svc := accpsvc.New(client, mockFileService, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

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
		require.Equal(t, accountingperiod.StatusAchievementCollection, period.Status)

		file1 := createTestFile(t, client, &period.ID, "test1.txt", "content1")
		file2 := createTestFile(t, client, &period.ID, "test2.txt", "content2")
		file3 := createTestFile(t, client, &period.ID, "test3.txt", "content3")

		_, err = client.File.Get(ctx, file1.ID)
		require.NoError(t, err)
		_, err = client.File.Get(ctx, file2.ID)
		require.NoError(t, err)
		_, err = client.File.Get(ctx, file3.ID)
		require.NoError(t, err)

		updatedPeriod, err := svc.FinishPeriod(ctx, period.ID)
		require.NoError(t, err)
		require.Equal(t, accountingperiod.StatusFinished, updatedPeriod.Status)

		// Files from the current period being finished should NOT be deleted
		require.False(t, mockFileService.wasDeleteCalled(file1.ID), "File1 from current period should NOT be deleted")
		require.False(t, mockFileService.wasDeleteCalled(file2.ID), "File2 from current period should NOT be deleted")
		require.False(t, mockFileService.wasDeleteCalled(file3.ID), "File3 from current period should NOT be deleted")
	})

	t.Run("old_files_beyond_retention_threshold_are_deleted", func(t *testing.T) {
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		mockFileService := NewMockFileService()

		// Set retention to keep only 2 most recent periods
		svc := accpsvc.New(client, mockFileService, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

		// Create 4 periods and finish them
		periods := make([]*ent.AccountingPeriod, 4)
		for i := range 4 {
			opt := accountingperiod.CreateOptions{
				Name:                   fmt.Sprintf("Q%d 2025", i+1),
				Description:            fmt.Sprintf("Quarter %d", i+1),
				StartPlanningDate:      fmt.Sprintf("2025-%02d-01", i*3+1),
				StartAchCollectionDate: fmt.Sprintf("2025-%02d-12", i*3+4),
				FinishDate:             fmt.Sprintf("2025-%02d-08", i*3+7),
			}

			period, err := svc.CreateAccountingPeriod(ctx, opt)
			require.NoError(t, err)
			period, err = svc.BeginCollection(ctx, period.ID)
			require.NoError(t, err)
			period, err = svc.FinishPeriod(ctx, period.ID)
			require.NoError(t, err)
			periods[i] = period
		}

		// Create files for each period
		files := make([]*ent.File, 4)
		for i, period := range periods {
			files[i] = createTestFile(t, client, &period.ID, fmt.Sprintf("period%d_file.txt", i+1), "content")
		}

		// Create a new period and finish it to trigger cleanup
		opt := accountingperiod.CreateOptions{
			Name:                   "Q5 2025",
			Description:            "Fifth quarter",
			StartPlanningDate:      "2025-10-01",
			StartAchCollectionDate: "2026-01-12",
			FinishDate:             "2026-04-08",
		}

		newPeriod, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)
		newPeriod, err = svc.BeginCollection(ctx, newPeriod.ID)
		require.NoError(t, err)
		newPeriodFile := createTestFile(t, client, &newPeriod.ID, "new_period_file.txt", "content")
		_, err = svc.FinishPeriod(ctx, newPeriod.ID)
		require.NoError(t, err)

		// With PeriodsToDeleteDocuments=2, we should keep the 2 most recent periods
		// We have 5 finished periods total (4 created + 1 new), so we keep the 2 most recent
		// and delete the 3 oldest. The periods are ordered by ID desc, so:
		// - newPeriod (ID=5) = most recent (keep)
		// - periods[3] (ID=4) = second most recent (keep)
		// - periods[2] (ID=3) = third most recent (delete)
		// - periods[1] (ID=2) = fourth most recent (delete)
		// - periods[0] (ID=1) = oldest (delete)
		require.True(
			t,
			mockFileService.wasDeleteCalled(files[0].ID),
			"File from oldest period should be deleted (beyond retention)",
		)
		require.True(
			t,
			mockFileService.wasDeleteCalled(files[1].ID),
			"File from second oldest period should be deleted (beyond retention)",
		)
		require.True(
			t,
			mockFileService.wasDeleteCalled(files[2].ID),
			"File from third most recent period should be deleted (beyond retention)",
		)
		require.False(
			t,
			mockFileService.wasDeleteCalled(files[3].ID),
			"File from second most recent period should NOT be deleted (within retention)",
		)
		require.False(
			t,
			mockFileService.wasDeleteCalled(newPeriodFile.ID),
			"File from current period should NOT be deleted",
		)
	})

	t.Run("files_without_period_association_are_not_affected", func(t *testing.T) {
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		mockFileService := NewMockFileService()

		svc := accpsvc.New(client, mockFileService, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 2})

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

		fileWithPeriod := createTestFile(t, client, &period.ID, "with_period.txt", "content1")
		fileWithoutPeriod := createTestFile(t, client, nil, "without_period.txt", "content2")

		_, err = svc.FinishPeriod(ctx, period.ID)
		require.NoError(t, err)

		require.False(
			t,
			mockFileService.wasDeleteCalled(fileWithPeriod.ID),
			"File with period should NOT have delete called",
		)
		require.False(
			t,
			mockFileService.wasDeleteCalled(fileWithoutPeriod.ID),
			"File without period should not have delete called",
		)
	})

	t.Run("no_cleanup_when_retention_threshold_not_exceeded", func(t *testing.T) {
		ctx, _ := testutil.CreateTestContext(t)
		client := testutil.SetupDatabase(t)

		mockFileService := NewMockFileService()

		// Set retention to keep 5 periods, but only create 2
		svc := accpsvc.New(client, mockFileService, &config.AccountingPeriodConfig{PeriodsToDeleteDocuments: 5})

		// Create and finish 2 periods
		for i := range 2 {
			opt := accountingperiod.CreateOptions{
				Name:                   fmt.Sprintf("Q%d 2025", i+1),
				Description:            fmt.Sprintf("Quarter %d", i+1),
				StartPlanningDate:      fmt.Sprintf("2025-%02d-01", i*3+1),
				StartAchCollectionDate: fmt.Sprintf("2025-%02d-12", i*3+4),
				FinishDate:             fmt.Sprintf("2025-%02d-08", i*3+7),
			}

			period, err := svc.CreateAccountingPeriod(ctx, opt)
			require.NoError(t, err)
			period, err = svc.BeginCollection(ctx, period.ID)
			require.NoError(t, err)
			_, err = svc.FinishPeriod(ctx, period.ID)
			require.NoError(t, err)
		}

		// Create a third period and finish it
		opt := accountingperiod.CreateOptions{
			Name:                   "Q3 2025",
			Description:            "Third quarter",
			StartPlanningDate:      "2025-07-01",
			StartAchCollectionDate: "2025-10-12",
			FinishDate:             "2026-01-08",
		}

		period, err := svc.CreateAccountingPeriod(ctx, opt)
		require.NoError(t, err)
		period, err = svc.BeginCollection(ctx, period.ID)
		require.NoError(t, err)
		periodFile := createTestFile(t, client, &period.ID, "period3_file.txt", "content")
		_, err = svc.FinishPeriod(ctx, period.ID)
		require.NoError(t, err)

		// No files should be deleted because we haven't exceeded the retention threshold
		// and current period files are not deleted
		require.False(
			t,
			mockFileService.wasDeleteCalled(periodFile.ID),
			"File from current period should NOT be deleted",
		)
	})
}

func createTestFile(t *testing.T, client *ent.Client, periodID *int, name, content string) *ent.File {
	ctx := t.Context()

	file, err := client.File.Create().
		SetS3ObjectKey("test-key-" + name).
		SetName(name).
		SetSize(len(content)).
		SetNillableAccountingPeriodID(periodID).
		Save(ctx)
	require.NoError(t, err)

	return file
}

type MockFileService struct {
	deletedFiles map[string]bool
}

func NewMockFileService() *MockFileService {
	return &MockFileService{
		deletedFiles: make(map[string]bool),
	}
}

func (m *MockFileService) Delete(_ context.Context, id uuid.UUID) error {
	// Track the deletion call
	m.deletedFiles[id.String()] = true
	return nil
}

func (m *MockFileService) wasDeleteCalled(id uuid.UUID) bool {
	return m.deletedFiles[id.String()]
}

// Simple mock file service that actually deletes from database
type SimpleMockFileService struct {
	client     *ent.Client
	deletedIDs map[string]bool
}

func (m *SimpleMockFileService) Delete(ctx context.Context, id uuid.UUID) error {
	// Track the deletion
	if m.deletedIDs == nil {
		m.deletedIDs = make(map[string]bool)
	}
	m.deletedIDs[id.String()] = true

	// Actually delete from database
	if m.client != nil {
		_, err := m.client.File.Delete().Where(entFile.ID(id)).Exec(ctx)
		return err
	}

	return nil
}
