package achsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestRemoveDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template, achievement, and file
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		file := testutil.CreateTestFile(ctx, t, client)

		// Add a document first
		doc, err := svc.AddDocument(ctx, achievement.AddDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			FileID:        file.ID,
			Name:          "Test Document",
		})
		require.NoError(t, err)

		// Call the method being tested
		err = svc.RemoveDocument(ctx, achievement.RemoveDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			DocumentID:    doc.ID,
		})

		// Verify the results
		require.NoError(t, err)
	})

	t.Run("achievement_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		nonExistentAchievementID := testutil.RandomUUID()
		nonExistentDocumentID := testutil.RandomUUID()

		// Call the method being tested
		err := svc.RemoveDocument(ctx, achievement.RemoveDocumentOptions{
			AchievementID: nonExistentAchievementID,
			OwnerID:       user.ID,
			DocumentID:    nonExistentDocumentID,
		})

		// Verify the error
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("document_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template, and achievement
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		nonExistentDocumentID := testutil.RandomUUID()

		// Call the method being tested
		err := svc.RemoveDocument(ctx, achievement.RemoveDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			DocumentID:    nonExistentDocumentID,
		})

		// Verify the error
		require.ErrorIs(t, err, achievement.ErrDocumentNotFound)
	})

	t.Run("wrong_status", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template, and achievement with non-draft status
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDone)
		nonExistentDocumentID := testutil.RandomUUID()

		// Call the method being tested
		err := svc.RemoveDocument(ctx, achievement.RemoveDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			DocumentID:    nonExistentDocumentID,
		})

		// Verify the error
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})
}
