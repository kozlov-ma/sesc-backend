package usersvc

import (
	"testing"

	"github.com/kozlov-ma/sesc-backend/internal/services/testutil"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		opts := UserUpdateOptions{
			FirstName: "Test",
			LastName:  "User",
			NewRole:   sesc.Role(1),
		}
		created, err := svc.CreateUser(ctx, opts)

		require.NoError(t, err)

		require.Equal(t, "Test", created.FirstName)
		require.Nil(t, created.DepartmentID)
		require.Equal(t, sesc.Role(1), created.Role)
	})
	t.Run("incorrect_name", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		opts := UserUpdateOptions{
			LastName: "User",
			NewRole:  sesc.Role(2),
		}
		_, err := svc.CreateUser(ctx, opts)

		require.Error(t, err)
	})
	t.Run("with_department", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		dep := testutil.CreateTestDepartment(ctx, t, client)

		opts := UserUpdateOptions{
			FirstName:    "Test",
			LastName:     "User",
			DepartmentID: &dep.ID,
			NewRole:      sesc.Role(1),
		}
		created, err := svc.CreateUser(ctx, opts)

		require.NoError(t, err)
		require.Equal(t, &dep.ID, created.DepartmentID)
	})
	t.Run("incorrect_department", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		// Create a test user
		_ = testutil.CreateTestDepartment(ctx, t, client)

		nexistentUUID := testutil.RandomUUID()
		opts := UserUpdateOptions{
			FirstName:    "Test",
			LastName:     "User",
			DepartmentID: &nexistentUUID,
			NewRole:      sesc.Role(2),
		}
		_, err := svc.CreateUser(ctx, opts)

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
			Subdivision:    old.Subdivision,
			JobTitle:       old.JobTitle,
			EmploymentRate: old.EmploymentRate,
		}

		_, err := svc.CreateUser(ctx, opts)

		require.NoError(t, err)
	})
	t.Run("database_closed", func(t *testing.T) {
		// Setup test context with database
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		client := testutil.SetupDatabase(t)

		// Create the service
		svc := New(client)

		opts := UserUpdateOptions{
			FirstName: "Test",
			LastName:  "User",
			NewRole:   sesc.Role(2),
		}

		client.Close()

		_, err := svc.CreateUser(ctx, opts)
		require.Error(t, err)
	})
}
