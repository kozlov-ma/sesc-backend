package api

import (
	"context"

	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	SESC interface {
		// UpdateUser updates user with the new fields.
		//
		// Returns an ErrInvalidRole if the new role id is invalid.
		// Returns an ErrInvalidName if the first or last name is missing.
		UpdateUser(ctx context.Context, id sesc.UUID, upd sesc.UserUpdateOptions) (sesc.User, error)
		// CreateUser creates a new User with a specified role.
		//
		// Returns an ErrInvalidName if the first or last name is missing.z
		CreateUser(ctx context.Context, opt sesc.UserOptions, role sesc.Role) (sesc.User, error)
		// Return a sesc.DepartmentAlreadyExists if the department already exists
		CreateDepartment(ctx context.Context, name, description string) (sesc.Department, error)
		UpdateDepartment(ctx context.Context, id sesc.UUID, name, description string) error
		// User returns a User by ID. If the user does not exist, returns a sesc.ErrUserNotFound.
		User(ctx context.Context, id sesc.UUID) (sesc.User, error)

		// Users returns all the users currently registered within the system.
		Users(ctx context.Context) ([]sesc.User, error)

		// Departments returns all the departments currently registered within the system.
		Departments(ctx context.Context) ([]sesc.Department, error)
		DepartmentByID(ctx context.Context, id sesc.UUID) (sesc.Department, error)
		DeleteDepartment(ctx context.Context, id sesc.UUID) error
		UpdateProfilePicture(ctx context.Context, id sesc.UUID, pictureURL string) error
	}
)
