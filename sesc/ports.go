package sesc

import (
	"context"
)

type (
	DB interface {
		// SaveUser saves a new user to the database.
		SaveUser(context.Context, User) error

		// UpdateUser updates an existing user in the database.
		// Returns an ErrUserNotFound if the user does not exist.
		UpdateUser(context.Context, UUID, UserUpdateOptions) (User, error)

		// UpdateProfilePicture updates the profile picture of a user in the database.
		// Returns an ErrUserNotFound if the user does not exist.
		UpdateProfilePicture(context.Context, UUID, string) error

		// Returns an ErrUserNotFound if user does not exist.
		UserByID(context.Context, UUID) (User, error)

		Users(context.Context) ([]User, error)

		// Returns an ErrInvalidDepartment if department already exists.
		CreateDepartment(ctx context.Context, id UUID, name, description string) (Department, error)

		// Modifies the Department. Returns an ErrInvalidDepartment if the department at this id does not exist.
		UpdateDepartment(ctx context.Context, id UUID, name, description string) error

		// Deletes the Department. Returns an ErrInvalidDepartment if the department at this id does not exist.
		// Returns an ErrCannotRemoveDepartment if the department is currently assigned to a user.
		DeleteDepartment(ctx context.Context, id UUID) error

		// Returns all currenlty registered departments.
		Departments(ctx context.Context) ([]Department, error)

		// Returns ErrInvalidDepartment if not exist
		DepartmentByID(ctx context.Context, id UUID) (Department, error)
	}
)
