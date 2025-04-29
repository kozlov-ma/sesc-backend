package sesc

import (
	"context"

	"github.com/kozlov-ma/sesc-backend/auth"
)

type (
	DB interface {
		SaveUser(context.Context, User) error

		// Return a db.ErrNotFound if user does not exist.
		UserByID(context.Context, UUID) (User, error)
	}

	IAM interface {
		// Returns auth.DuplicateUsername if username is already taken.
		Register(context.Context, auth.Credentials) (auth.ID, error)
	}
)
