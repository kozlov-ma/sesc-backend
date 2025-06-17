package achsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestAddDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template, achievement, and file
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, sesc.OlympiadDeputy)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		file := testutil.CreateTestFile(ctx, t, client)

		// Call the method being tested
		doc, err := svc.AddDocument(ctx, achievement.AddDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			FileID:        file.ID,
			Name:          "Test Document",
		})

		// Verify the results
		require.NoError(t, err)

		// Verify the document was created
		require.NotEmpty(t, doc.ID)
		require.Equal(t, "Test Document", doc.Name)
		require.Equal(t, file.ID, doc.FileID)
	})

	t.Run("non_existent_achievement", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user and file but use non-existent achievement
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		file := testutil.CreateTestFile(ctx, t, client)
		nonExistentAchievementID := testutil.RandomUUID()

		// Call the method being tested
		_, err := svc.AddDocument(ctx, achievement.AddDocumentOptions{
			AchievementID: nonExistentAchievementID,
			OwnerID:       user.ID,
			FileID:        file.ID,
			Name:          "Test Document",
		})

		// Verify the results
		require.Error(t, err)
		require.Equal(t, achievement.ErrAchievementNotFound, err)
	})

	t.Run("non_existent_file", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template and achievement but use non-existent file
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, sesc.DevelopmentDeputy)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		nonExistentFileID := testutil.RandomUUID()

		// Call the method being tested
		_, err := svc.AddDocument(ctx, achievement.AddDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			FileID:        nonExistentFileID,
			Name:          "Test Document",
		})

		// Verify the results
		require.Error(t, err)
		require.Error(t, err)
	})

	t.Run("approved_achievement", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, reviewer, template, achievement, and file
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, sesc.OlympiadDeputy)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDone)
		file := testutil.CreateTestFile(ctx, t, client)

		// Achievement is already in approved status based on our setup

		// Try to add document to the approved achievement
		_, err := svc.AddDocument(ctx, achievement.AddDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			FileID:        file.ID,
			Name:          "Test Document",
		})

		// Verify error is about wrong achievement status
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template, achievement, and file
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, sesc.OlympiadDeputy)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		file := testutil.CreateTestFile(ctx, t, client)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		_, err := svc.AddDocument(ctx, achievement.AddDocumentOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
			FileID:        file.ID,
			Name:          "Test Document",
		})

		// Verify the results
		require.Error(t, err)
	})
}
