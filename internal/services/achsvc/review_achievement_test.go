package achsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestReviewAchievement(t *testing.T) {
	t.Run("success_dephead_review", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, reviewer, template, and achievement
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		reviewer := testutil.CreateTestUser(ctx, t, client, "Test", "Reviewer", sesc.Dephead)
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDepheadReview)

		// Call the method being tested
		result, err := svc.ReviewAchievement(ctx, achievement.ReviewOptions{
			AchievementID:      ach.ID,
			AchievementOwnerID: user.ID,
			ReviewerID:         reviewer.ID,
			PointsAssigned:     50,
			Comment:            "Good work!",
		})

		// Verify the results
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, ach.ID, result.ID)
		require.Equal(t, 50, result.Points)
		require.Equal(t, string(achievement.StatusInspectorReview), result.Status)
	})

	t.Run("achievement_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user and reviewer
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		reviewer := testutil.CreateTestUser(ctx, t, client, "Test", "Reviewer", sesc.Dephead)
		nonExistentID := testutil.RandomUUID()

		// Call the method being tested
		_, err := svc.ReviewAchievement(ctx, achievement.ReviewOptions{
			AchievementID:      nonExistentID,
			AchievementOwnerID: user.ID,
			ReviewerID:         reviewer.ID,
			PointsAssigned:     50,
			Comment:            "Good work!",
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
		svc := New(client)

		// Create test user, reviewer, template, and achievement with wrong status
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		reviewer := testutil.CreateTestUser(ctx, t, client, "Test", "Reviewer", sesc.Dephead)
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)

		// Call the method being tested
		_, err := svc.ReviewAchievement(ctx, achievement.ReviewOptions{
			AchievementID:      ach.ID,
			AchievementOwnerID: user.ID,
			ReviewerID:         reviewer.ID,
			PointsAssigned:     50,
			Comment:            "Good work!",
		})

		// Verify the error
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})

	t.Run("invalid_reviewer", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test user, invalid reviewer, template, and achievement
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		invalidReviewer := testutil.CreateTestUser(ctx, t, client, "Test", "InvalidReviewer", sesc.Role(1)) // Not a dephead
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		ach := testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDepheadReview)

		// Call the method being tested
		_, err := svc.ReviewAchievement(ctx, achievement.ReviewOptions{
			AchievementID:      ach.ID,
			AchievementOwnerID: user.ID,
			ReviewerID:         invalidReviewer.ID,
			PointsAssigned:     50,
			Comment:            "Good work!",
		})

		// Verify the error
		require.ErrorIs(t, err, sesc.ErrInvalidRole)
	})
}
