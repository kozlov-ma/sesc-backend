package deletiondaemon

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/filesvc"
	"github.com/kozlov-ma/sesc-backend/internal/filesvc/mocks"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestDeletionDaemon(t *testing.T) {
	t.Run("processes_scheduled_deletions", func(t *testing.T) {
		// Setup
		client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
		defer client.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := mocks.NewMockObjectStorage(ctrl)
		fileService := filesvc.New(client, storage, "test-bucket", time.Millisecond*100)

		cfg := &config.DeletionDaemonConfig{
			Enabled:  true,
			Interval: time.Millisecond * 100,
		}

		daemon := New(fileService, cfg)

		// Create and schedule files for deletion
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")

		files := make([]string, 2)
		for i := range 2 {
			content := []byte(fmt.Sprintf("test file %d", i))

			storage.EXPECT().
				PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).
				Return(nil)

			file, err := fileService.Create(ctx, bytes.NewReader(content), sesc.FileCreateOptions{
				FileName: fmt.Sprintf("test%d.txt", i),
				FileSize: len(content),
			})
			require.NoError(t, err)

			// Update to schedule for deletion in the past
			_, err = client.File.UpdateOneID(file.ID).
				SetDeletionScheduled(true).
				SetScheduledDeletionAt(time.Now().Add(-time.Hour)).
				Save(ctx)
			require.NoError(t, err)

			files[i] = *file.S3ObjectKey
		}

		// Setup storage expectations
		for _, key := range files {
			storage.EXPECT().RemoveObject(gomock.Any(), key).Return(nil)
		}

		// Run daemon for a short time
		ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond*300)
		defer cancel()

		done := make(chan struct{})
		go func() {
			ctx, _ := event.NewRecord(ctx, "daemon_test")
			daemon.Start(ctx)
			close(done)
		}()

		// Wait for context to finish
		<-ctx.Done()
		daemon.Stop()
		<-done

		// Verify files were processed
		verifyCtx := t.Context()
		verifyCtx, _ = event.NewRecord(verifyCtx, "verify")
		stats, _, err := fileService.GetFileStats(verifyCtx)
		require.NoError(t, err)
		require.Equal(t, 2, stats.DeletedFiles, "Both files should be marked as deleted")
		require.Equal(t, 0, stats.ReadyForDeletion, "No files should be ready for deletion")
	})

	t.Run("daemon_disabled", func(t *testing.T) {
		client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
		defer client.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := mocks.NewMockObjectStorage(ctrl)
		fileService := filesvc.New(client, storage, "test-bucket", time.Hour)

		cfg := &config.DeletionDaemonConfig{
			Enabled:  false,
			Interval: time.Millisecond * 100,
		}

		daemon := New(fileService, cfg)

		// Create and schedule file for deletion
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")

		content := []byte("test file")
		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).Return(nil)

		file, err := fileService.Create(ctx, bytes.NewReader(content), sesc.FileCreateOptions{
			FileName: "test.txt",
			FileSize: len(content),
		})
		require.NoError(t, err)

		_, err = client.File.UpdateOneID(file.ID).
			SetDeletionScheduled(true).
			SetScheduledDeletionAt(time.Now().Add(-time.Hour)).
			Save(ctx)
		require.NoError(t, err)

		// Run daemon (should exit immediately since it's disabled)
		ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond*200)
		defer cancel()

		done := make(chan struct{})
		go func() {
			ctx, _ := event.NewRecord(ctx, "daemon_test")
			daemon.Start(ctx)
			close(done)
		}()

		<-done // Should finish quickly since daemon is disabled

		// Verify file was NOT processed
		verifyCtx := t.Context()
		verifyCtx, _ = event.NewRecord(verifyCtx, "verify")
		stats, _, err := fileService.GetFileStats(verifyCtx)
		require.NoError(t, err)
		require.Equal(t, 0, stats.DeletedFiles, "File should not be deleted")
		require.Equal(t, 1, stats.ReadyForDeletion, "File should still be ready for deletion")
	})

	t.Run("stops_on_signal", func(t *testing.T) {
		client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
		defer client.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := mocks.NewMockObjectStorage(ctrl)
		fileService := filesvc.New(client, storage, "test-bucket", time.Hour)

		cfg := &config.DeletionDaemonConfig{
			Enabled:  true,
			Interval: time.Second * 10, // Long interval
		}

		daemon := New(fileService, cfg)

		ctx, cancel := context.WithTimeout(t.Context(), time.Second*5)
		defer cancel()

		done := make(chan struct{})
		go func() {
			ctx, _ := event.NewRecord(ctx, "daemon_test")
			daemon.Start(ctx)
			close(done)
		}()

		// Stop daemon after a short delay
		time.Sleep(time.Millisecond * 100)
		daemon.Stop()

		// Should finish quickly
		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("Daemon did not stop in time")
		}
	})
}
