package depsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestDeleteDepartment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test department
		dept, err := svc.CreateDepartment(ctx, "Test Department", "Department for deletion test")
		require.NoError(t, err)
		require.NotEmpty(t, dept.ID)

		// Call the method being tested
		err = svc.DeleteDepartment(ctx, dept.ID)

		// Verify the results
		require.NoError(t, err)

		// Verify the department was actually deleted
		_, err = svc.DepartmentByID(ctx, dept.ID)
		require.Error(t, err)
		require.Equal(t, sesc.ErrDepartmentNotFound, err)
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
		err := svc.DeleteDepartment(ctx, nonExistentID)

		// Verify the results
		require.Error(t, err)
		require.Equal(t, sesc.ErrDepartmentNotFound, err)
	})

	t.Run("database_error", func(t *testing.T) {
		// Setup test context with database that will be closed to cause errors
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test department
		dept, err := svc.CreateDepartment(ctx, "Error Department", "Department for testing database errors")
		require.NoError(t, err)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		err = svc.DeleteDepartment(ctx, dept.ID)

		// Verify the results
		require.Error(t, err)
	})
}
