package tests

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateRandomString returns a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}

func TestAdminLogin(t *testing.T) {
	// Skip if test API URL is not set
	SkipIfNoTestAPI(t)

	// Create a client for the test API
	client := NewTestClient()

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
	// Skip if test API URL is not set
	SkipIfNoTestAPI(t)

	// Create a client for the test API
	client := NewTestClient()
	ctx := t.Context()

	// Login as admin
	adminToken, err := client.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, adminToken)
	client.SetToken(adminToken)

	// Create a test user with unique data
	randomSuffix := generateRandomString(8)
	userData := CreateValidUserData(
		fmt.Sprintf("Test_%s", randomSuffix),
		fmt.Sprintf("User_%s", randomSuffix),
		2,
	)

	user, err := client.CreateUser(ctx, userData)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Register credentials for the user with unique username
	credentialsData := RegisterUserRequest{
		Username: fmt.Sprintf("testuser_%s", uuid.Must(uuid.NewV7()).String()),
		Password: "password123",
	}

	err = client.RegisterUser(ctx, user.ID.String(), credentialsData)
	require.NoError(t, err)

	// Now try to login as the new user
	userClient := NewTestClient()
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
