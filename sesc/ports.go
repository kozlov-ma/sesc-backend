package sesc

import (
	"context"

	"github.com/kozlov-ma/sesc-backend/auth"
)

type (
	DB interface {
		SaveUser(context.Context, User) error

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

		UserByID(ctx context.Context, id UUID) (User, error)
	}
)
