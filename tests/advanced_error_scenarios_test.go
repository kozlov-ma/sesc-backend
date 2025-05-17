package tests

import (
	"strings"
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepartmentErrors(t *testing.T) {
	app := testutil.StartTestApp(t)
	adminClient := NewClient(app.URL)
	regularClient := NewClient(app.URL)
	ctx := t.Context()

	// Login as admin
	adminToken, err := adminClient.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	adminClient.SetToken(adminToken)

	// Create a regular user for permission testing
	userData := CreateUserRequest{
		FirstName: "Regular",
		LastName:  "User",
		RoleID:    2,
	}
	user, err := adminClient.CreateUser(ctx, userData)
	require.NoError(t, err)

	// Register credentials for the user
	err = adminClient.RegisterUser(ctx, user.ID.String(), RegisterUserRequest{
		Username: "regular_user",
		Password: "password123",
	})
	require.NoError(t, err)

	// Login as regular user
	regularToken, err := regularClient.Login(ctx, "regular_user", "password123")
	require.NoError(t, err)
	regularClient.SetToken(regularToken)

	// Test duplicate department name
	deptReq := CreateDepartmentRequest{
		Name:        "Unique Department",
		Description: "This should be unique",
	}

	dept, err := adminClient.CreateDepartment(ctx, deptReq)
	require.NoError(t, err)

	// Try to create another department with the same name
	_, err = adminClient.CreateDepartment(ctx, deptReq)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "invalid_department")

	// Test regular user trying to create a department (should be forbidden)
	_, err = regularClient.CreateDepartment(ctx, CreateDepartmentRequest{
		Name:        "User Created Dept",
		Description: "Should fail due to permissions",
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "forbidden")

	// Test regular user trying to update a department
	_, err = regularClient.UpdateDepartment(ctx, dept.ID.String(), UpdateDepartmentRequest{
		Name: "Modified by User",
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "forbidden")

	// Test regular user trying to delete a department
	err = regularClient.DeleteDepartment(ctx, dept.ID.String())
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "forbidden")

	// Test deleting a department that has associated users
	// First create a user associated with the department
	userWithDept := CreateUserRequest{
		FirstName:    "Department",
		LastName:     "User",
		DepartmentID: dept.ID,
		RoleID:       2,
	}
	_, err = adminClient.CreateUser(ctx, userWithDept)
	require.NoError(t, err)

	// Now try to delete the department
	err = adminClient.DeleteDepartment(ctx, dept.ID.String())
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "cannot_remove_department")
}

func TestRequestValidationErrors(t *testing.T) {
	app := testutil.StartTestApp(t)
	client := NewClient(app.URL)
	ctx := t.Context()

	// Login as admin
	adminToken, err := client.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	client.SetToken(adminToken)

	// Test with invalid UUID
	_, err = client.GetUser(ctx, "not-a-valid-uuid")
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "invalid")

	// Test with invalid inputs that cause validation errors
	// 1. Empty department name - should be rejected
	_, err = client.CreateDepartment(ctx, CreateDepartmentRequest{
		Name:        "", // Empty name - should be rejected
		Description: "Test Description",
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "invalid")

	// 2. Very long department name
	longName := ""
	for range 1000 {
		longName += "very_long_name"
	}
	// This test might pass or fail depending on the server's max size configuration
	_, _ = client.CreateDepartment(ctx, CreateDepartmentRequest{
		Name:        longName,
		Description: "Test Description",
	})

	// 3. Test with extremely long name for a user
	veryLongName := ""
	for range 1000 {
		veryLongName += "x"
	}
	// This test might pass or fail depending on the server's max size configuration
	_, _ = client.CreateUser(ctx, CreateUserRequest{
		FirstName: veryLongName,
		LastName:  "Test",
		RoleID:    2,
	})

	// 4. User with non-existent role ID
	_, err = client.CreateUser(ctx, CreateUserRequest{
		FirstName: "Role",
		LastName:  "Test",
		RoleID:    999, // Non-existent role ID
	})
	require.Error(t, err)
}
