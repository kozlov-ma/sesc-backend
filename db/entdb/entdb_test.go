package entdb

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/sesc"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// Helper functions for validation
func requireDepartmentMatches(t *testing.T, expected, actual sesc.Department) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "Department ID mismatch")
	require.Equal(t, expected.Name, actual.Name, "Department name mismatch")
	require.Equal(t, expected.Description, actual.Description, "Department description mismatch")
}

func requireUserMatches(t *testing.T, expected, actual sesc.User) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "User ID mismatch")
	require.Equal(t, expected.FirstName, actual.FirstName, "User FirstName mismatch")
	require.Equal(t, expected.LastName, actual.LastName, "User LastName mismatch")

	// Only check department if expected has one
	if expected.Department.ID != uuid.Nil {
		require.Equal(t, expected.Department.ID, actual.Department.ID, "User Department.ID mismatch")
	}

	if expected.Role.ID != 0 {
		require.Equal(t, expected.Role.ID, actual.Role.ID, "User Role.ID mismatch")
	}

	if expected.PictureURL != "" {
		require.Equal(t, expected.PictureURL, actual.PictureURL, "User PictureURL mismatch")
	}
}

func setupDB(t *testing.T) *DB {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		_ = client.Close()
	})
	return New(slog.New(slog.DiscardHandler), client)
}

func TestCreateDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB) {
		ctx = context.Background()
		db = setupDB(t)
		return ctx, db
	}

	t.Run("success", func(t *testing.T) {
		ctx, db := setup(t)

		id := uuid.Must(uuid.NewV7())
		name := "HR"
		desc := "Human Resources"
		expected := sesc.Department{ID: id, Name: name, Description: desc}

		dep, err := db.CreateDepartment(ctx, id, name, desc)
		require.NoError(t, err, "CreateDepartment failed")
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("duplicate id", func(t *testing.T) {
		ctx, db := setup(t)

		id := uuid.Must(uuid.NewV7())
		_, _ = db.CreateDepartment(ctx, id, "IT", "IT Dept")
		_, err := db.CreateDepartment(ctx, id, "Duplicate", "Duplicate Dept")
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestDeleteDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB, id uuid.UUID) {
		ctx = context.Background()
		db = setupDB(t)
		id = uuid.Must(uuid.NewV7())
		_, _ = db.CreateDepartment(ctx, id, "Test", "Test Dept")
		return ctx, db, id
	}

	t.Run("success", func(t *testing.T) {
		ctx, db, id := setup(t)

		err := db.DeleteDepartment(ctx, id)
		require.NoError(t, err, "DeleteDepartment failed")

		_, err = db.DepartmentByID(ctx, id)
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, db, _ := setup(t)

		err := db.DeleteDepartment(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})

	t.Run("with dependent users", func(t *testing.T) {
		ctx, db, depID := setup(t)

		userID := uuid.Must(uuid.NewV7())
		db.c.User.Create().
			SetID(userID).
			SetFirstName("John").
			SetLastName("Doe").
			SetDepartmentID(depID).
			SetRoleID(1).
			ExecX(ctx)

		err := db.DeleteDepartment(ctx, depID)
		require.ErrorIs(t, err, sesc.ErrCannotRemoveDepartment)
	})
}

func TestDepartmentByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB, id uuid.UUID) {
		ctx = context.Background()
		db = setupDB(t)
		id = uuid.Must(uuid.NewV7())
		name := "Test"
		desc := "Test Dept"
		_, _ = db.CreateDepartment(ctx, id, name, desc)
		return ctx, db, id
	}

	t.Run("existing department", func(t *testing.T) {
		ctx, db, id := setup(t)

		dep, err := db.DepartmentByID(ctx, id)
		require.NoError(t, err, "DepartmentByID failed")

		expected := sesc.Department{ID: id, Name: "Test", Description: "Test Dept"}
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, db, _ := setup(t)

		_, err := db.DepartmentByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestDepartments(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB) {
		ctx = context.Background()
		db = setupDB(t)
		return ctx, db
	}

	t.Run("empty", func(t *testing.T) {
		ctx, db := setup(t)

		deps, err := db.Departments(ctx)
		require.NoError(t, err, "Departments failed")
		require.Empty(t, deps, "Expected 0 departments")
	})

	t.Run("multiple departments", func(t *testing.T) {
		ctx, db := setup(t)

		// Create departments
		expectedDeps := make([]sesc.Department, 2)
		for i := range expectedDeps {
			id := uuid.Must(uuid.NewV7())
			name := fmt.Sprintf("Dep %s", id)
			desc := "Desc"
			dep, err := db.CreateDepartment(ctx, id, name, desc)
			require.NoError(t, err)
			expectedDeps[i] = dep
		}

		deps, err := db.Departments(ctx)
		require.NoError(t, err, "Departments failed")
		require.Len(t, deps, len(expectedDeps), "Unexpected number of departments")

		// Verify that each created department exists in the result
		for _, expected := range expectedDeps {
			found := false
			for _, actual := range deps {
				if actual.ID == expected.ID {
					requireDepartmentMatches(t, expected, actual)
					found = true
					break
				}
			}
			require.True(t, found, "Created department not found in results")
		}
	})
}

func TestSaveUser(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB, depID uuid.UUID) {
		ctx = context.Background()
		db = setupDB(t)
		depID = uuid.Must(uuid.NewV7())
		_, _ = db.CreateDepartment(ctx, depID, "Dep", "Dep")
		return ctx, db, depID
	}

	t.Run("success", func(t *testing.T) {
		ctx, db, depID := setup(t)

		opts := sesc.UserUpdateOptions{
			FirstName:    "John",
			LastName:     "Doe",
			DepartmentID: depID,
			NewRoleID:    1,
		}

		user, err := db.SaveUser(ctx, opts)
		require.NoError(t, err, "SaveUser failed")

		expected := sesc.User{
			ID:         user.ID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: sesc.Department{ID: depID},
			Role:       sesc.Role{ID: 1},
		}
		requireUserMatches(t, expected, user)

		savedUser, err := db.UserByID(ctx, user.ID)
		require.NoError(t, err)
		requireUserMatches(t, expected, savedUser)
	})

	t.Run("without_department", func(t *testing.T) {
		ctx, db, _ := setup(t)

		opts := sesc.UserUpdateOptions{
			FirstName: "John",
			LastName:  "Doe",
			NewRoleID: 1,
		}

		user, err := db.SaveUser(ctx, opts)
		require.NoError(t, err, "SaveUser failed")

		expected := sesc.User{
			ID:         user.ID, // Use the ID returned by SaveUser
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: sesc.NoDepartment,
			Role:       sesc.Role{ID: 1},
		}
		requireUserMatches(t, expected, user)

		// Verify user is retrievable
		savedUser, err := db.UserByID(ctx, user.ID)
		require.NoError(t, err)
		requireUserMatches(t, expected, savedUser)
	})

	t.Run("invalid department", func(t *testing.T) {
		ctx, db, _ := setup(t)

		opts := sesc.UserUpdateOptions{
			FirstName:    "Jane",
			LastName:     "Doe",
			DepartmentID: uuid.Must(uuid.NewV7()),
			NewRoleID:    1,
		}

		_, err := db.SaveUser(ctx, opts)
		require.Error(t, err)
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestUpdateDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB, id uuid.UUID) {
		ctx = context.Background()
		db = setupDB(t)
		id = uuid.Must(uuid.NewV7())
		_, _ = db.CreateDepartment(ctx, id, "Old", "Old Desc")
		return ctx, db, id
	}

	t.Run("success", func(t *testing.T) {
		ctx, db, id := setup(t)

		newName, newDesc := "New", "New Desc"
		err := db.UpdateDepartment(ctx, id, newName, newDesc)
		require.NoError(t, err, "UpdateDepartment failed")

		dep, err := db.DepartmentByID(ctx, id)
		require.NoError(t, err)

		expected := sesc.Department{ID: id, Name: newName, Description: newDesc}
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, db, _ := setup(t)

		err := db.UpdateDepartment(ctx, uuid.Must(uuid.NewV7()), "Name", "Desc")
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestUpdateProfilePicture(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB, userID uuid.UUID) {
		ctx = context.Background()
		db = setupDB(t)
		userID = uuid.Must(uuid.NewV7())
		db.c.User.Create().
			SetID(userID).
			SetFirstName("John").
			SetLastName("Doe").
			SetRoleID(1).
			ExecX(ctx)
		return ctx, db, userID
	}

	t.Run("success", func(t *testing.T) {
		ctx, db, userID := setup(t)

		newURL := "http://example.com/new.jpg"
		err := db.UpdateProfilePicture(ctx, userID, newURL)
		require.NoError(t, err, "UpdateProfilePicture failed")

		user, err := db.UserByID(ctx, userID)
		require.NoError(t, err)

		expected := sesc.User{
			ID:         userID,
			FirstName:  "John",
			LastName:   "Doe",
			Role:       sesc.Role{ID: 1},
			PictureURL: newURL,
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, db, _ := setup(t)

		err := db.UpdateProfilePicture(ctx, uuid.Must(uuid.NewV7()), "url")
		require.ErrorIs(t, err, sesc.ErrUserNotFound)
	})
}

func TestUpdateUser(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB, depID uuid.UUID, userID uuid.UUID) {
		ctx = context.Background()
		db = setupDB(t)
		depID = uuid.Must(uuid.NewV7())
		_, _ = db.CreateDepartment(ctx, depID, "Dep", "Dep")
		userID = uuid.Must(uuid.NewV7())
		db.c.User.Create().
			SetID(userID).
			SetFirstName("Original").
			SetLastName("User").
			SetDepartmentID(depID).
			SetRoleID(1).
			ExecX(ctx)

		return ctx, db, depID, userID
	}

	t.Run("success", func(t *testing.T) {
		ctx, db, depID, userID := setup(t)
		opts := sesc.UserUpdateOptions{
			FirstName:    "Updated",
			LastName:     "User",
			DepartmentID: depID,
			NewRoleID:    2,
		}

		user, err := db.UpdateUser(ctx, userID, opts)
		require.NoError(t, err, "UpdateUser failed")

		expected := sesc.User{
			ID:         userID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: sesc.Department{ID: depID},
			Role:       sesc.Role{ID: opts.NewRoleID},
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, db, _, _ := setup(t)
		_, err := db.UpdateUser(ctx, uuid.Must(uuid.NewV7()), sesc.UserUpdateOptions{})
		require.ErrorIs(t, err, sesc.ErrUserNotFound)
	})

	t.Run("invalid department", func(t *testing.T) {
		ctx, db, _, userID := setup(t)
		opts := sesc.UserUpdateOptions{DepartmentID: uuid.Must(uuid.NewV7())}
		_, err := db.UpdateUser(ctx, userID, opts)
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})

	t.Run("remove department", func(t *testing.T) {
		ctx, db, _, userID := setup(t)
		opts := sesc.UserUpdateOptions{
			FirstName: "Updated",
			LastName:  "User",
			NewRoleID: 2,
		}
		res, err := db.UpdateUser(ctx, userID, opts)
		require.NoError(t, err)

		expected := sesc.User{
			ID:         userID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: sesc.NoDepartment,
			Role:       sesc.Role{ID: opts.NewRoleID},
		}
		requireUserMatches(t, expected, res)
	})

	t.Run("invalid role", func(t *testing.T) {
		ctx, db, _, userID := setup(t)
		opts := sesc.UserUpdateOptions{NewRoleID: 999}
		_, err := db.UpdateUser(ctx, userID, opts)
		require.ErrorIs(t, err, sesc.ErrInvalidRole)
	})
}

func TestUserByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB, userID uuid.UUID) {
		ctx = context.Background()
		db = setupDB(t)
		userID = uuid.Must(uuid.NewV7())
		db.c.User.Create().
			SetID(userID).
			SetFirstName("John").
			SetLastName("Doe").
			SetRoleID(1).
			ExecX(ctx)
		return ctx, db, userID
	}

	t.Run("existing user", func(t *testing.T) {
		ctx, db, userID := setup(t)

		user, err := db.UserByID(ctx, userID)
		require.NoError(t, err, "UserByID failed")

		expected := sesc.User{
			ID:        userID,
			FirstName: "John",
			LastName:  "Doe",
			Role:      sesc.Role{ID: 1},
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, db, _ := setup(t)

		_, err := db.UserByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, sesc.ErrUserNotFound)
	})
}

func TestUsers(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, db *DB) {
		ctx = context.Background()
		db = setupDB(t)

		user1ID := uuid.Must(uuid.NewV7())
		db.c.User.Create().
			SetID(user1ID).
			SetFirstName("User1").
			SetLastName("User1").
			SetRoleID(1).
			ExecX(ctx)

		user2ID := uuid.Must(uuid.NewV7())
		db.c.User.Create().
			SetID(user2ID).
			SetFirstName("User2").
			SetLastName("User2").
			SetRoleID(1).
			ExecX(ctx)

		return ctx, db
	}

	t.Run("fetch all users", func(t *testing.T) {
		ctx, db := setup(t)

		users, err := db.Users(ctx)
		require.NoError(t, err, "Users failed")
		require.Len(t, users, 2, "Expected 2 users")

		// Verify user fields
		for _, user := range users {
			require.NotEqual(t, uuid.Nil, user.ID, "User ID should not be nil")
			require.NotEmpty(t, user.FirstName, "User FirstName should not be empty")
			require.NotEmpty(t, user.LastName, "User LastName should not be empty")
			require.Equal(t, int32(1), user.Role.ID, "User Role.ID should be 1")
		}
	})
}
