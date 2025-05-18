package tests

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepartmentCRUD(t *testing.T) {
	// Skip if test API URL is not set
	SkipIfNoTestAPI(t)

	// Create a client for the test API
	client := NewTestClient()
	ctx := t.Context()

	adminToken, err := client.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	client.SetToken(adminToken)

	// 1. Get initial departments (should be empty)
	initialDepts, err := client.GetDepartments(ctx)
	require.NoError(t, err)
	initialCount := len(initialDepts)

	// 2. Create a department with a unique random name
	uniqueID := uuid.Must(uuid.NewV7()).String()
	createReq := CreateDepartmentRequest{
		Name:        fmt.Sprintf("Test Department %s", uniqueID),
		Description: fmt.Sprintf("This is a test department %s", uniqueID),
	}

	createdDept, err := client.CreateDepartment(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createdDept)

	// Verify department was created with correct data
	assert.Equal(t, createReq.Name, createdDept.Name)
	assert.Equal(t, createReq.Description, createdDept.Description)
	assert.NotEqual(t, uuid.Nil, createdDept.ID)

	// 3. Get all departments and verify the new one is there
	depts, err := client.GetDepartments(ctx)
	require.NoError(t, err)
	assert.Len(t, depts, initialCount+1)

	// Find our department in the list
	var found bool
	for _, dept := range depts {
		if dept.ID == createdDept.ID {
			found = true
			assert.Equal(t, createReq.Name, dept.Name)
			assert.Equal(t, createReq.Description, dept.Description)
			break
		}
	}
	assert.True(t, found, "Newly created department not found in departments list")

	// 4. Update the department
	updateReq := UpdateDepartmentRequest{
		Name:        fmt.Sprintf("Updated Department %s", uniqueID),
		Description: fmt.Sprintf("This department has been updated %s", uniqueID),
	}

	updatedDept, err := client.UpdateDepartment(ctx, createdDept.ID.String(), updateReq)
	require.NoError(t, err)
	require.NotNil(t, updatedDept)

	// Verify department was updated with correct data
	assert.Equal(t, updateReq.Name, updatedDept.Name)
	assert.Equal(t, updateReq.Description, updatedDept.Description)
	assert.Equal(t, createdDept.ID, updatedDept.ID)

	// 5. Delete the department
	err = client.DeleteDepartment(ctx, createdDept.ID.String())
	require.NoError(t, err)

	// 6. Verify department was deleted
	finalDepts, err := client.GetDepartments(ctx)
	require.NoError(t, err)
	assert.Len(t, finalDepts, initialCount)

	// Verify our department is not in the list
	for _, dept := range finalDepts {
		assert.NotEqual(t, createdDept.ID, dept.ID, "Department should have been deleted")
	}
}
