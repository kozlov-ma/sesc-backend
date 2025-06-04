package achsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestGetUserAchievements(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template, and achievements
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)

		// Create multiple achievements for the user
		ach1 := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		ach2 := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDone)

		// Call the method being tested
		achievements, total, err := svc.GetUserAchievements(ctx, user.ID, 0, 10)

		// Verify the results
		require.NoError(t, err)
		require.Len(t, achievements, 2)
		require.Equal(t, 2, total)

		// Verify that the returned achievements have the correct structure
		achievementIDs := []string{string(ach1.ID), string(ach2.ID)}
		for _, ach := range achievements {
			require.NotEmpty(t, ach.ID)
			require.Contains(t, achievementIDs, string(ach.ID))
			require.Equal(t, user.ID, ach.OwnerID)
			require.Equal(t, template.ID, ach.TemplateID)
		}
	})

	t.Run("no_achievements", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user but no achievements
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		// Call the method being tested
		achievements, total, err := svc.GetUserAchievements(ctx, user.ID, 0, 10)

		// Verify the results
		require.NoError(t, err)
		require.Empty(t, achievements)
		require.Equal(t, 0, total)
	})

	t.Run("pagination", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, template, and multiple achievements
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)

		for i := 0; i < 5; i++ {
			testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		}

		// Call the method being tested with pagination
		achievements, total, err := svc.GetUserAchievements(ctx, user.ID, 0, 2)

		// Verify the results
		require.NoError(t, err)
		require.LessOrEqual(t, len(achievements), 2)
		require.Equal(t, 5, total)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		_, _, err := svc.GetUserAchievements(ctx, user.ID, 0, 10)

		// Verify the results
		require.Error(t, err)
	})
}
