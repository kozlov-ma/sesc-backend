package sesc

import "context"

type (
	DB interface {
		SaveUser(ctx context.Context, user User) error

		// Return a db.ErrNotFound if user does not exist.
		UserByID(ctx context.Context, id UUID) (User, error)
	}
)
