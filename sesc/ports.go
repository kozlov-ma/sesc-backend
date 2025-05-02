package sesc

import (
	"context"
)

type (
	DB interface {
		SaveUser(context.Context, User) error

		// If the user already has the extra permission, or it is granted by a role, it is a no-op.
		// If the user does not exist, it returns an ErrUserNotFound.
		// If the permission is not valid, it returns an ErrInvalidPermission.
		GrantExtraPermissions(context.Context, User, ...Permission) (User, error)

		// If the user does not have the permission, it is a no-op.
		// If the user does not exist, it returns an ErrUserNotFound.
		// If the permission is not valid, it returns an ErrInvalidPermission.
		//
		// Permissions granted by roles are not affected by this operation.
		RevokeExtraPermissions(context.Context, User, ...Permission) (User, error)

		// Returns an ErrUserNotFound if user does not exist.
		UserByID(context.Context, UUID) (User, error)

		// Returns an ErrInvalidDepartment if department already exists.
		CreateDepartment(ctx context.Context, id UUID, name, description string) (Department, error)

		// Assigns a head to an existing department
		// Returns an ErrUserNotFound if user does not exist.
		// Returns an ErrInvalidDepartment if department does not exist.
		AssignHeadOfDepartment(ctx context.Context, departmentID UUID, userID UUID) error

		// Returns all currenlty registered departments.
		Departments(ctx context.Context) ([]Department, error)

		// Returns ErrInvalidDepartment if not exist
		DepartmentByID(ctx context.Context, id UUID) (Department, error)

		InsertDefaultPermissions(ctx context.Context, permissions []Permission) error

		InsertDefaultRoles(ctx context.Context, roles []Role) error
	}
)
