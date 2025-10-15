package filesvc

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestScheduleDeletion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, _ := setupFileService(t)

		// Create test files
		fileIDs := make([]uuid.UUID, 3)
		for i := range 3 {
			content := []byte(fmt.Sprintf("test file %d", i))
			reader := bytes.NewReader(content)

			opts := FileOpts{
				FileName: fmt.Sprintf("test%d.txt", i),
				FileSize: len(content),
			}

			storage.EXPECT().
				PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).
				Return(nil)

			file, err := svc.Create(ctx, reader, opts)
			require.NoError(t, err)
			fileIDs[i] = file.ID
		}

		// Schedule first two files for deletion
		opts := sesc.ScheduleDeletionOptions{
			FileIDs: fileIDs[:2],
		}

		err := svc.ScheduleDeletion(ctx, opts)
		require.NoError(t, err)

		// Verify first two files are scheduled
		for i := range 2 {
			file, err := svc.ByID(ctx, fileIDs[i])
			require.NoError(t, err)
			require.True(t, file.DeletionScheduled, "File %d should be scheduled for deletion", i)
			require.NotNil(t, file.ScheduledDeletionAt, "File %d should have scheduled deletion time", i)
			require.True(
				t,
				file.ScheduledDeletionAt.After(time.Now()),
				"Scheduled deletion time should be in the future",
			)
		}

		// Verify third file is NOT scheduled
		file, err := svc.ByID(ctx, fileIDs[2])
		require.NoError(t, err)
		require.False(t, file.DeletionScheduled, "File 2 should NOT be scheduled for deletion")
		require.Nil(t, file.ScheduledDeletionAt)
	})

	t.Run("empty_file_ids", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, _, _ := setupFileService(t)

		opts := sesc.ScheduleDeletionOptions{
			FileIDs: []uuid.UUID{},
		}

		err := svc.ScheduleDeletion(ctx, opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no file IDs provided")
	})
}

func TestProcessScheduledDeletions(t *testing.T) {
	t.Run("process_ready_files", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, client := setupFileService(t)

		// Create files and schedule them for deletion
		files := make([]uuid.UUID, 3)
		objectKeys := make([]string, 3)

		for i := range 3 {
			content := []byte(fmt.Sprintf("test file %d", i))
			reader := bytes.NewReader(content)

			opts := FileOpts{
				FileName: fmt.Sprintf("test%d.txt", i),
				FileSize: len(content),
			}

			storage.EXPECT().
				PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).
				Return(nil)

			file, err := svc.Create(ctx, reader, opts)
			require.NoError(t, err)
			files[i] = file.ID
			objectKeys[i] = *file.S3ObjectKey

			// Schedule for deletion with past time (ready for deletion)
			_, err = client.File.UpdateOneID(file.ID).
				SetDeletionScheduled(true).
				SetScheduledDeletionAt(time.Now().Add(-time.Hour)).
				Save(ctx)
			require.NoError(t, err)
		}

		// Set expectations for storage deletion
		for _, key := range objectKeys {
			storage.EXPECT().RemoveObject(gomock.Any(), key).Return(nil)
		}

		// Process scheduled deletions
		err := svc.ProcessScheduledDeletions(ctx)
		require.NoError(t, err)

		// Verify files are marked as deleted and S3 keys are cleared
		for i, fileID := range files {
			file, err := svc.ByID(ctx, fileID)
			require.NoError(t, err)
			require.True(t, file.FileDeleted, "File %d should be marked as deleted", i)
			require.Nil(t, file.S3ObjectKey, "File %d S3ObjectKey should be cleared", i)
			require.False(t, file.DeletionScheduled, "File %d should no longer be scheduled", i)
			require.Nil(t, file.ScheduledDeletionAt, "File %d scheduled deletion time should be cleared", i)
		}
	})

	t.Run("ignore_future_scheduled_files", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, client := setupFileService(t)

		// Create file scheduled for future deletion
		content := []byte("test file")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "future.txt",
			FileSize: len(content),
		}

		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).Return(nil)

		file, err := svc.Create(ctx, reader, opts)
		require.NoError(t, err)

		// Schedule for deletion with future time
		_, err = client.File.UpdateOneID(file.ID).
			SetDeletionScheduled(true).
			SetScheduledDeletionAt(time.Now().Add(time.Hour)).
			Save(ctx)
		require.NoError(t, err)

		// Process scheduled deletions - should not affect this file
		err = svc.ProcessScheduledDeletions(ctx)
		require.NoError(t, err)

		// Verify file is still scheduled and not deleted
		retrievedFile, err := svc.ByID(ctx, file.ID)
		require.NoError(t, err)
		require.False(t, retrievedFile.FileDeleted, "File should NOT be marked as deleted")
		require.True(t, retrievedFile.DeletionScheduled, "File should still be scheduled")
		require.NotNil(t, retrievedFile.S3ObjectKey, "S3ObjectKey should NOT be cleared")
	})

	t.Run("no_files_to_process", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, _, _ := setupFileService(t)

		// Process when no files are scheduled
		err := svc.ProcessScheduledDeletions(ctx)
		require.NoError(t, err)
	})

	t.Run("storage_error_continues_processing", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, client := setupFileService(t)

		// Create two files
		files := make([]uuid.UUID, 2)
		objectKeys := make([]string, 2)

		for i := range 2 {
			content := []byte(fmt.Sprintf("test file %d", i))
			reader := bytes.NewReader(content)

			opts := FileOpts{
				FileName: fmt.Sprintf("test%d.txt", i),
				FileSize: len(content),
			}

			storage.EXPECT().
				PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).
				Return(nil)

			file, err := svc.Create(ctx, reader, opts)
			require.NoError(t, err)
			files[i] = file.ID
			objectKeys[i] = *file.S3ObjectKey

			// Schedule for deletion with past time
			_, err = client.File.UpdateOneID(file.ID).
				SetDeletionScheduled(true).
				SetScheduledDeletionAt(time.Now().Add(-time.Hour)).
				Save(ctx)
			require.NoError(t, err)
		}

		// First file deletion succeeds, second fails
		storage.EXPECT().RemoveObject(gomock.Any(), objectKeys[0]).Return(nil)
		storage.EXPECT().RemoveObject(gomock.Any(), objectKeys[1]).Return(errors.New("storage error"))

		// Process should continue despite error
		err := svc.ProcessScheduledDeletions(ctx)
		require.NoError(t, err)

		// First file should be deleted
		file1, err := svc.ByID(ctx, files[0])
		require.NoError(t, err)
		require.True(t, file1.FileDeleted, "File 0 should be marked as deleted")

		// Second file should still be scheduled (not deleted due to storage error)
		file2, err := svc.ByID(ctx, files[1])
		require.NoError(t, err)
		require.False(t, file2.FileDeleted, "File 1 should NOT be marked as deleted")
		require.True(t, file2.DeletionScheduled, "File 1 should still be scheduled")
	})
}

