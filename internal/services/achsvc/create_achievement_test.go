package achsvc_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/achsvc"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
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
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Kind("scientific"))

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
			ForUser:    tc.User,
			TemplateID: tc.Template.ID,
		}

		ach, err := svc.CreateAchievement(ctx, opt)
		require.NoError(t, err, "CreateAchievement failed")

		// Verify achievement was created correctly
		require.NotEqual(t, uuid.Nil, ach.ID, "Achievement ID should not be nil")
		require.Equal(t, tc.User.ID, ach.Owner.ID, "Achievement owner should match")
		require.Equal(t, tc.Template.ID, ach.Template.ID, "Achievement template should match")
		require.Equal(t, string(achievement.StatusDraft), string(ach.Status), "Achievement should be in draft status")
		require.Equal(t, tc.Template.PointsLimit, ach.Points, "Achievement should have template points limit")
	})

	t.Run("template not found", func(t *testing.T) {
		// Setup test
		ctx, tc, svc := setup(t)

		// Use non-existent template ID
		opt := achievement.CreateOptions{
			ForUser:    tc.User,
			TemplateID: uuid.Must(uuid.NewV7()), // random ID that doesn't exist
		}

		// Try to create achievement with non-existent template
		_, err := svc.CreateAchievement(ctx, opt)
		require.Error(t, err, "CreateAchievement should fail with non-existent template")
	})
}
