package usersvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestUpdateProfilePicture(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		// New picture URL
		newPictureURL := "https://example.com/profile-pictures/test-user.jpg"

		// Call the method being tested
		err := svc.UpdateProfilePicture(ctx, user.ID, newPictureURL)

		// Verify the results
		require.NoError(t, err)

		// Verify the profile picture was actually updated
		updatedUser, err := svc.UserByID(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, newPictureURL, updatedUser.PictureURL)
	})

	t.Run("user_not_found", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID that doesn't exist
		nonExistentID := testutil.RandomUUID()

		// New picture URL
		newPictureURL := "https://example.com/profile-pictures/test-user.jpg"

		// Call the method being tested
		err := svc.UpdateProfilePicture(ctx, nonExistentID, newPictureURL)

		// Verify the results
		require.ErrorIs(t, err, sesc.ErrUserNotFound)
	})

	t.Run("empty_picture_url", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		// Call the method with empty URL to clear the picture
		err := svc.UpdateProfilePicture(ctx, user.ID, "")

		// Verify the results
		require.NoError(t, err)

		// Verify the profile picture was actually cleared
		updatedUser, err := svc.UserByID(ctx, user.ID)
		require.NoError(t, err)
		require.Empty(t, updatedUser.PictureURL)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		user := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		// Close the database to force errors
		client.Close()

		// New picture URL
		newPictureURL := "https://example.com/profile-pictures/test-user.jpg"

		// Call the method being tested
		err := svc.UpdateProfilePicture(ctx, user.ID, newPictureURL)

		// Verify the results
		require.Error(t, err)
	})
}
