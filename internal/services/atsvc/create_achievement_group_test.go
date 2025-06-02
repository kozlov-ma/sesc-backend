package atsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestCreateAchievementGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Prepare test data
		options := achievement.GroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}

		// Call the method being tested
		group, err := svc.CreateAchievementGroup(ctx, options)

		// Verify the results
		require.NoError(t, err)
		require.NotEmpty(t, group.ID, "Group ID should be generated")
		require.Equal(t, options.Name, group.Name)
		require.Equal(t, options.Description, group.Description)
		require.True(t, group.Active, "Group should be active by default")

		// Verify the group was actually created in the database
		dbGroup, err := client.AchievementGroup.Get(ctx, group.ID)
		require.NoError(t, err)
		require.Equal(t, options.Name, dbGroup.Name)
		require.Equal(t, options.Description, dbGroup.Description)
		require.True(t, dbGroup.Active)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Close the database to force errors
		client.Close()

		// Prepare test data
		options := achievement.GroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}

		// Call the method being tested
		group, err := svc.CreateAchievementGroup(ctx, options)

		// Verify the results
		require.Error(t, err)
		require.Empty(t, group.ID, "No group should be created")
	})
}
