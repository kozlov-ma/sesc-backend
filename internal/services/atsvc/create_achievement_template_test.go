package atsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestCreateAchievementTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group for the template
		groupOptions := achievement.GroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}

		group, err := svc.CreateAchievementGroup(ctx, groupOptions)
		require.NoError(t, err)

		// Prepare test data for template creation
		options := achievement.TemplateCreateOptions{
			Name:         "Test Template",
			Description:  "Test Template Description",
			PointsLimit:  100,
			GroupID:      group.ID,
			ReviewerRole: achievement.ReviewerRole(sesc.OlympiadDeputy),
		}

		// Call the method being tested
		template, err := svc.CreateAchievementTemplate(ctx, options)

		// Verify the results
		require.NoError(t, err)
		require.NotEmpty(t, template.ID, "Template ID should be generated")
		require.Equal(t, options.Name, template.Name)
		require.Equal(t, options.Description, template.Description)
		require.Equal(t, options.PointsLimit, template.PointsLimit)
		require.Equal(t, options.GroupID, template.GroupID)
		require.Equal(t, options.ReviewerRole, template.ReviewerRole)
		require.True(t, template.Active, "Template should be active by default")

		// Verify the template was actually created in the database
		dbTemplate, err := client.AchievementTemplate.Get(ctx, template.ID)
		require.NoError(t, err)
		require.Equal(t, options.Name, dbTemplate.Name)
		require.Equal(t, options.Description, dbTemplate.Description)
		require.Equal(t, options.PointsLimit, dbTemplate.PointsLimit)
		require.Equal(t, options.GroupID, dbTemplate.GroupID)
		require.Equal(t, options.ReviewerRole, dbTemplate.ReviewerRole)
	})

	t.Run("invalid_reviewer_role", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group for the template
		groupOptions := achievement.GroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}

		group, err := svc.CreateAchievementGroup(ctx, groupOptions)
		require.NoError(t, err)

		// Prepare test data with invalid reviewer role
		options := achievement.TemplateCreateOptions{
			Name:         "Test Template",
			Description:  "Test Template Description",
			PointsLimit:  100,
			GroupID:      group.ID,
			ReviewerRole: achievement.ReviewerRole(999), // Invalid role
		}

		// Call the method being tested
		template, err := svc.CreateAchievementTemplate(ctx, options)

		// Verify the results
		require.Error(t, err, "Should return error for invalid reviewer role")
		require.Nil(t, template, "No template should be created")
	})

	t.Run("group_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID for non-existent group
		nonExistentGroupID := testutil.RandomUUID()

		// Prepare test data with non-existent group ID
		options := achievement.TemplateCreateOptions{
			Name:        "Test Template",
			Description: "Test Template Description",
			PointsLimit: 100,
			GroupID:     nonExistentGroupID,
			Kind:        achievement.Kind("olympiad"),
		}

		// Call the method being tested
		template, err := svc.CreateAchievementTemplate(ctx, options)

		// Verify the results
		require.Equal(t, achievement.ErrAchievementGroupNotFound, err)
		require.Nil(t, template, "No template should be created")
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group for the template
		groupOptions := achievement.GroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}

		group, err := svc.CreateAchievementGroup(ctx, groupOptions)
		require.NoError(t, err)

		// Close the database to force errors
		client.Close()

		// Prepare test data
		options := achievement.TemplateCreateOptions{
			Name:         "Test Template",
			Description:  "Test Template Description",
			PointsLimit:  100,
			GroupID:      group.ID,
			ReviewerRole: achievement.ReviewerRole(sesc.OlympiadDeputy),
		}

		// Call the method being tested
		template, err := svc.CreateAchievementTemplate(ctx, options)

		// Verify the results
		require.Error(t, err)
		require.Nil(t, template, "No template should be created")
	})
}
