package achsvc_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/achsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestUpdateAchievementPoints(t *testing.T) {
	// Setup function to reuse across test cases
	setup := func(t *testing.T) (context.Context, *testutil.TestContext, *achsvc.ACS) {
		// Create test context with event record
		ctx, _ := testutil.CreateTestContext(t)

		// Setup database and service
		client := testutil.SetupDatabase(t)
		t.Cleanup(func() { client.Close() })

		// Create test user and template
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", 1) // 1 is teacher role
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, sesc.ScientificDeputy)

		// Create achievement service
		svc := achsvc.New(client)

		return ctx, &testutil.TestContext{
			Client:   client,
			User:     user,
			Template: template,
		}, svc
	}

	t.Run("success - dephead points change", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Create achievement
		createOpt := achievement.CreateOptions{
			ForUserID:  tc.User.ID,
			TemplateID: tc.Template.ID,
			Points:     8,
		}
		ach, err := svc.CreateAchievement(ctx, createOpt)
		require.NoError(t, err)

		// Submit achievement
		submitOpt := achievement.SubmitOptions{
			OwnerID:       tc.User.ID,
			AchievementID: ach.ID,
		}
		_, err = svc.SubmitAchievement(ctx, submitOpt)
		require.NoError(t, err)

		// Create dephead user
		dephead := testutil.CreateTestUser(ctx, t, tc.Client, "Dephead", "User", 2) // 2 is dephead role

		// Review achievement with lower points (requesting points change)
		reviewOpt := achievement.ReviewOptions{
			AchievementOwnerID: tc.User.ID,
			AchievementID:      ach.ID,
			ReviewerID:         dephead.ID,
			PointsAssigned:     6, // Lower than original 8
			Comment:            "Please adjust points",
		}
		_, err = svc.ReviewAchievement(ctx, reviewOpt)
		require.NoError(t, err)

		// Update achievement points
		updateOpt := achievement.UpdatePointsOptions{
			OwnerID:       tc.User.ID,
			AchievementID: ach.ID,
			Points:        6,
		}
		updatedAch, err := svc.UpdateAchievementPoints(ctx, updateOpt)
		require.NoError(t, err, "UpdateAchievementPoints failed")

		// Verify achievement was updated correctly
		require.Equal(t, 6, updatedAch.Points, "Achievement points should be updated")
		require.Equal(t, achievement.StatusDraft, updatedAch.Status, "Achievement should be back to draft status")
	})

	t.Run("success - inspector points change", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Create achievement
		createOpt := achievement.CreateOptions{
			ForUserID:  tc.User.ID,
			TemplateID: tc.Template.ID,
			Points:     8,
		}
		ach, err := svc.CreateAchievement(ctx, createOpt)
		require.NoError(t, err)

		// Submit achievement
		submitOpt := achievement.SubmitOptions{
			OwnerID:       tc.User.ID,
			AchievementID: ach.ID,
		}
		_, err = svc.SubmitAchievement(ctx, submitOpt)
		require.NoError(t, err)

		// Create dephead user
		dephead := testutil.CreateTestUser(ctx, t, tc.Client, "Dephead", "User", 2) // 2 is dephead role

		// Review by dephead (approve)
		reviewOpt := achievement.ReviewOptions{
			AchievementOwnerID: tc.User.ID,
			AchievementID:      ach.ID,
			ReviewerID:         dephead.ID,
			PointsAssigned:     8,
			Comment:            "Approved by dephead",
		}
		_, err = svc.ReviewAchievement(ctx, reviewOpt)
		require.NoError(t, err)

		// Create inspector user
		inspector := testutil.CreateTestUser(ctx, t, tc.Client, "Inspector", "User", 3) // 3 is ScientificDeputy

		// Review by inspector with lower points (requesting points change)
		reviewOpt = achievement.ReviewOptions{
			AchievementOwnerID: tc.User.ID,
			AchievementID:      ach.ID,
			ReviewerID:         inspector.ID,
			PointsAssigned:     6, // Lower than original 8
			Comment:            "Please adjust points",
		}
		_, err = svc.ReviewAchievement(ctx, reviewOpt)
		require.NoError(t, err)

		// Update achievement points
		updateOpt := achievement.UpdatePointsOptions{
			OwnerID:       tc.User.ID,
			AchievementID: ach.ID,
			Points:        6,
		}
		updatedAch, err := svc.UpdateAchievementPoints(ctx, updateOpt)
		require.NoError(t, err, "UpdateAchievementPoints failed")

		// Verify achievement was updated correctly
		require.Equal(t, 6, updatedAch.Points, "Achievement points should be updated")
		require.Equal(t, achievement.StatusDraft, updatedAch.Status, "Achievement should be back to draft status")
	})

	t.Run("wrong status", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Create achievement
		createOpt := achievement.CreateOptions{
			ForUserID:  tc.User.ID,
			TemplateID: tc.Template.ID,
			Points:     8,
		}
		ach, err := svc.CreateAchievement(ctx, createOpt)
		require.NoError(t, err)

		// Try to update points when achievement is in draft status
		updateOpt := achievement.UpdatePointsOptions{
			OwnerID:       tc.User.ID,
			AchievementID: ach.ID,
			Points:        6,
		}
		_, err = svc.UpdateAchievementPoints(ctx, updateOpt)
		require.Error(t, err, "UpdateAchievementPoints should fail with wrong status")
		require.Equal(t, achievement.ErrWrongAchievementStatus, err, "Should return wrong achievement status error")
	})

	t.Run("points exceed limit", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Create achievement
		createOpt := achievement.CreateOptions{
			ForUserID:  tc.User.ID,
			TemplateID: tc.Template.ID,
			Points:     8,
		}
		ach, err := svc.CreateAchievement(ctx, createOpt)
		require.NoError(t, err)

		// Submit achievement
		submitOpt := achievement.SubmitOptions{
			OwnerID:       tc.User.ID,
			AchievementID: ach.ID,
		}
		_, err = svc.SubmitAchievement(ctx, submitOpt)
		require.NoError(t, err)

		// Create dephead user
		dephead := testutil.CreateTestUser(ctx, t, tc.Client, "Dephead", "User", 2) // 2 is dephead role

		// Review achievement with lower points (requesting points change)
		reviewOpt := achievement.ReviewOptions{
			AchievementOwnerID: tc.User.ID,
			AchievementID:      ach.ID,
			ReviewerID:         dephead.ID,
			PointsAssigned:     6,
			Comment:            "Please adjust points",
		}
		_, err = svc.ReviewAchievement(ctx, reviewOpt)
		require.NoError(t, err)

		// Try to update achievement points exceeding template limit
		updateOpt := achievement.UpdatePointsOptions{
			OwnerID:       tc.User.ID,
			AchievementID: ach.ID,
			Points:        tc.Template.PointsLimit + 1, // Exceed limit
		}
		_, err = svc.UpdateAchievementPoints(ctx, updateOpt)
		require.Error(t, err, "UpdateAchievementPoints should fail with points exceeding limit")
		require.Equal(t, achievement.ErrPointsLimitExceeded, err, "Should return points limit exceeded error")
	})

	t.Run("achievement not found", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Try to update non-existent achievement
		updateOpt := achievement.UpdatePointsOptions{
			OwnerID:       tc.User.ID,
			AchievementID: uuid.Must(uuid.NewV7()),
			Points:        6,
		}
		_, err := svc.UpdateAchievementPoints(ctx, updateOpt)
		require.Error(t, err, "UpdateAchievementPoints should fail with achievement not found")
		require.Equal(t, achievement.ErrAchievementNotFound, err, "Should return achievement not found error")
	})
}
