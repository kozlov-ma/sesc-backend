package atsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestAchievementGroupByID(t *testing.T) {
	t.Run("existing_group", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// First create a group that we can then retrieve
		createOptions := achievement.GroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}

		createdGroup, err := svc.CreateAchievementGroup(ctx, createOptions)
		require.NoError(t, err)

		// Call the method being tested
		group, err := svc.AchievementGroupByID(ctx, createdGroup.ID)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, createdGroup.ID, group.ID)
		require.Equal(t, createdGroup.Name, group.Name)
		require.Equal(t, createdGroup.Description, group.Description)
		require.Equal(t, createdGroup.Active, group.Active)
	})

	t.Run("non_existent_group", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID that doesn't exist
		nonExistentID := testutil.RandomUUID()

		// Call the method being tested
		group, err := svc.AchievementGroupByID(ctx, nonExistentID)

		// Verify the results
		require.Equal(t, achievement.ErrAchievementGroupNotFound, err)
		require.Nil(t, group)
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
			Name:        "Test Group",
			Description: "Test Description",
		}

		createdGroup, err := svc.CreateAchievementGroup(ctx, createOptions)
		require.NoError(t, err)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		group, err := svc.AchievementGroupByID(ctx, createdGroup.ID)

		// Verify the results
		require.Error(t, err)
		require.Nil(t, group)
	})
}
