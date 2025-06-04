package depsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestCreateDepartment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Test department data
		name := "Test Department"
		description := "Department created for testing"

		// Call the method being tested
		department, err := svc.CreateDepartment(ctx, name, description)

		// Verify the results
		require.NoError(t, err)
		require.NotEmpty(t, department.ID)
		require.Equal(t, name, department.Name)
		require.Equal(t, description, department.Description)

		// Verify department was actually created in the database
		createdDepartment, err := svc.DepartmentByID(ctx, department.ID)
		require.NoError(t, err)
		require.Equal(t, department.ID, createdDepartment.ID)
		require.Equal(t, name, createdDepartment.Name)
		require.Equal(t, description, createdDepartment.Description)
	})

	t.Run("duplicate_name", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Test department data
		name := "Duplicate Department"
		description := "Department created for testing duplicates"

		// Create the first department
		firstDepartment, err := svc.CreateDepartment(ctx, name, description)
		require.NoError(t, err)
		require.NotEmpty(t, firstDepartment.ID)

		// Try to create another department with the same name
		_, err = svc.CreateDepartment(ctx, name, "Different description")

		// Verify the results
		require.Error(t, err)
		require.Equal(t, sesc.ErrInvalidDepartmentName, err)
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

		// Test department data
		name := "Error Department"
		description := "Department for testing database errors"

		// Call the method being tested
		_, err := svc.CreateDepartment(ctx, name, description)

		// Verify the results
		require.Error(t, err)
	})
}
