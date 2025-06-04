package achsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestGetUsersWithAchievements(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test users, template, and achievements
		user1 := testutil.CreateTestUser(ctx, t, client, "Test1", "User1", sesc.Role(1))
		user2 := testutil.CreateTestUser(ctx, t, client, "Test2", "User2", sesc.Role(1))
		asker := testutil.CreateTestUser(ctx, t, client, "Asker", "User", sesc.Dephead)
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)

		// Create achievements for both users
		testutil.CreateTestAchievement(ctx, t, client, user1, template, achievement.StatusDraft)
		testutil.CreateTestAchievement(ctx, t, client, user2, template, achievement.StatusDone)

		// Call the method being tested
		users, total, err := svc.GetUsersWithAchievements(ctx, asker.ID, 0, 10)

		// Verify the results
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(users), 2)
		require.GreaterOrEqual(t, total, 2)

		// Verify that the returned users have the correct structure
		for _, user := range users {
			require.NotEmpty(t, user.ID)
			require.NotEmpty(t, user.FirstName)
			require.NotEmpty(t, user.LastName)
		}
	})

	t.Run("no_users_with_achievements", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create an asker but no users with achievements
		asker := testutil.CreateTestUser(ctx, t, client, "Asker", "User", sesc.Dephead)

		// Call the method being tested
		users, total, err := svc.GetUsersWithAchievements(ctx, asker.ID, 0, 10)

		// Verify the results
		require.NoError(t, err)
		require.Empty(t, users)
		require.Equal(t, 0, total)
	})

	t.Run("pagination", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create multiple test users with achievements
		asker := testutil.CreateTestUser(ctx, t, client, "Asker", "User", sesc.Dephead)
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)

		for i := 0; i < 5; i++ {
			user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
			testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		}

		// Call the method being tested with pagination
		users, total, err := svc.GetUsersWithAchievements(ctx, asker.ID, 0, 2)

		// Verify the results
		require.NoError(t, err)
		require.LessOrEqual(t, len(users), 2)
		require.GreaterOrEqual(t, total, 5)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create an asker
		asker := testutil.CreateTestUser(ctx, t, client, "Asker", "User", sesc.Dephead)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		_, _, err := svc.GetUsersWithAchievements(ctx, asker.ID, 0, 10)

		// Verify the results
		require.Error(t, err)
	})
}
