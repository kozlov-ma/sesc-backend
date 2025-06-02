package tests

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCRUD(t *testing.T) {
	// Skip if test API URL is not set
	SkipIfNoTestAPI(t)

	// Create a client for the test API
	client := NewTestClient()
	ctx := t.Context()

	adminToken, err := client.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	client.SetToken(adminToken)

	// 1. Get initial users count
	initialUsers, err := client.GetUsers(ctx)
	require.NoError(t, err)
	initialCount := len(initialUsers)

	// 2. Create a test department to assign to the user with a unique name
	uniqueID := uuid.Must(uuid.NewV7()).String()
	deptReq := CreateDepartmentRequest{
		Name:        fmt.Sprintf("Test Department %s", uniqueID),
		Description: fmt.Sprintf("Department for test users %s", uniqueID),
	}

	dept, err := client.CreateDepartment(ctx, deptReq)
	require.NoError(t, err)
	require.NotNil(t, dept)

	// 3. Create a new user with unique data
	randomSuffix := uuid.Must(uuid.NewV7()).String()
	userData := CreateValidUserData(
		fmt.Sprintf("John_%s", randomSuffix),
		fmt.Sprintf("Doe_%s", randomSuffix),
		2,
	)
	userData.MiddleName = fmt.Sprintf("Smith_%s", randomSuffix)
	userData.PictureURL = fmt.Sprintf("/images/users/john_%s.jpg", randomSuffix)
	userData.DepartmentID = &dept.ID

	user, err := client.CreateUser(ctx, userData)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Verify user was created with correct data
	assert.Equal(t, userData.FirstName, user.FirstName)
	assert.Equal(t, userData.LastName, user.LastName)
	assert.Equal(t, userData.MiddleName, user.MiddleName)
	assert.Equal(t, userData.PictureURL, user.PictureURL)
	assert.NotEqual(t, uuid.Nil, user.ID)

	// Verify department ID only if it's set
	if user.DepartmentID != nil {
		assert.Equal(t, &dept.ID, user.DepartmentID)
	}

	// 4. Get specific user by ID
	fetchedUser, err := client.GetUser(ctx, user.ID.String())
	require.NoError(t, err)
	require.NotNil(t, fetchedUser)
	assert.Equal(t, user.ID, fetchedUser.ID)
	assert.Equal(t, user.FirstName, fetchedUser.FirstName)

	// 5. Update the user with patch
	suspended := true
	newFirstName := fmt.Sprintf("Jane_%s", randomSuffix)

	nr := 3
	patchReq := PatchUserRequest{
		FirstName: &newFirstName,
		LastName:  &newFirstName,
		Suspended: &suspended,
		Role:      &nr,
	}

	patchedUser, err := client.PatchUser(ctx, user.ID.String(), patchReq)
	require.NoError(t, err)
	require.NotNil(t, patchedUser)

	// Verify user was patched with correct data
	assert.Equal(t, newFirstName, patchedUser.FirstName)
	assert.Equal(t, suspended, patchedUser.Suspended)

	// 6. Register credentials for the user with unique username
	credentialsReq := RegisterUserRequest{
		Username: fmt.Sprintf("johndoe_%s", uuid.Must(uuid.NewV7()).String()),
		Password: "password123",
	}

	err = client.RegisterUser(ctx, user.ID.String(), credentialsReq)
	require.NoError(t, err)

	// 7. Test login with the new credentials
	userClient := NewTestClient()
	token, err := userClient.Login(ctx, credentialsReq.Username, credentialsReq.Password)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// 8. Get all users and verify count increased
	allUsers, err := client.GetUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, allUsers, initialCount+1)

	// Find our user in the list
	var found bool
	for _, u := range allUsers {
		if u.ID == user.ID {
			found = true
			assert.Equal(t, newFirstName, u.FirstName) // Should have updated first name
			assert.Equal(t, suspended, u.Suspended)    // Should be suspended
			break
		}
	}
	assert.True(t, found, "Newly created user not found in users list")
}