func TestGetFileStats(t *testing.T) {
	t.Run("empty_database", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, _, _ := setupFileService(t)

		stats, _, err := svc.GetFileStats(ctx)
		require.NoError(t, err)
		require.NotNil(t, stats)
		require.Equal(t, 0, stats.TotalFiles)
		require.Equal(t, 0, stats.DeletedFiles)
		require.Equal(t, 0, stats.ScheduledForDeletion)
		require.Equal(t, 0, stats.ReadyForDeletion)
		require.Equal(t, 0, stats.NotScheduled)
	})

	t.Run("various_file_states", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, client := setupFileService(t)

		// Create normal file
		content1 := []byte("normal file")
		storage.EXPECT().
			PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content1)))).
			Return(nil)
		_, err := svc.Create(ctx, bytes.NewReader(content1), FileOpts{
			FileName: "normal.txt",
			FileSize: len(content1),
		})
		require.NoError(t, err)

		// Create deleted file
		_, err = client.File.Create().
			SetName("deleted.txt").
			SetSize(100).
			SetFileDeleted(true).
			Save(ctx)
		require.NoError(t, err)

		// Create scheduled file (future)
		content2 := []byte("scheduled file")
		storage.EXPECT().
			PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content2)))).
			Return(nil)
		scheduledFile, err := svc.Create(ctx, bytes.NewReader(content2), FileOpts{
			FileName: "scheduled.txt",
			FileSize: len(content2),
		})
		require.NoError(t, err)
		_, err = client.File.UpdateOneID(scheduledFile.ID).
			SetDeletionScheduled(true).
			SetScheduledDeletionAt(time.Now().Add(time.Hour)).
			Save(ctx)
		require.NoError(t, err)

		// Create ready file (past deadline)
		content3 := []byte("ready file")
		storage.EXPECT().
			PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content3)))).
			Return(nil)
		readyFile, err := svc.Create(ctx, bytes.NewReader(content3), FileOpts{
			FileName: "ready.txt",
			FileSize: len(content3),
		})
		require.NoError(t, err)
		_, err = client.File.UpdateOneID(readyFile.ID).
			SetDeletionScheduled(true).
			SetScheduledDeletionAt(time.Now().Add(-time.Hour)).
			Save(ctx)
		require.NoError(t, err)

		// Get stats
		stats, _, err := svc.GetFileStats(ctx)
		require.NoError(t, err)
		require.Equal(t, 4, stats.TotalFiles)
		require.Equal(t, 1, stats.DeletedFiles)
		require.Equal(t, 2, stats.ScheduledForDeletion)
		require.Equal(t, 1, stats.ReadyForDeletion)
		require.Equal(t, 1, stats.NotScheduled)
	})
}
