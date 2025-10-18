package usersvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestUpdateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		old := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		opts := UserUpdateOptions{
			FirstName: "TestUpdated",
			LastName:  "User",
			NewRole:   sesc.Role(2),
		}
		upd, err := svc.UpdateUser(ctx, old.ID, opts)

		require.NoError(t, err)

		require.NotEqual(t, old.FirstName, upd.FirstName)
		require.Equal(t, old.LastName, upd.LastName)
		require.Equal(t, old.MiddleName, upd.MiddleName)
		require.Nil(t, upd.DepartmentID)

		require.NotEqual(t, old.Role, upd.Role)
	})
	t.Run("incorrect_name", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		old := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		opts := UserUpdateOptions{
			LastName: "User",
			NewRole:  sesc.Role(2),
		}
		_, err := svc.UpdateUser(ctx, old.ID, opts)

		require.Error(t, err)
	})
	t.Run("with_department", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		old := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		dep := testutil.CreateTestDepartment(ctx, t, client)

		opts := UserUpdateOptions{
			FirstName:    "TestUpdated",
			LastName:     "User",
			DepartmentID: &dep.ID,
			NewRole:      sesc.Role(2),
		}
		upd, err := svc.UpdateUser(ctx, old.ID, opts)

		require.NoError(t, err)
		require.NotEqual(t, old.DepartmentID, upd.DepartmentID)
		require.Equal(t, opts.DepartmentID, upd.DepartmentID)
	})
	t.Run("non_existent_user", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		_ = testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		opts := UserUpdateOptions{
			FirstName: "TestUpdated",
			LastName:  "User",
			NewRole:   sesc.Role(2),
		}

		_, err := svc.UpdateUser(ctx, testutil.RandomUUID(), opts)
		require.Error(t, err)
	})
	t.Run("incorrect_department", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		old := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		_ = testutil.CreateTestDepartment(ctx, t, client)

		nexistentUUID := testutil.RandomUUID()
		opts := UserUpdateOptions{
			FirstName:    "TestUpdated",
			LastName:     "User",
			DepartmentID: &nexistentUUID,
			NewRole:      sesc.Role(2),
		}
		_, err := svc.UpdateUser(ctx, old.ID, opts)

		require.Error(t, err)
	})
	t.Run("with_similar_fields", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		old := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))
		opts := UserUpdateOptions{
			FirstName:      old.FirstName,
			LastName:       old.LastName,
			MiddleName:     old.MiddleName,
			PictureURL:     old.PictureURL,
			Suspended:      old.Suspended,
			DepartmentID:   old.DepartmentID,
			NewRole:        old.Role,
			JobTitle:       old.JobTitle,
			EmploymentRate: old.EmploymentRate,
		}

		_, err := svc.UpdateUser(ctx, old.ID, opts)

		require.NoError(t, err)
	})
	t.Run("database_closed", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		old := testutil.CreateTestUser(ctx, t, client, "Test", "User", sesc.Role(1))

		opts := UserUpdateOptions{
			FirstName: "TestUpdated",
			LastName:  "User",
			NewRole:   sesc.Role(2),
		}

		client.Close()

		_, err := svc.UpdateUser(ctx, old.ID, opts)
		require.Error(t, err)
	})
}
