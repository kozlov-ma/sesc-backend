package depsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestUpdateDepartment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test department
		originalName := "Original Department"
		originalDesc := "Original description"

		dept, err := svc.CreateDepartment(ctx, originalName, originalDesc)
		require.NoError(t, err)
		require.NotEmpty(t, dept.ID)

		// New values for update
		updatedName := "Updated Department"
		updatedDesc := "Updated description"

		// Call the method being tested
		err = svc.UpdateDepartment(ctx, dept.ID, updatedName, updatedDesc)

		// Verify the results
		require.NoError(t, err)

		// Verify the department was actually updated in the database
		updatedDept, err := svc.DepartmentByID(ctx, dept.ID)
		require.NoError(t, err)
		require.Equal(t, dept.ID, updatedDept.ID)
		require.Equal(t, updatedName, updatedDept.Name)
		require.Equal(t, updatedDesc, updatedDept.Description)
	})

	t.Run("non_existent_department", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Generate a random UUID that doesn't exist
		nonExistentID := testutil.RandomUUID()

		// Call the method being tested
		err := svc.UpdateDepartment(ctx, nonExistentID, "Updated Name", "Updated description")

		// Verify the results
		require.Error(t, err)
		// The error should contain both the not found error and the invalid department error
		require.Contains(t, err.Error(), sesc.ErrInvalidDepartment.Error())
	})

	t.Run("duplicate_name", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create two test departments
		dept1, err := svc.CreateDepartment(ctx, "Department 1", "First department")
		require.NoError(t, err)

		dept2, err := svc.CreateDepartment(ctx, "Department 2", "Second department")
		require.NoError(t, err)

		// Try to update dept2 with dept1's name
		err = svc.UpdateDepartment(ctx, dept2.ID, dept1.Name, "Updated description")

		// This should fail with a constraint error
		require.Error(t, err)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test department
		dept, err := svc.CreateDepartment(ctx, "Test Department", "Test description")
		require.NoError(t, err)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		err = svc.UpdateDepartment(ctx, dept.ID, "Updated Name", "Updated description")

		// Verify the results
		require.Error(t, err)
	})
}
