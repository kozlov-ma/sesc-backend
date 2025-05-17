package tests

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminLogin(t *testing.T) {
	// Start a test application
	app := testutil.StartTestApp(t)

	// Create a client
	client := NewClient(app.URL)

	// Test admin login with correct credentials
	ctx := t.Context()
	token, err := client.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token is valid
	err = client.ValidateToken(ctx)
	require.NoError(t, err)

	// Test admin login with incorrect credentials
	_, err = client.LoginAdmin(ctx, "wrong", "wrong")
	assert.Error(t, err)
}

func TestLoginFlow(t *testing.T) {
	// Start a test application
	app := testutil.StartTestApp(t)

	// Create a client and login as admin
	client := NewClient(app.URL)
	ctx := t.Context()

	// Login as admin
	adminToken, err := client.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, adminToken)
	client.SetToken(adminToken)

	// Create a test user
	userData := CreateUserRequest{
		FirstName:  "Test",
		LastName:   "User",
		RoleID:     2,
		PictureURL: "/test.jpg",
	}

	user, err := client.CreateUser(ctx, userData)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Register credentials for the user
	credentialsData := RegisterUserRequest{
		Username: "testuser",
		Password: "password123",
	}

	err = client.RegisterUser(ctx, user.ID.String(), credentialsData)
	require.NoError(t, err)

	// Now try to login as the new user
	userClient := NewClient(app.URL)
	userToken, err := userClient.Login(ctx, credentialsData.Username, credentialsData.Password)
	require.NoError(t, err)
	assert.NotEmpty(t, userToken)

	// Verify the token
	userClient.SetToken(userToken)
	err = userClient.ValidateToken(ctx)
	require.NoError(t, err)

	// Get current user and verify details
	currentUser, err := userClient.GetCurrentUser(ctx)
	require.NoError(t, err)
	assert.Equal(t, user.ID, currentUser.ID)
	assert.Equal(t, userData.FirstName, currentUser.FirstName)
	assert.Equal(t, userData.LastName, currentUser.LastName)
}
