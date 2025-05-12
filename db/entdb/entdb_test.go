package entdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/sesc"
	_ "github.com/mattn/go-sqlite3"
)

func setupDB(t *testing.T) *DB {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		client.Close()
	})
	return New(slog.New(slog.DiscardHandler), client)
}

func TestCreateDepartment(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)

	t.Run("success", func(t *testing.T) {
		id := uuid.Must(uuid.NewV7())
		name := "HR"
		desc := "Human Resources"

		dep, err := db.CreateDepartment(ctx, id, name, desc)
		if err != nil {
			t.Fatalf("CreateDepartment failed: %v", err)
		}

		if dep.ID != id || dep.Name != name || dep.Description != desc {
			t.Errorf("Expected department %v, got %v", sesc.Department{ID: id, Name: name, Description: desc}, dep)
		}
	})

	t.Run("duplicate id", func(t *testing.T) {
		id := uuid.Must(uuid.NewV7())
		_, _ = db.CreateDepartment(ctx, id, "IT", "IT Dept")
		_, err := db.CreateDepartment(ctx, id, "Duplicate", "Duplicate Dept")
		if !errors.Is(err, sesc.ErrInvalidDepartment) {
			t.Errorf("Expected ErrInvalidDepartment, got %v", err)
		}
	})
}

func TestDeleteDepartment(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	id := uuid.Must(uuid.NewV7())
	_, _ = db.CreateDepartment(ctx, id, "Test", "Test Dept")

	t.Run("success", func(t *testing.T) {
		err := db.DeleteDepartment(ctx, id)
		if err != nil {
			t.Fatalf("DeleteDepartment failed: %v", err)
		}

		_, err = db.DepartmentByID(ctx, id)
		if !errors.Is(err, sesc.ErrInvalidDepartment) {
			t.Errorf("Expected department to be deleted, got %v", err)
		}
	})

	t.Run("non-existent department", func(t *testing.T) {
		err := db.DeleteDepartment(ctx, uuid.Must(uuid.NewV7()))
		if !errors.Is(err, sesc.ErrInvalidDepartment) {
			t.Errorf("Expected ErrInvalidDepartment, got %v", err)
		}
	})

	t.Run("with dependent users", func(t *testing.T) {
		depID := uuid.Must(uuid.NewV7())
		_, _ = db.CreateDepartment(ctx, depID, "Dep", "Dep")

		userID := uuid.Must(uuid.NewV7())
		db.c.User.Create().
			SetID(userID).
			SetFirstName("John").
			SetLastName("Doe").
			SetDepartmentID(depID).
			SetRoleID(1).
			ExecX(ctx)

		err := db.DeleteDepartment(ctx, depID)
		if !errors.Is(err, sesc.ErrCannotRemoveDepartment) {
			t.Errorf("Expected ErrCannotRemoveDepartment, got %v", err)
		}
	})
}

func TestDepartmentByID(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	id := uuid.Must(uuid.NewV7())
	_, _ = db.CreateDepartment(ctx, id, "Test", "Test Dept")

	t.Run("existing department", func(t *testing.T) {
		dep, err := db.DepartmentByID(ctx, id)
		if err != nil {
			t.Fatalf("DepartmentByID failed: %v", err)
		}
		if dep.ID != id {
			t.Errorf("Expected department ID %s, got %s", id, dep.ID)
		}
	})

	t.Run("non-existent department", func(t *testing.T) {
		_, err := db.DepartmentByID(ctx, uuid.Must(uuid.NewV7()))
		if !errors.Is(err, sesc.ErrInvalidDepartment) {
			t.Errorf("Expected ErrInvalidDepartment, got %v", err)
		}
	})
}

func TestDepartments(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)

	t.Run("empty", func(t *testing.T) {
		deps, err := db.Departments(ctx)
		if err != nil {
			t.Fatalf("Departments failed: %v", err)
		}
		if len(deps) != 0 {
			t.Errorf("Expected 0 departments, got %d", len(deps))
		}
	})

	t.Run("multiple departments", func(t *testing.T) {
		ids := []sesc.UUID{
			uuid.Must(uuid.NewV7()),
			uuid.Must(uuid.NewV7()),
		}
		for _, id := range ids {
			_, _ = db.CreateDepartment(ctx, id, fmt.Sprintf("Dep %s", id), "Desc")
		}

		deps, err := db.Departments(ctx)
		if err != nil {
			t.Fatalf("Departments failed: %v", err)
		}
		if len(deps) != len(ids) {
			t.Errorf("Expected %d departments, got %d", len(ids), len(deps))
		}
	})
}

func TestSaveUser(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	depID := uuid.Must(uuid.NewV7())
	_, _ = db.CreateDepartment(ctx, depID, "Dep", "Dep")

	t.Run("success", func(t *testing.T) {
		user := sesc.User{
			ID:         uuid.Must(uuid.NewV7()),
			FirstName:  "John",
			LastName:   "Doe",
			Department: sesc.Department{ID: depID},
			Role:       sesc.Role{ID: 1},
		}

		err := db.SaveUser(ctx, user)
		if err != nil {
			t.Fatalf("SaveUser failed: %v", err)
		}
	})

	t.Run("invalid department", func(t *testing.T) {
		user := sesc.User{
			ID:         uuid.Must(uuid.NewV7()),
			FirstName:  "Jane",
			LastName:   "Doe",
			Department: sesc.Department{ID: uuid.Must(uuid.NewV7())},
			Role:       sesc.Role{ID: 1},
		}

		err := db.SaveUser(ctx, user)
		if err == nil || !strings.Contains(err.Error(), "constraint") {
			t.Errorf("Expected constraint error, got %v", err)
		}
	})
}

