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

func TestCreateAchievement(t *testing.T) {
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

	t.Run("success", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Create achievement
		opt := achievement.CreateOptions{
			ForUserID:  tc.User.ID,
			TemplateID: tc.Template.ID,
			Points:     8, // User sets their own points
		}

		ach, err := svc.CreateAchievement(ctx, opt)
		require.NoError(t, err, "CreateAchievement failed")

		// Verify achievement was created correctly
		require.NotEqual(t, uuid.Nil, ach.ID, "Achievement ID should not be nil")
		require.Equal(t, tc.User.ID, ach.OwnerID, "Achievement owner should match")
		require.Equal(t, tc.Template.ID, ach.TemplateID, "Achievement template should match")
		require.Equal(t, achievement.StatusDraft, ach.Status, "Achievement should be in draft status")
		require.Equal(t, 8, ach.Points, "Achievement should have user-specified points")
	})

	t.Run("points exceed limit", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Create achievement with points exceeding template limit
		opt := achievement.CreateOptions{
			ForUserID:  tc.User.ID,
			TemplateID: tc.Template.ID,
			Points:     tc.Template.PointsLimit + 1, // Exceed limit
		}

		_, err := svc.CreateAchievement(ctx, opt)
		require.Error(t, err, "CreateAchievement should fail with points exceeding limit")
		require.Equal(t, achievement.ErrPointsLimitExceeded, err, "Should return points limit exceeded error")
	})

	t.Run("template not found", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Use non-existent template ID
		opt := achievement.CreateOptions{
			ForUserID:  tc.User.ID,
			TemplateID: uuid.Must(uuid.NewV7()), // random ID that doesn't exist
		}

		// Try to create achievement with non-existent template
		_, err := svc.CreateAchievement(ctx, opt)
		require.Error(t, err, "CreateAchievement should fail with non-existent template")
	})
}
