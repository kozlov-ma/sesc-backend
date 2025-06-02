package usersvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestUsers(t *testing.T) {
	t.Run("get_all_users", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test users
		user1 := testutil.CreateTestUser(ctx, t, client, "First", "User", sesc.Role(1))
		user2 := testutil.CreateTestUser(ctx, t, client, "Second", "User", sesc.Role(2))
		user3 := testutil.CreateTestUser(ctx, t, client, "Third", "User", sesc.Role(1))

		// Call the method being tested
		users, err := svc.Users(ctx)

		// Verify the results
		require.NoError(t, err)
		require.Len(t, users, 3)

		// Create a map of user IDs for easier verification
		userMap := make(map[string]User)
		for _, u := range users {
			userMap[u.ID.String()] = u
		}

		// Verify each user is in the results
		u1, exists := userMap[user1.ID.String()]
		require.True(t, exists)
		require.Equal(t, user1.FirstName, u1.FirstName)
		require.Equal(t, user1.LastName, u1.LastName)

		u2, exists := userMap[user2.ID.String()]
		require.True(t, exists)
		require.Equal(t, user2.FirstName, u2.FirstName)
		require.Equal(t, user2.LastName, u2.LastName)

		u3, exists := userMap[user3.ID.String()]
		require.True(t, exists)
		require.Equal(t, user3.FirstName, u3.FirstName)
		require.Equal(t, user3.LastName, u3.LastName)
	})

	t.Run("empty_user_list", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Call the method being tested without creating any users
		users, err := svc.Users(ctx)

		// Verify the results
		require.NoError(t, err)
		require.Empty(t, users)
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

		// Call the method being tested
		users, err := svc.Users(ctx)

		// Verify the results
		require.Error(t, err)
		require.Empty(t, users)
	})
}
