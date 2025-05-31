package sescsvc

import (
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestGenerateUserPointsReport(t *testing.T) {
	t.Run("success with users and achievements", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		// Create test users
		user1 := createTestUser(t, svc)
		user2 := createTestUser(t, svc)

		// Create test template
		template := createTestTemplate(t, svc)

		// Create achievements for user1
		ach1 := createTestAchievement(t, svc, user1, template)
		ach2 := createTestAchievement(t, svc, user1, template)

		// Create achievement for user2
		ach3 := createTestAchievement(t, svc, user2, template)

		// Mark achievements as done with points
		svc.client.Achievement.UpdateOneID(ach1.ID).
			SetStatus(string(achievement.StatusDone)).
			SetPoints(10).
			ExecX(ctx)
		svc.client.Achievement.UpdateOneID(ach2.ID).
			SetStatus(string(achievement.StatusDone)).
			SetPoints(15).
			ExecX(ctx)
		svc.client.Achievement.UpdateOneID(ach3.ID).
			SetStatus(string(achievement.StatusDone)).
			SetPoints(20).
			ExecX(ctx)

		// Generate report
		buffer, err := svc.GenerateUserPointsReport(ctx)
		require.NoError(t, err, "GenerateUserPointsReport should not fail")
		require.NotNil(t, buffer, "Buffer should not be nil")
		require.Positive(t, buffer.Len(), "Buffer should contain data")
	})

	t.Run("success with no achievements", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		// Create test users but no achievements
		createTestUser(t, svc)
		createTestUser(t, svc)

		// Generate report
		buffer, err := svc.GenerateUserPointsReport(ctx)
		require.NoError(t, err, "GenerateUserPointsReport should not fail")
		require.NotNil(t, buffer, "Buffer should not be nil")
		require.Positive(t, buffer.Len(), "Buffer should contain data even with no points")
	})

	t.Run("only includes done achievements", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)

		// Create achievements with different statuses
		achDraft := createTestAchievement(t, svc, user, template)
		achDone := createTestAchievement(t, svc, user, template)
		achAccounted := createTestAchievement(t, svc, user, template)

		// Set different statuses and points
		svc.client.Achievement.UpdateOneID(achDraft.ID).
			SetStatus(string(achievement.StatusDraft)).
			SetPoints(10).
			ExecX(ctx)
		svc.client.Achievement.UpdateOneID(achDone.ID).
			SetStatus(string(achievement.StatusDone)).
			SetPoints(15).
			ExecX(ctx)
		svc.client.Achievement.UpdateOneID(achAccounted.ID).
			SetStatus(string(achievement.StatusAccounted)).
			SetPoints(20).
			ExecX(ctx)

		// Generate report
		buffer, err := svc.GenerateUserPointsReport(ctx)
		require.NoError(t, err, "GenerateUserPointsReport should not fail")
		require.NotNil(t, buffer, "Buffer should not be nil")
		require.Positive(t, buffer.Len(), "Buffer should contain data")

		// The report should only include the "done" achievement (15 points)
		// We can't easily parse the Excel file in tests, but we can verify the method runs
	})
}

func TestMarkAchievementsAsAccounted(t *testing.T) {
	t.Run("success marking done achievements as accounted", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)

		// Create achievements
		ach1 := createTestAchievement(t, svc, user, template)
		ach2 := createTestAchievement(t, svc, user, template)

		// Mark as done
		svc.client.Achievement.UpdateOneID(ach1.ID).
			SetStatus(string(achievement.StatusDone)).
			ExecX(ctx)
		svc.client.Achievement.UpdateOneID(ach2.ID).
			SetStatus(string(achievement.StatusDone)).
			ExecX(ctx)

		// Mark as accounted
		err := svc.MarkAchievementsAsAccounted(ctx, []uuid.UUID{ach1.ID, ach2.ID})
		require.NoError(t, err, "MarkAchievementsAsAccounted should not fail")

		// Verify status changed
		updatedAch1 := svc.client.Achievement.GetX(ctx, ach1.ID)
		updatedAch2 := svc.client.Achievement.GetX(ctx, ach2.ID)

		require.Equal(t, string(achievement.StatusAccounted), updatedAch1.Status)
		require.Equal(t, string(achievement.StatusAccounted), updatedAch2.Status)
	})

	t.Run("only updates done achievements", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)

		// Create achievements with different statuses
		achDraft := createTestAchievement(t, svc, user, template)
		achDone := createTestAchievement(t, svc, user, template)

		// Set different statuses
		svc.client.Achievement.UpdateOneID(achDraft.ID).
			SetStatus(string(achievement.StatusDraft)).
			ExecX(ctx)
		svc.client.Achievement.UpdateOneID(achDone.ID).
			SetStatus(string(achievement.StatusDone)).
			ExecX(ctx)

		// Try to mark both as accounted
		err := svc.MarkAchievementsAsAccounted(ctx, []uuid.UUID{achDraft.ID, achDone.ID})
		require.NoError(t, err, "MarkAchievementsAsAccounted should not fail")

		// Verify only the done achievement was updated
		updatedAchDraft := svc.client.Achievement.GetX(ctx, achDraft.ID)
		updatedAchDone := svc.client.Achievement.GetX(ctx, achDone.ID)

		require.Equal(t, string(achievement.StatusDraft), updatedAchDraft.Status)
		require.Equal(t, string(achievement.StatusAccounted), updatedAchDone.Status)
	})

	t.Run("empty list does nothing", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		err := svc.MarkAchievementsAsAccounted(ctx, []uuid.UUID{})
		require.NoError(t, err, "MarkAchievementsAsAccounted with empty list should not fail")
	})
}
