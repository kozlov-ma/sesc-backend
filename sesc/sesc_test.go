package sesc

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func requireDepartmentMatches(t *testing.T, expected, actual Department) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "Department ID mismatch")
	require.Equal(t, expected.Name, actual.Name, "Department name mismatch")
	require.Equal(t, expected.Description, actual.Description, "Department description mismatch")
}

func requireUserMatches(t *testing.T, expected, actual User) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "User ID mismatch")
	require.Equal(t, expected.FirstName, actual.FirstName, "User FirstName mismatch")
	require.Equal(t, expected.LastName, actual.LastName, "User LastName mismatch")

	// Only check department if expected has one
	if expected.Department.ID != uuid.Nil {
		require.Equal(
			t,
			expected.Department.ID,
			actual.Department.ID,
			"User Department.ID mismatch",
		)
	}

	if expected.Role.ID != 0 {
		require.Equal(t, expected.Role.ID, actual.Role.ID, "User Role.ID mismatch")
	}

	if expected.PictureURL != "" {
		require.Equal(t, expected.PictureURL, actual.PictureURL, "User PictureURL mismatch")
	}
}

func setupSESC(t *testing.T) *SESC {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		_ = client.Close()
	})
	return New(client)
}

func TestCreateDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")

		svc = setupSESC(t)
		return ctx, svc
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc := setup(t)

		name := "HR"
		desc := "Human Resources"

		dep, err := svc.CreateDepartment(ctx, name, desc)
		expected := Department{ID: dep.ID, Name: name, Description: desc}
		require.NoError(t, err, "CreateDepartment failed")
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("duplicate id", func(t *testing.T) {
		ctx, svc := setup(t)

		_, _ = svc.CreateDepartment(ctx, "IT", "IT Dept")
		// Trying to create another department with the same name
		_, err := svc.CreateDepartment(ctx, "IT", "Duplicate Dept")
		require.ErrorIs(t, err, ErrInvalidDepartment)
	})
}

func TestDeleteDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, id UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		dep, _ := svc.CreateDepartment(ctx, "Test", "Test Dept")
		return ctx, svc, dep.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, id := setup(t)

		err := svc.DeleteDepartment(ctx, id)
		require.NoError(t, err, "DeleteDepartment failed")

		_, err = svc.DepartmentByID(ctx, id)
		require.ErrorIs(t, err, ErrInvalidDepartment)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		err := svc.DeleteDepartment(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, ErrInvalidDepartment)
	})

	t.Run("with dependent users", func(t *testing.T) {
		ctx, svc, depID := setup(t)

		// Create a user with this department
		opt := UserUpdateOptions{
			FirstName:    "John",
			LastName:     "Doe",
			DepartmentID: depID,
			NewRoleID:    1,
		}
		_, err := svc.CreateUser(ctx, opt)
		require.NoError(t, err)

		err = svc.DeleteDepartment(ctx, depID)
		require.ErrorIs(t, err, ErrCannotRemoveDepartment)
	})
}

func TestDepartmentByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, id UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		name := "Test"
		desc := "Test Dept"
		dep, _ := svc.CreateDepartment(ctx, name, desc)
		return ctx, svc, dep.ID
	}

	t.Run("existing department", func(t *testing.T) {
		ctx, svc, id := setup(t)

		dep, err := svc.DepartmentByID(ctx, id)
		require.NoError(t, err, "DepartmentByID failed")

		expected := Department{ID: id, Name: "Test", Description: "Test Dept"}
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		_, err := svc.DepartmentByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, ErrInvalidDepartment)
	})
}

func TestGetAllDepartments(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		return ctx, svc
	}

	t.Run("empty", func(t *testing.T) {
		ctx, svc := setup(t)

		deps, err := svc.Departments(ctx)
		require.NoError(t, err, "Departments failed")
		require.Empty(t, deps, "Expected 0 departments")
	})

	t.Run("multiple departments", func(t *testing.T) {
		ctx, svc := setup(t)

		// Create departments
		expectedDeps := make([]Department, 2)
		for i := range expectedDeps {
			name := fmt.Sprintf("Dep %d", i)
			desc := "Desc"
			dep, err := svc.CreateDepartment(ctx, name, desc)
			require.NoError(t, err)
			expectedDeps[i] = dep
		}

		deps, err := svc.Departments(ctx)
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

func TestCreateUser(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, depID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		dep, _ := svc.CreateDepartment(ctx, "Dep", "Dep")
		depID = dep.ID
		return ctx, svc, depID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, depID := setup(t)

		opts := UserUpdateOptions{
			FirstName:    "John",
			LastName:     "Doe",
			DepartmentID: depID,
			NewRoleID:    1,
		}

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err, "CreateUser failed")

		expected := User{
			ID:         user.ID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: Department{ID: depID},
			Role:       Role{ID: 1},
		}
		requireUserMatches(t, expected, user)

		savedUser, err := svc.UserByID(ctx, user.ID)
		require.NoError(t, err)
		requireUserMatches(t, expected, savedUser)

		us, err := svc.Users(ctx)
		require.NoError(t, err)
		require.Len(t, us, 1)
	})

	t.Run("without_department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		opts := UserUpdateOptions{
			FirstName: "John",
			LastName:  "Doe",
			NewRoleID: 1,
		}

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err, "CreateUser failed")

		expected := User{
			ID:         user.ID, // Use the ID returned by CreateUser
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: NoDepartment,
			Role:       Role{ID: 1},
		}
		requireUserMatches(t, expected, user)

		// Verify user is retrievable
		savedUser, err := svc.UserByID(ctx, user.ID)
		require.NoError(t, err)
		requireUserMatches(t, expected, savedUser)
	})

	t.Run("invalid department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		opts := UserUpdateOptions{
			FirstName:    "Jane",
			LastName:     "Doe",
			DepartmentID: uuid.Must(uuid.NewV7()),
			NewRoleID:    1,
		}

		_, err := svc.CreateUser(ctx, opts)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidDepartment)
	})
}

func TestUpdateDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, id UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		dep, _ := svc.CreateDepartment(ctx, "Old", "Old Desc")
		id = dep.ID
		return ctx, svc, id
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, id := setup(t)

		newName, newDesc := "New", "New Desc"
		err := svc.UpdateDepartment(ctx, id, newName, newDesc)
		require.NoError(t, err, "UpdateDepartment failed")

		dep, err := svc.DepartmentByID(ctx, id)
		require.NoError(t, err)

		expected := Department{ID: id, Name: newName, Description: newDesc}
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		err := svc.UpdateDepartment(ctx, uuid.Must(uuid.NewV7()), "Name", "Desc")
		require.ErrorIs(t, err, ErrInvalidDepartment)
	})
}

func TestUpdateProfilePicture(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, userID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create a user
		opts := UserUpdateOptions{
			FirstName: "John",
			LastName:  "Doe",
			NewRoleID: 1,
		}

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err)

		return ctx, svc, user.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, userID := setup(t)

		newURL := "http://example.com/new.jpg"
		err := svc.UpdateProfilePicture(ctx, userID, newURL)
		require.NoError(t, err, "UpdateProfilePicture failed")

		user, err := svc.UserByID(ctx, userID)
		require.NoError(t, err)

		expected := User{
			ID:         userID,
			FirstName:  "John",
			LastName:   "Doe",
			Role:       Role{ID: 1},
			PictureURL: newURL,
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		err := svc.UpdateProfilePicture(ctx, uuid.Must(uuid.NewV7()), "url")
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestUpdateUser(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, depID UUID, userID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create department
		dep, _ := svc.CreateDepartment(ctx, "Dep", "Dep")
		depID = dep.ID

		// Create user
		opts := UserUpdateOptions{
			FirstName:    "Original",
			LastName:     "User",
			DepartmentID: depID,
			NewRoleID:    1,
		}

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err)

		return ctx, svc, depID, user.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, depID, userID := setup(t)
		opts := UserUpdateOptions{
			FirstName:    "Updated",
			LastName:     "User",
			DepartmentID: depID,
			NewRoleID:    2,
		}

		user, err := svc.UpdateUser(ctx, userID, opts)
		require.NoError(t, err, "UpdateUser failed")

		expected := User{
			ID:         userID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: Department{ID: depID},
			Role:       Role{ID: opts.NewRoleID},
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, svc, _, _ := setup(t)
		_, err := svc.UpdateUser(ctx, uuid.Must(uuid.NewV7()), UserUpdateOptions{})
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("invalid department", func(t *testing.T) {
		ctx, svc, _, userID := setup(t)
		opts := UserUpdateOptions{
			FirstName:    "Updated",
			LastName:     "User",
			DepartmentID: uuid.Must(uuid.NewV7()),
			NewRoleID:    1,
		}
		_, err := svc.UpdateUser(ctx, userID, opts)
		require.ErrorIs(t, err, ErrInvalidDepartment)
	})

	t.Run("remove department", func(t *testing.T) {
		ctx, svc, _, userID := setup(t)
		opts := UserUpdateOptions{
			FirstName: "Updated",
			LastName:  "User",
			NewRoleID: 2,
		}
		res, err := svc.UpdateUser(ctx, userID, opts)
		require.NoError(t, err)

		expected := User{
			ID:         userID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: NoDepartment,
			Role:       Role{ID: opts.NewRoleID},
		}
		requireUserMatches(t, expected, res)
	})

	t.Run("invalid role", func(t *testing.T) {
		ctx, svc, _, userID := setup(t)
		opts := UserUpdateOptions{
			FirstName: "Updated",
			LastName:  "User",
			NewRoleID: 999,
		}
		_, err := svc.UpdateUser(ctx, userID, opts)
		require.ErrorIs(t, err, ErrInvalidRole)
	})
}

func TestUserByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, userID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create user
		opts := UserUpdateOptions{
			FirstName: "John",
			LastName:  "Doe",
			NewRoleID: 1,
		}

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err)

		return ctx, svc, user.ID
	}

	t.Run("existing user", func(t *testing.T) {
		ctx, svc, userID := setup(t)

		user, err := svc.UserByID(ctx, userID)
		require.NoError(t, err, "UserByID failed")

		expected := User{
			ID:        userID,
			FirstName: "John",
			LastName:  "Doe",
			Role:      Role{ID: 1},
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		_, err := svc.UserByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestGetAllUsers(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create some users
		for i := range 2 {
			opts := UserUpdateOptions{
				FirstName: fmt.Sprintf("User%d", i+1),
				LastName:  fmt.Sprintf("User%d", i+1),
				NewRoleID: 1,
			}
			_, err := svc.CreateUser(ctx, opts)
			require.NoError(t, err)
		}

		return ctx, svc
	}

	t.Run("fetch all users", func(t *testing.T) {
		ctx, svc := setup(t)

		users, err := svc.Users(ctx)
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
