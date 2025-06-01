package depsvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestDepartments(t *testing.T) {
	t.Run("get_all_departments", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create test departments
		dept1, err := svc.CreateDepartment(ctx, "Department 1", "First test department")
		require.NoError(t, err)

		dept2, err := svc.CreateDepartment(ctx, "Department 2", "Second test department")
		require.NoError(t, err)

		dept3, err := svc.CreateDepartment(ctx, "Department 3", "Third test department")
		require.NoError(t, err)

		// Call the method being tested
		departments, err := svc.Departments(ctx)

		// Verify the results
		require.NoError(t, err)
		require.Len(t, departments, 3)

		// Create a map of department IDs for easier verification
		deptMap := make(map[string]Department)
		for _, d := range departments {
			deptMap[d.ID.String()] = d
		}

		// Verify each department is in the results
		d1, exists := deptMap[dept1.ID.String()]
		require.True(t, exists)
		require.Equal(t, dept1.Name, d1.Name)
		require.Equal(t, dept1.Description, d1.Description)

		d2, exists := deptMap[dept2.ID.String()]
		require.True(t, exists)
		require.Equal(t, dept2.Name, d2.Name)
		require.Equal(t, dept2.Description, d2.Description)

		d3, exists := deptMap[dept3.ID.String()]
		require.True(t, exists)
		require.Equal(t, dept3.Name, d3.Name)
		require.Equal(t, dept3.Description, d3.Description)
	})

	t.Run("empty_department_list", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Call the method being tested without creating any departments
		departments, err := svc.Departments(ctx)

		// Verify the results
		require.NoError(t, err)
		require.Empty(t, departments)
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

		// Call the method being tested
		departments, err := svc.Departments(ctx)

		// Verify the results
		require.Error(t, err)
		require.Empty(t, departments)
	})
}