func TestUpdateDepartment(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	id := uuid.Must(uuid.NewV7())
	_, _ = db.CreateDepartment(ctx, id, "Old", "Old Desc")

	t.Run("success", func(t *testing.T) {
		newName, newDesc := "New", "New Desc"
		err := db.UpdateDepartment(ctx, id, newName, newDesc)
		if err != nil {
			t.Fatalf("UpdateDepartment failed: %v", err)
		}

		dep, _ := db.DepartmentByID(ctx, id)
		if dep.Name != newName || dep.Description != newDesc {
			t.Errorf("Department not updated: %v", dep)
		}
	})

	t.Run("non-existent department", func(t *testing.T) {
		err := db.UpdateDepartment(ctx, uuid.Must(uuid.NewV7()), "Name", "Desc")
		if !errors.Is(err, sesc.ErrInvalidDepartment) {
			t.Errorf("Expected ErrInvalidDepartment, got %v", err)
		}
	})
}

func TestUpdateProfilePicture(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	userID := uuid.Must(uuid.NewV7())
	db.c.User.Create().
		SetID(userID).
		SetFirstName("John").
		SetLastName("Doe").
		SetRoleID(1).
		ExecX(ctx)

	t.Run("success", func(t *testing.T) {
		newURL := "http://example.com/new.jpg"
		err := db.UpdateProfilePicture(ctx, userID, newURL)
		if err != nil {
			t.Fatalf("UpdateProfilePicture failed: %v", err)
		}

		user, _ := db.UserByID(ctx, userID)
		if user.PictureURL != newURL {
			t.Errorf("PictureURL not updated: got %s", user.PictureURL)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		err := db.UpdateProfilePicture(ctx, uuid.Must(uuid.NewV7()), "url")
		if !errors.Is(err, sesc.ErrUserNotFound) {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestUpdateUser(t *testing.T) {
	t.Skip("row-level locks cannot be used in sqlite")
	ctx := context.Background()
	db := setupDB(t)
	depID := uuid.Must(uuid.NewV7())
	_, _ = db.CreateDepartment(ctx, depID, "Dep", "Dep")
	userID := uuid.Must(uuid.NewV7())
	db.c.User.Create().
		SetID(userID).
		SetFirstName("Original").
		SetLastName("User").
		SetDepartmentID(depID).
		SetRoleID(1).
		ExecX(ctx)

	t.Run("success", func(t *testing.T) {
		opts := sesc.UserUpdateOptions{
			FirstName:    "Updated",
			LastName:     "User",
			DepartmentID: depID,
			NewRoleID:    2,
		}

		user, err := db.UpdateUser(ctx, userID, opts)
		if err != nil {
			t.Fatalf("UpdateUser failed: %v", err)
		}

		if user.FirstName != opts.FirstName || user.Role.ID != opts.NewRoleID {
			t.Errorf("User not updated correctly: %v", user)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := db.UpdateUser(ctx, uuid.Must(uuid.NewV7()), sesc.UserUpdateOptions{})
		if !errors.Is(err, sesc.ErrUserNotFound) {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("invalid department", func(t *testing.T) {
		opts := sesc.UserUpdateOptions{DepartmentID: uuid.Must(uuid.NewV7())}
		_, err := db.UpdateUser(ctx, userID, opts)
		if !errors.Is(err, sesc.ErrInvalidDepartment) {
			t.Errorf("Expected ErrInvalidDepartment, got %v", err)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		opts := sesc.UserUpdateOptions{NewRoleID: 999}
		_, err := db.UpdateUser(ctx, userID, opts)
		if !errors.Is(err, sesc.ErrInvalidRole) {
			t.Errorf("Expected ErrInvalidRole, got %v", err)
		}
	})
}

func TestUserByID(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	userID := uuid.Must(uuid.NewV7())
	db.c.User.Create().
		SetID(userID).
		SetFirstName("John").
		SetLastName("Doe").
		SetRoleID(1).
		ExecX(ctx)

	t.Run("existing user", func(t *testing.T) {
		user, err := db.UserByID(ctx, userID)
		if err != nil {
			t.Fatalf("UserByID failed: %v", err)
		}
		if user.ID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, user.ID)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := db.UserByID(ctx, uuid.Must(uuid.NewV7()))
		if !errors.Is(err, sesc.ErrUserNotFound) {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestUsers(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	db.c.User.Create().
		SetID(uuid.Must(uuid.NewV7())).
		SetFirstName("User1").
		SetLastName("User1").
		SetRoleID(1).
		ExecX(ctx)
	db.c.User.Create().
		SetID(uuid.Must(uuid.NewV7())).
		SetFirstName("User2").
		SetLastName("User2").
		SetRoleID(1).
		ExecX(ctx)

	t.Run("fetch all users", func(t *testing.T) {
		users, err := db.Users(ctx)
		if err != nil {
			t.Fatalf("Users failed: %v", err)
		}
		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}
	})
}
