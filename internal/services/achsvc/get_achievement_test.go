package achsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestGetAchievement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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

		// Call the method being tested
		result, err := svc.GetAchievement(ctx, ach.ID)

		// Verify the results
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, ach.ID, result.ID)
		require.Equal(t, user.ID, result.OwnerID)
		require.Equal(t, template.ID, result.TemplateID)
	})

	t.Run("achievement_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Use non-existent achievement ID
		nonExistentID := testutil.RandomUUID()

		// Call the method being tested
		_, err := svc.GetAchievement(ctx, nonExistentID)

		// Verify the error
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Close the database to force errors
		client.Close()

		// Use any ID
		testID := testutil.RandomUUID()

		// Call the method being tested
		_, err := svc.GetAchievement(ctx, testID)

		// Verify the results
		require.Error(t, err)
	})
}
