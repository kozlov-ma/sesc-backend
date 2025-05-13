// Package sesc models the SESC employees and relationships between them.
package sesc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
)

type UUID = uuid.UUID

// SESC represents the organization's structure and provides methods to interact with it.
type SESC struct {
	log *slog.Logger
	db  DB
}

func New(log *slog.Logger, db DB) *SESC {
	return &SESC{
		log: log,
		db:  db,
	}
}

func (s *SESC) newUUID() (UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		s.log.Error("couldn't create uuid", slog.Any("error", err))
		return id, fmt.Errorf("couldn't create UUID: %w", err)
	}

	return id, err
}

// UserUpdateOptions represents the options for updating a user.
type UserUpdateOptions struct {
	FirstName    string
	LastName     string
	MiddleName   string
	PictureURL   string
	Suspended    bool
	DepartmentID UUID
	NewRoleID    int32
}

func (u UserUpdateOptions) Validate() error {
	if u.FirstName == "" || u.LastName == "" {
		return ErrInvalidName
	}

	if _, ok := RoleByID(u.NewRoleID); !ok {
		return ErrInvalidRole
	}

	return nil
}

// UpdateUser updates user with the new fields.
//
// Returns an ErrInvalidRole if the new role id is invalid.
// Returns an ErrInvalidName if the first or last name is missing.
func (s *SESC) UpdateUser(ctx context.Context, id UUID, upd UserUpdateOptions) (User, error) {
	if upd.NewRoleID != 0 {
		_, ok := RoleByID(upd.NewRoleID)
		if !ok {
			return User{}, ErrInvalidRole
		}
	}

	if upd.FirstName == "" || upd.LastName == "" {
		return User{}, ErrInvalidName
	}

	updated, err := s.db.UpdateUser(ctx, id, upd)
	if errors.Is(err, ErrUserNotFound) {
		s.log.DebugContext(ctx, "tried to update user that does not exist", slog.Any("id", id))
		return User{}, ErrUserNotFound
	} else if err != nil {
		s.log.ErrorContext(ctx, "failed to update user because of a db error", slog.Any("id", id), slog.Any("updates", upd), slog.Any("error", err))
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}

	s.log.InfoContext(ctx, "updated user", slog.Any("id", id), slog.Any("updates", upd))
	return updated, nil
}

// CreateUser creates a new User with a specified role.
//
// Returns an ErrInvalidName if the first or last name is missing.z
func (s *SESC) CreateUser(ctx context.Context, opt UserUpdateOptions) (User, error) {
	if err := opt.Validate(); err != nil {
		return User{}, err
	}
	u, err := s.db.SaveUser(ctx, opt)
	switch {
	case errors.Is(err, ErrInvalidDepartment):
		return User{}, ErrInvalidDepartment
	case err != nil:
		s.log.ErrorContext(ctx, "got a db error when saving user", slog.Any("error", err))
		return User{}, fmt.Errorf("couldn't save user: %w", err)
	}

	s.log.InfoContext(ctx, "created user", slog.Any("user", u))
	return u, nil
}
