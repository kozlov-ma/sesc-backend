package sesc

import (
	"context"

	"github.com/kozlov-ma/sesc-backend/auth"
)

type (
	DB interface {
		SaveUser(context.Context, User) error

		// If the user already has the extra permission, or it is granted by a role, it is a no-op.
		// If the user does not exist, it returns a db.ErrUserNotFound.
		// If the permission is not valid, it returns a db.ErrInvalidPermission.
		GrantExtraPermissions(context.Context, User, ...Permission) (User, error)

		// If the user does not have the permission, it is a no-op.
		// If the user does not exist, it returns a db.ErrUserNotFound.
		// If the permission is not valid, it returns a db.ErrInvalidPermission.
		//
		// Permissions granted by roles are not affected by this operation.
		RevokeExtraPermissions(context.Context, User, ...Permission) (User, error)

		// Return a db.ErrUserNotFound if user does not exist.
		UserByID(context.Context, UUID) (User, error)

		// Return a db.ErrAlreadyExists if department already exists.
		CreateDepartment(ctx context.Context, id UUID, name, description string) (Department, error)

		// Return a db.ErrUserNotFound if department or user does not exist.
		AssignHeadOfDepartment(ctx context.Context, departmentID UUID, userID UUID) error
	}

	IAM interface {
		// Returns auth.DuplicateUsername if username is already taken.
		Register(context.Context, auth.Credentials) (auth.ID, error)
	}
)
