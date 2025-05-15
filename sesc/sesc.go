// Package sesc models the SESC employees and relationships between them.
package sesc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type UUID = uuid.UUID

// SESC represents the organization's structure and provides methods to interact with it.
type SESC struct {
	db DB
}

func New(db DB) *SESC {
	return &SESC{
		db: db,
	}
}

func (s *SESC) newUUID() (UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return id, fmt.Errorf("couldn't create UUID: %w", err)
	}

	return id, nil
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
	rec := event.Get(ctx).Sub("sesc/update_user")

	rec.Sub("params").Set(
		"id", id,
		"first_name", upd.FirstName,
		"last_name", upd.LastName,
		"middle_name", upd.MiddleName,
		"picture_url", upd.PictureURL,
		"suspended", upd.Suspended,
		"department_id", upd.DepartmentID,
		"new_role_id", upd.NewRoleID,
	)

	if upd.NewRoleID != 0 {
		_, ok := RoleByID(upd.NewRoleID)
		if !ok {
			return User{}, ErrInvalidRole
		}
	}

	if upd.FirstName == "" || upd.LastName == "" {
		return User{}, ErrInvalidName
	}

	statrec := event.Get(ctx).Sub("stats")
	dbStart := time.Now()
	updated, err := s.db.UpdateUser(ctx, id, upd)
	statrec.Add(events.PostgresTime, time.Since(dbStart))

	if errors.Is(err, ErrUserNotFound) {
		rec.Set("user_exists", false)
		return User{}, ErrUserNotFound
	} else if err != nil {
		err := fmt.Errorf("failed to update user: %w", err)
		rec.Add(events.Error, err)
		return User{}, err
	}

	rec.Set("success", true)
	rec.Set("user", updated.EventRecord())
	return updated, nil
}

// CreateUser creates a new User with a specified role.
//
// Returns an ErrInvalidName if the first or last name is missing.
func (s *SESC) CreateUser(ctx context.Context, opt UserUpdateOptions) (User, error) {
	rec := event.Get(ctx).Sub("sesc/create_user")

	rec.Sub("params").Set(
		"first_name", opt.FirstName,
		"last_name", opt.LastName,
		"middle_name", opt.MiddleName,
		"picture_url", opt.PictureURL,
		"suspended", opt.Suspended,
		"department_id", opt.DepartmentID,
		"new_role_id", opt.NewRoleID,
	)

	if err := opt.Validate(); err != nil {
		return User{}, err
	}

	statrec := event.Get(ctx).Sub("stats")
	dbStart := time.Now()
	u, err := s.db.SaveUser(ctx, opt)
	statrec.Add(events.PostgresTime, time.Since(dbStart))

	switch {
	case errors.Is(err, ErrInvalidDepartment):
		return User{}, ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't save user: %w", err)
		rec.Add(events.Error, err)
		return User{}, err
	}

	rec.Set("success", true)
	rec.Set("user", u.EventRecord())
	return u, nil
}
