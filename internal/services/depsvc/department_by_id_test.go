package depsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestDepartmentByID(t *testing.T) {
	t.Run("existing_department", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test department
		name := "Test Department"
		description := "Department for testing retrieval"

		createdDept, err := svc.CreateDepartment(ctx, name, description)
		require.NoError(t, err)
		require.NotEmpty(t, createdDept.ID)

		// Call the method being tested
		retrievedDept, err := svc.DepartmentByID(ctx, createdDept.ID)

		// Verify the results
		require.NoError(t, err)
		require.Equal(t, createdDept.ID, retrievedDept.ID)
		require.Equal(t, name, retrievedDept.Name)
		require.Equal(t, description, retrievedDept.Description)
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
		_, err := svc.DepartmentByID(ctx, nonExistentID)

		// Verify the results
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
		name := "Error Department"
		description := "Department for testing database errors"

		createdDept, err := svc.CreateDepartment(ctx, name, description)
		require.NoError(t, err)
		require.NotEmpty(t, createdDept.ID)

		// Close the database to force errors
		client.Close()

		// Call the method being tested
		_, err = svc.DepartmentByID(ctx, createdDept.ID)

		// Verify the results
		require.Error(t, err)
	})
}
