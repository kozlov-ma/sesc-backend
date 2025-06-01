package usersvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestUserByID(t *testing.T) {
	t.Run("existing_user", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user in the database
		expectedUser := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		// Call the method being tested
		user, err := svc.UserByID(ctx, expectedUser.ID)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, expectedUser.ID, user.ID)
		require.Equal(t, expectedUser.FirstName, user.FirstName)
		require.Equal(t, expectedUser.LastName, user.LastName)
		require.Equal(t, expectedUser.Department.ID, user.Department.ID)
		require.Equal(t, expectedUser.Role, user.Role)
	})

	t.Run("non_existent_user", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID that doesn't exist
		nonExistentID := testutil.RandomUUID()

		// Call the method being tested
		user, err := svc.UserByID(ctx, nonExistentID)

		// Verify the results
		require.Equal(t, sesc.ErrUserNotFound, err)
		require.Empty(t, user.ID)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		testUser := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		user, err := svc.UserByID(ctx, testUser.ID)

		// Verify the results
		require.Error(t, err)
		require.Empty(t, user.ID)
	})
}
