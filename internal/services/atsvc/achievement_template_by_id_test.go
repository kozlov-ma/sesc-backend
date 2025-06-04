package atsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestAchievementTemplateByID(t *testing.T) {
	t.Run("existing_template", func(t *testing.T) {
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

		// Create a template that we can then retrieve
		templateOptions := achievement.TemplateCreateOptions{
			Name:        "Test Template",
			Description: "Test Template Description",
			PointsLimit: 100,
			GroupID:     group.ID,
			Kind:        achievement.Kind("olympiad"),
		}

		createdTemplate, err := svc.CreateAchievementTemplate(ctx, templateOptions)
		require.NoError(t, err)

		// Call the method being tested
		template, err := svc.AchievementTemplateByID(ctx, createdTemplate.ID)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, createdTemplate.ID, template.ID)
		require.Equal(t, createdTemplate.Name, template.Name)
		require.Equal(t, createdTemplate.Description, template.Description)
		require.Equal(t, createdTemplate.PointsLimit, template.PointsLimit)
		require.Equal(t, createdTemplate.GroupID, template.GroupID)
		require.Equal(t, createdTemplate.Kind, template.Kind)
		require.Equal(t, createdTemplate.Active, template.Active)
	})

	t.Run("non_existent_template", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID that doesn't exist
		nonExistentID := testutil.RandomUUID()

		// Call the method being tested
		template, err := svc.AchievementTemplateByID(ctx, nonExistentID)

		// Verify the results
		require.Equal(t, achievement.ErrAchievementTemplateNotFound, err)
		require.Nil(t, template)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test data
		tc := testutil.SetupTestContext(t)

		// Close the client we're using to force errors, but not the one used to set up test context
		client.Close()

		// Call the method being tested with a valid template ID from test context
		template, err := svc.AchievementTemplateByID(ctx, tc.Template.ID)

		// Verify the results
		require.Error(t, err)
		require.Nil(t, template)
	})
}
