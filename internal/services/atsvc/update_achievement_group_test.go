package atsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestUpdateAchievementGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group to update
		createOptions := achievement.GroupCreateOptions{
			Name:        "Original Name",
			Description: "Original Description",
		}

		originalGroup, err := svc.CreateAchievementGroup(ctx, createOptions)
		require.NoError(t, err)

		// Prepare update data
		newName := "Updated Name"
		newDesc := "Updated Description"
		newActive := false

		updateOptions := achievement.GroupUpdateOptions{
			Name:        &newName,
			Description: &newDesc,
			Active:      &newActive,
		}

		// Call the method being tested
		updatedGroup, err := svc.UpdateAchievementGroup(ctx, originalGroup.ID, updateOptions)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, originalGroup.ID, updatedGroup.ID)
		require.Equal(t, newName, updatedGroup.Name)
		require.Equal(t, newDesc, updatedGroup.Description)
		require.Equal(t, newActive, updatedGroup.Active)

		// Verify the group was actually updated in the database
		dbGroup, err := client.AchievementGroup.Get(ctx, originalGroup.ID)
		require.NoError(t, err)
		require.Equal(t, newName, dbGroup.Name)
		require.Equal(t, newDesc, dbGroup.Description)
		require.Equal(t, newActive, dbGroup.Active)
	})

	t.Run("partial_update", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group to update
		createOptions := achievement.GroupCreateOptions{
			Name:        "Original Name",
			Description: "Original Description",
		}

		originalGroup, err := svc.CreateAchievementGroup(ctx, createOptions)
		require.NoError(t, err)

		// Prepare update data with only name updated
		newName := "Updated Name Only"

		updateOptions := achievement.GroupUpdateOptions{
			Name: &newName,
		}

		// Call the method being tested
		updatedGroup, err := svc.UpdateAchievementGroup(ctx, originalGroup.ID, updateOptions)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, originalGroup.ID, updatedGroup.ID)
		require.Equal(t, newName, updatedGroup.Name)
		require.Equal(t, originalGroup.Description, updatedGroup.Description)
		require.Equal(t, originalGroup.Active, updatedGroup.Active)
	})

	t.Run("group_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID that doesn't exist
		nonExistentID := testutil.RandomUUID()

		// Prepare update data
		newName := "Updated Name"
		updateOptions := achievement.GroupUpdateOptions{
			Name: &newName,
		}

		// Call the method being tested
		updatedGroup, err := svc.UpdateAchievementGroup(ctx, nonExistentID, updateOptions)

		// Verify the results
		require.Equal(t, achievement.ErrAchievementGroupNotFound, err)
		require.Empty(t, updatedGroup.ID)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group
		createOptions := achievement.GroupCreateOptions{
			Name:        "Original Name",
			Description: "Original Description",
		}

		originalGroup, err := svc.CreateAchievementGroup(ctx, createOptions)
		require.NoError(t, err)

		// Close the database to force errors
		client.Close()

		// Prepare update data
		newName := "Updated Name"
		updateOptions := achievement.GroupUpdateOptions{
			Name: &newName,
		}

		// Call the method being tested
		updatedGroup, err := svc.UpdateAchievementGroup(ctx, originalGroup.ID, updateOptions)

		// Verify the results
		require.Error(t, err)
		require.Empty(t, updatedGroup.ID)
	})
}
