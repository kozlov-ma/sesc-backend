package achsvc

import (
	"testing"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestSubmitAchievement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client, 24*time.Hour)

		// Create test user, template, and achievement
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, sesc.OlympiadDeputy)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)

		// Call the method being tested
		result, err := svc.SubmitAchievement(ctx, achievement.SubmitOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
		})

		// Verify the results
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, ach.ID, result.ID)
		require.Equal(t, string(achievement.StatusDepheadReview), result.Status)
	})

	t.Run("achievement_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client, 24*time.Hour)

		// Create test user
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		nonExistentID := testutil.RandomUUID()

		// Call the method being tested
		_, err := svc.SubmitAchievement(ctx, achievement.SubmitOptions{
			AchievementID: nonExistentID,
			OwnerID:       user.ID,
		})

		// Verify the error
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("wrong_status", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client, 24*time.Hour)

		// Create test user, template, and achievement with non-draft status
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, sesc.OlympiadDeputy)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDone)

		// Call the method being tested
		_, err := svc.SubmitAchievement(ctx, achievement.SubmitOptions{
			AchievementID: ach.ID,
			OwnerID:       user.ID,
		})

		// Verify the error
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})
}
