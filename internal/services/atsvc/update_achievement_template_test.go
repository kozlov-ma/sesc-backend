package atsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestUpdateAchievementTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group for the template
		group := testutil.CreateTestAchievementGroup(ctx, t, client)

		// Create a template to update
		templateOptions := achievement.TemplateCreateOptions{
			Name:        "Original Template",
			Description: "Original Description",
			PointsLimit: 50,
			GroupID:     group.ID,
			Kind:        achievement.Kind("olympiad"),
		}

		originalTemplate, err := svc.CreateAchievementTemplate(ctx, templateOptions)
		require.NoError(t, err)

		// Prepare update data
		newName := "Updated Template"
		newDesc := "Updated Description"
		newPointsLimit := 100
		newActive := false
		newKind := achievement.Scientific

		updateOptions := achievement.TemplateUpdateOptions{
			Name:        &newName,
			Description: &newDesc,
			PointsLimit: &newPointsLimit,
			Active:      &newActive,
			Kind:        &newKind,
		}

		// Call the method being tested
		updatedTemplate, err := svc.UpdateAchievementTemplate(ctx, originalTemplate.ID, updateOptions)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, originalTemplate.ID, updatedTemplate.ID)
		require.Equal(t, newName, updatedTemplate.Name)
		require.Equal(t, newDesc, updatedTemplate.Description)
		require.Equal(t, newPointsLimit, updatedTemplate.PointsLimit)
		require.Equal(t, originalTemplate.GroupID, updatedTemplate.GroupID, "GroupID should not change")
		require.Equal(t, newActive, updatedTemplate.Active)
		require.Equal(t, newKind, updatedTemplate.Kind)

		// Verify the template was actually updated in the database
		dbTemplate, err := client.AchievementTemplate.Get(ctx, originalTemplate.ID)
		require.NoError(t, err)
		require.Equal(t, newName, dbTemplate.Name)
		require.Equal(t, newDesc, dbTemplate.Description)
		require.Equal(t, newPointsLimit, dbTemplate.PointsLimit)
		require.Equal(t, originalTemplate.GroupID, dbTemplate.GroupID)
		require.Equal(t, newActive, dbTemplate.Active)
		require.Equal(t, newKind, dbTemplate.Kind)
	})

	t.Run("partial_update", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group for the template
		group := testutil.CreateTestAchievementGroup(ctx, t, client)

		// Create a template to update
		templateOptions := achievement.TemplateCreateOptions{
			Name:        "Original Template",
			Description: "Original Description",
			PointsLimit: 50,
			GroupID:     group.ID,
			Kind:        achievement.Kind("olympiad"),
		}

		originalTemplate, err := svc.CreateAchievementTemplate(ctx, templateOptions)
		require.NoError(t, err)

		// Prepare update data with only name updated
		newName := "Updated Name Only"

		updateOptions := achievement.TemplateUpdateOptions{
			Name: &newName,
		}

		// Call the method being tested
		updatedTemplate, err := svc.UpdateAchievementTemplate(ctx, originalTemplate.ID, updateOptions)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, originalTemplate.ID, updatedTemplate.ID)
		require.Equal(t, newName, updatedTemplate.Name)
		require.Equal(t, originalTemplate.Description, updatedTemplate.Description)
		require.Equal(t, originalTemplate.PointsLimit, updatedTemplate.PointsLimit)
		require.Equal(t, originalTemplate.GroupID, updatedTemplate.GroupID)
		require.Equal(t, originalTemplate.Active, updatedTemplate.Active)
		require.Equal(t, originalTemplate.Kind, updatedTemplate.Kind)
	})

	t.Run("invalid_kind", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group for the template
		group := testutil.CreateTestAchievementGroup(ctx, t, client)

		// Create a template to update
		templateOptions := achievement.TemplateCreateOptions{
			Name:        "Original Template",
			Description: "Original Description",
			PointsLimit: 50,
			GroupID:     group.ID,
			Kind:        achievement.Kind("olympiad"),
		}

		originalTemplate, err := svc.CreateAchievementTemplate(ctx, templateOptions)
		require.NoError(t, err)

		// Prepare update data with invalid kind
		invalidKind := achievement.Kind("invalid_kind")

		updateOptions := achievement.TemplateUpdateOptions{
			Kind: &invalidKind,
		}

		// Call the method being tested
		updatedTemplate, err := svc.UpdateAchievementTemplate(ctx, originalTemplate.ID, updateOptions)

		// Verify the results
		require.Error(t, err, "Should return error for invalid kind")
		require.Nil(t, updatedTemplate)
	})

	t.Run("template_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID for non-existent template
		nonExistentID := testutil.RandomUUID()

		// Prepare update data
		newName := "Updated Name"
		updateOptions := achievement.TemplateUpdateOptions{
			Name: &newName,
		}

		// Call the method being tested
		updatedTemplate, err := svc.UpdateAchievementTemplate(ctx, nonExistentID, updateOptions)

		// Verify the results
		require.Equal(t, achievement.ErrAchievementTemplateNotFound, err)
		require.Nil(t, updatedTemplate)
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

		// Prepare update data
		newName := "Updated Name"
		updateOptions := achievement.TemplateUpdateOptions{
			Name: &newName,
		}

		// Call the method being tested with a valid template ID from test context
		updatedTemplate, err := svc.UpdateAchievementTemplate(ctx, tc.Template.ID, updateOptions)

		// Verify the results
		require.Error(t, err)
		require.Nil(t, updatedTemplate)
	})
}
