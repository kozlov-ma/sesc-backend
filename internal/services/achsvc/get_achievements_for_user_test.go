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
	t.Run("success_basic", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test users and department
		dept := testutil.CreateTestDepartmentWithName(ctx, t, client, "Test Department")
		owner := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test", "Owner", sesc.Teacher, dept)

		// Create template and achievements
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		ach1 := testutil.CreateTestAchievement(ctx, t, client, owner, template, achievement.StatusDraft)
		ach2 := testutil.CreateTestAchievement(ctx, t, client, owner, template, achievement.StatusDone)

		// Call the method being tested
		achievements, total, err := svc.GetUserAchievements(ctx, owner.ID, owner.ID, 0, 10)

		// Verify the results
		require.NoError(t, err)
		require.Len(t, achievements, 2)
		require.Equal(t, 2, total)

		// Verify that the returned achievements have the correct structure
		achievementIDs := []string{ach1.ID.String(), ach2.ID.String()}
		for _, ach := range achievements {
			require.NotEmpty(t, ach.ID)
			require.Contains(t, achievementIDs, ach.ID.String())
			require.Equal(t, owner.ID, ach.OwnerID)
			require.Equal(t, template.ID, ach.TemplateID)
		}
	})

	t.Run("department_head_filtering", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create departments and users
		dept1 := testutil.CreateTestDepartmentWithName(ctx, t, client, "Department 1")
		_ = testutil.CreateTestDepartmentWithName(ctx, t, client, "Department 2")

		owner := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test", "Owner", sesc.Teacher, dept1)
		dephead := testutil.CreateTestUserWithDepartment(ctx, t, client, "Dep", "Head", sesc.Dephead, dept1)

		// Create template and achievements
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)

		// Create achievements with different statuses
		testutil.CreateTestAchievement(ctx, t, client, owner, template, achievement.StatusDraft)
		_ = testutil.CreateTestAchievement(
			ctx,
			t,
			client,
			owner,
			template,
			achievement.StatusDepheadReview,
		)
		testutil.CreateTestAchievement(ctx, t, client, owner, template, achievement.StatusDone)

		// Call the method with department head asking
		achievements, total, err := svc.GetUserAchievements(ctx, owner.ID, dephead.ID, 0, 1)

		// Verify the results - should only see DepheadReview achievements from their department
		require.NoError(t, err)
		require.Len(t, achievements, 1)
		require.Equal(t, 2, total)
	})

	t.Run("olympiad_deputy_filtering", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create users and templates
		dept := testutil.CreateTestDepartmentWithName(ctx, t, client, "Test Department")
		owner := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test", "Owner", sesc.Teacher, dept)
		olympiadDeputy := testutil.CreateTestUserWithDepartment(
			ctx,
			t,
			client,
			"Olympiad",
			"Deputy",
			sesc.OlympiadDeputy,
			dept,
		)

		// Create templates for different kinds
		olympiadTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		devTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Development)

		// Create achievements with different kinds and statuses
		testutil.CreateTestAchievement(ctx, t, client, owner, olympiadTemplate, achievement.StatusDepheadReview)
		achOlympiadInspector := testutil.CreateTestAchievement(
			ctx,
			t,
			client,
			owner,
			olympiadTemplate,
			achievement.StatusInspectorReview,
		)
		testutil.CreateTestAchievement(ctx, t, client, owner, devTemplate, achievement.StatusInspectorReview)

		// Call the method with olympiad deputy asking
		achievements, total, err := svc.GetUserAchievements(ctx, owner.ID, olympiadDeputy.ID, 0, 10)

		// Verify the results - should only see InspectorReview achievements with Olympiad kind
		require.NoError(t, err)
		require.Len(t, achievements, 1)
		require.Equal(t, 1, total)
		require.Equal(t, achOlympiadInspector.ID, achievements[0].ID)
		require.Equal(t, string(achievement.StatusInspectorReview), achievements[0].Status)
	})

	t.Run("academic_director_filtering", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create users and templates
		dept := testutil.CreateTestDepartmentWithName(ctx, t, client, "Test Department")
		owner := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test", "Owner", sesc.Teacher, dept)
		academicDirector := testutil.CreateTestUserWithDepartment(
			ctx,
			t,
			client,
			"Academic",
			"Director",
			sesc.AcademicDirector,
			dept,
		)

		// Create templates for different kinds
		olympiadTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)
		devTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Development)

		// Create achievements with different kinds and statuses
		testutil.CreateTestAchievement(ctx, t, client, owner, olympiadTemplate, achievement.StatusInspectorReview)
		achDevInspector := testutil.CreateTestAchievement(
			ctx,
			t,
			client,
			owner,
			devTemplate,
			achievement.StatusInspectorReview,
		)
		testutil.CreateTestAchievement(ctx, t, client, owner, devTemplate, achievement.StatusDepheadReview)

		// Call the method with academic director asking
		achievements, total, err := svc.GetUserAchievements(ctx, owner.ID, academicDirector.ID, 0, 10)

		// Verify the results - should only see InspectorReview achievements with Development kind
		require.NoError(t, err)
		require.Len(t, achievements, 1)
		require.Equal(t, 1, total)
		require.Equal(t, achDevInspector.ID, achievements[0].ID)
		require.Equal(t, string(achievement.StatusInspectorReview), achievements[0].Status)
	})

	t.Run("no_achievements", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test users but no achievements
		dept := testutil.CreateTestDepartmentWithName(ctx, t, client, "Test Department")
		owner := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test", "Owner", sesc.Teacher, dept)
		asker := testutil.CreateTestUserWithDepartment(ctx, t, client, "Asking", "User", sesc.Teacher, dept)

		// Call the method being tested
		achievements, total, err := svc.GetUserAchievements(ctx, owner.ID, asker.ID, 0, 10)

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

		// Create test users and template
		dept := testutil.CreateTestDepartmentWithName(ctx, t, client, "Test Department")
		owner := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test", "Owner", sesc.Teacher, dept)
		asker := testutil.CreateTestUserWithDepartment(ctx, t, client, "Asking", "User", sesc.Dephead, dept)
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.Olympiad)

		// Create multiple achievements
		for range 5 {
			ach := testutil.CreateTestAchievement(ctx, t, client, owner, template, achievement.StatusDraft)
			_, err := svc.SubmitAchievement(ctx, achievement.SubmitOptions{
				OwnerID:       ach.OwnerID,
				AchievementID: ach.ID,
			})
			require.NoError(t, err)
		}

		// Call the method being tested with pagination
		achievements, total, err := svc.GetUserAchievements(ctx, owner.ID, asker.ID, 0, 2)

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

		// Create test users
		dept := testutil.CreateTestDepartmentWithName(ctx, t, client, "Test Department")
		owner := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test", "Owner", sesc.Teacher, dept)
		asker := testutil.CreateTestUserWithDepartment(ctx, t, client, "Asking", "User", sesc.Teacher, dept)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		_, _, err := svc.GetUserAchievements(ctx, owner.ID, asker.ID, 0, 10)

		// Verify the results
		require.Error(t, err)
	})
}
