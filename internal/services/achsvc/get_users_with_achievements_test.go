package achsvc

import (
	"fmt"
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestGetUsersWithAchievements(t *testing.T) {
	t.Run("success_basic", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test users and department
		dept := testutil.CreateTestDepartment(ctx, t, client, "Test Department")
		user1 := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test1", "User1", sesc.Student, dept)
		user2 := testutil.CreateTestUserWithDepartment(ctx, t, client, "Test2", "User2", sesc.Student, dept)
		asker := testutil.CreateTestUserWithDepartment(ctx, t, client, "Asking", "User", sesc.Student, dept)

		// Create template and achievements
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindOlympiad)
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
			require.NotEmpty(t, user.Edges.Achievements)
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
		dept1 := testutil.CreateTestDepartment(ctx, t, client, "Department 1")
		dept2 := testutil.CreateTestDepartment(ctx, t, client, "Department 2")

		user1 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User1", "Dept1", sesc.Student, dept1)
		user2 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User2", "Dept2", sesc.Student, dept2)
		user3 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User3", "Dept1", sesc.Student, dept1)
		dephead := testutil.CreateTestUserWithDepartment(ctx, t, client, "Dep", "Head", sesc.Dephead, dept1)

		// Create template and achievements
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindOlympiad)

		// Create achievements with different statuses and departments
		testutil.CreateTestAchievement(ctx, t, client, user1, template, achievement.StatusDepheadReview) // Should be visible
		testutil.CreateTestAchievement(ctx, t, client, user1, template, achievement.StatusDraft)         // Should not be visible
		testutil.CreateTestAchievement(ctx, t, client, user2, template, achievement.StatusDepheadReview) // Should not be visible (different dept)
		testutil.CreateTestAchievement(ctx, t, client, user3, template, achievement.StatusDepheadReview) // Should be visible

		// Call the method with department head asking
		users, total, err := svc.GetUsersWithAchievements(ctx, dephead.ID, 0, 10)

		// Verify the results - should only see users from their department with DepheadReview achievements
		require.NoError(t, err)
		require.Len(t, users, 2) // user1 and user3
		require.Equal(t, 2, total)

		// Verify that all returned users are from dept1 and have DepheadReview achievements
		for _, user := range users {
			require.Equal(t, dept1.ID, user.DepartmentID)
			require.NotEmpty(t, user.Edges.Achievements)
			for _, ach := range user.Edges.Achievements {
				require.Equal(t, string(achievement.StatusDepheadReview), ach.Status)
			}
		}
	})

	t.Run("olympiad_deputy_filtering", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create users and templates
		dept := testutil.CreateTestDepartment(ctx, t, client, "Test Department")
		user1 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User1", "Test", sesc.Student, dept)
		user2 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User2", "Test", sesc.Student, dept)
		user3 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User3", "Test", sesc.Student, dept)
		olympiadDeputy := testutil.CreateTestUserWithDepartment(ctx, t, client, "Olympiad", "Deputy", sesc.OlympiadDeputy, dept)

		// Create templates for different kinds
		olympiadTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindOlympiad)
		devTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindDevelopment)

		// Create achievements with different kinds and statuses
		testutil.CreateTestAchievement(ctx, t, client, user1, olympiadTemplate, achievement.StatusInspectorReview) // Should be visible
		testutil.CreateTestAchievement(ctx, t, client, user2, devTemplate, achievement.StatusInspectorReview)      // Should not be visible (wrong kind)
		testutil.CreateTestAchievement(ctx, t, client, user3, olympiadTemplate, achievement.StatusDepheadReview)   // Should not be visible (wrong status)

		// Call the method with olympiad deputy asking
		users, total, err := svc.GetUsersWithAchievements(ctx, olympiadDeputy.ID, 0, 10)

		// Verify the results - should only see users with InspectorReview achievements of Olympiad kind
		require.NoError(t, err)
		require.Len(t, users, 1) // Only user1
		require.Equal(t, 1, total)
		require.Equal(t, user1.ID, users[0].ID)

		// Verify achievement properties
		require.NotEmpty(t, users[0].Edges.Achievements)
		for _, ach := range users[0].Edges.Achievements {
			require.Equal(t, string(achievement.StatusInspectorReview), ach.Status)
		}
	})

	t.Run("academic_director_filtering", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create users and templates
		dept := testutil.CreateTestDepartment(ctx, t, client, "Test Department")
		user1 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User1", "Test", sesc.Student, dept)
		user2 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User2", "Test", sesc.Student, dept)
		user3 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User3", "Test", sesc.Student, dept)
		academicDirector := testutil.CreateTestUserWithDepartment(ctx, t, client, "Academic", "Director", sesc.AcademicDirector, dept)

		// Create templates for different kinds
		olympiadTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindOlympiad)
		devTemplate := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindDevelopment)

		// Create achievements with different kinds and statuses
		testutil.CreateTestAchievement(ctx, t, client, user1, olympiadTemplate, achievement.StatusInspectorReview) // Should not be visible (wrong kind)
		testutil.CreateTestAchievement(ctx, t, client, user2, devTemplate, achievement.StatusInspectorReview)      // Should be visible
		testutil.CreateTestAchievement(ctx, t, client, user3, devTemplate, achievement.StatusDepheadReview)        // Should not be visible (wrong status)

		// Call the method with academic director asking
		users, total, err := svc.GetUsersWithAchievements(ctx, academicDirector.ID, 0, 10)

		// Verify the results - should only see users with InspectorReview achievements of Development kind
		require.NoError(t, err)
		require.Len(t, users, 1) // Only user2
		require.Equal(t, 1, total)
		require.Equal(t, user2.ID, users[0].ID)

		// Verify achievement properties
		require.NotEmpty(t, users[0].Edges.Achievements)
		for _, ach := range users[0].Edges.Achievements {
			require.Equal(t, string(achievement.StatusInspectorReview), ach.Status)
		}
	})

	t.Run("cross_department_filtering", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create departments and users
		dept1 := testutil.CreateTestDepartment(ctx, t, client, "Department 1")
		dept2 := testutil.CreateTestDepartment(ctx, t, client, "Department 2")

		user1 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User1", "Dept1", sesc.Student, dept1)
		user2 := testutil.CreateTestUserWithDepartment(ctx, t, client, "User2", "Dept2", sesc.Student, dept2)
		dephead1 := testutil.CreateTestUserWithDepartment(ctx, t, client, "Dep", "Head1", sesc.Dephead, dept1)

		// Create template and achievements
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindOlympiad)

		// Create achievements in both departments
		testutil.CreateTestAchievement(ctx, t, client, user1, template, achievement.StatusDepheadReview) // Should be visible
		testutil.CreateTestAchievement(ctx, t, client, user2, template, achievement.StatusDepheadReview) // Should not be visible

		// Call the method with department head from dept1
		users, total, err := svc.GetUsersWithAchievements(ctx, dephead1.ID, 0, 10)

		// Verify the results - should only see users from dept1
		require.NoError(t, err)
		require.Len(t, users, 1)
		require.Equal(t, 1, total)
		require.Equal(t, user1.ID, users[0].ID)
		require.Equal(t, dept1.ID, users[0].DepartmentID)
	})

	t.Run("no_users_with_achievements", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create an asker but no users with achievements
		dept := testutil.CreateTestDepartment(ctx, t, client, "Test Department")
		asker := testutil.CreateTestUserWithDepartment(ctx, t, client, "Asking", "User", sesc.Dephead, dept)

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
		dept := testutil.CreateTestDepartment(ctx, t, client, "Test Department")
		asker := testutil.CreateTestUserWithDepartment(ctx, t, client, "Asking", "User", sesc.Student, dept)
		template := testutil.CreateTestAchievementTemplate(ctx, t, client, achievement.KindOlympiad)

		// Create users with achievements in a predictable order
		for i := 0; i < 5; i++ {
			user := testutil.CreateTestUserWithDepartment(ctx, t, client, "User", fmt.Sprintf("Test%d", i), sesc.Student, dept)
			testutil.CreateTestAchievement(ctx, t, client, user, template, achievement.StatusDraft)
		}

		// Call the method being tested with pagination
		users, total, err := svc.GetUsersWithAchievements(ctx, asker.ID, 0, 2)

		// Verify the results
		require.NoError(t, err)
		require.LessOrEqual(t, len(users), 2)
		require.GreaterOrEqual(t, total, 5)

		// Verify ordering is consistent (by last name, first name)
		if len(users) > 1 {
			require.LessOrEqual(t, users[0].LastName, users[1].LastName)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create an asker
		dept := testutil.CreateTestDepartment(ctx, t, client, "Test Department")
		asker := testutil.CreateTestUserWithDepartment(ctx, t, client, "Asking", "User", sesc.Dephead, dept)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		_, _, err := svc.GetUsersWithAchievements(ctx, asker.ID, 0, 10)

		// Verify the results
		require.Error(t, err)
	})
}
