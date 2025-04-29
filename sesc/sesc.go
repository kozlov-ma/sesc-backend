// Package sesc models the SESC employees and relationships between them.
package sesc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/auth"
	"github.com/kozlov-ma/sesc-backend/db"
)

type UUID = uuid.UUID

// SESC represents the organization's structure and provides methods to interact with it.
type SESC struct {
	log *slog.Logger
	db  DB
	iam IAM
	// Notes on the implementation.
	//
	// 1. Create the necessary interfaces in sesc/ports.go.
	// 2. Use the *slog.Logger.
	// 3. Add logging in every method.
}

func New(log *slog.Logger, db DB, iam IAM) *SESC {
	return &SESC{
		log: log,
		db:  db,
		iam: iam,
	}
}

// Return a sesc.DepartmentAlreadyExists if the department already exists
func (s *SESC) CreateDepartment(ctx context.Context, name, description string) (Department, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Department{}, fmt.Errorf("couldn't create uuid: %w", err)
	}

	d, err := s.db.CreateDepartment(ctx, id, name, description)
	if errors.Is(err, db.ErrAlreadyExists) {
		s.log.DebugContext(ctx, "department already exists", slog.Any("department", id))
		return Department{}, errors.Join(err, ErrDepartmentAlreadyExists)
	} else if err != nil {
		s.log.ErrorContext(ctx, "got a db error when saving department", slog.Any("error", err))
		return Department{}, fmt.Errorf("couldn't save department: %w", err)
	}

	s.log.InfoContext(ctx, "created department", slog.Any("department", d))
	return d, nil
}

type UserOptions struct {
	FirstName  string
	LastName   string
	MiddleName string
	PictureURL string

	AuthCredentials auth.Credentials
}

// CreateTeacher creates a new teacher.
//
// Returns sesc.ErrUsernameTaken if the username in AuthCredentials is already taken.
//
// TODO: think of a solution for multiple teachers with the same name.
func (s *SESC) CreateTeacher(ctx context.Context, opt UserOptions, department Department) (User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return User{}, fmt.Errorf("couldn't create uuid: %w", err)
	}

	aid, err := s.iam.Register(ctx, opt.AuthCredentials)
	if errors.Is(err, auth.ErrDuplicateUsername) {
		return User{}, errors.Join(err, ErrUsernameTaken)
	} else if err != nil {
		s.log.ErrorContext(ctx, "couldn't register user because of IAM error", slog.Any("error", err))
		return User{}, fmt.Errorf("couldn't register user: %w", err)
	}

	u := User{
		ID:         id,
		FirstName:  opt.FirstName,
		LastName:   opt.LastName,
		MiddleName: opt.MiddleName,
		PictureURL: opt.PictureURL,
		Role:       Teacher,
		AuthID:     aid,
		Department: department,
	}

	if err := s.db.SaveUser(ctx, u); err != nil {
		s.log.ErrorContext(ctx, "got a db error when saving user", slog.Any("error", err))
		return User{}, fmt.Errorf("couldn't save user: %w", err)
	}

	s.log.InfoContext(ctx, "created teacher", slog.Any("user", u))
	return u, nil
}

// CreateUser creates a new User with a specified role.
//
// To create a Teacher, use CreateTeacher. If role is Teacher, returns sesc.ErrInvalidRole.
func (s *SESC) CreateUser(ctx context.Context, opt UserOptions, role Role) (User, error) {
	if role.ID == Teacher.ID {
		s.log.DebugContext(ctx, "tried to create a user with Teacher role", slog.Any("user", opt))
		return User{}, ErrInvalidRole
	}

	id, err := uuid.NewV7()
	if err != nil {
		s.log.ErrorContext(ctx, "couldn't create uuid", slog.Any("error", err))
		return User{}, fmt.Errorf("couldn't create user ID: %w", err)
	}

	aid, err := s.iam.Register(ctx, opt.AuthCredentials)
	if errors.Is(err, auth.ErrDuplicateUsername) {
		return User{}, errors.Join(err, ErrUsernameTaken)
	} else if err != nil {
		s.log.ErrorContext(ctx, "couldn't register user because of IAM error", slog.Any("error", err))
		return User{}, fmt.Errorf("couldn't register user: %w", err)
	}

	u := User{
		ID:         id,
		PictureURL: opt.PictureURL,
		Role:       role,
		FirstName:  opt.FirstName,
		LastName:   opt.LastName,
		MiddleName: opt.MiddleName,
		AuthID:     aid,
	}

	if err := s.db.SaveUser(ctx, u); err != nil {
		s.log.ErrorContext(ctx, "got a db error when saving user", slog.Any("error", err))
		return User{}, fmt.Errorf("couldn't save user: %w", err)
	}

	s.log.InfoContext(ctx, "created user", slog.Any("user", u))
	return u, nil
}

// User returns a User by ID. If the user does not exist, returns a sesc.ErrUserNotFound.
func (s *SESC) User(ctx context.Context, id UUID) (User, error) {
	u, err := s.db.UserByID(ctx, id)
	if errors.Is(err, db.ErrUserNotFound) {
		s.log.DebugContext(ctx, "user id not found", slog.Any("id", id))
		return u, errors.Join(err, ErrUserNotFound)
	}

	s.log.DebugContext(ctx, "user id found", slog.Any("id", id))
	return u, nil
}

// GrantExtraPermissions grants extra permissions to a user.
// If the user does not exist, returns a sesc.ErrUserNotFound.
// If one of the permissions is invalid, returns a sesc.ErrInvalidPermission.
func (s *SESC) GrantExtraPermissions(ctx context.Context, user User, permission ...Permission) (User, error) {
	panic("not implemented")
}

// RevokeExtraPermissions revokes extra permissions from a user.
// It does not affect the permissions that belong to the user's role.
// If the user does not exist, returns a sesc.ErrUserNotFound.
// If one of the permissions is invalid, or the User does not have it, returns a sesc.ErrInvalidPermission.
func (s *SESC) RevokeExtraPermissions(ctx context.Context, user User, permission ...Permission) (User, error) {
	panic("not implemented")
}

// SetRole sets the role of a user.
//
// It does nothing if the user's role is already the same as the new role.
//
// To change a user's role to a teacher, first assign it to a department.
//
// If the user does not have a department and role is Teacher, returns a sesc.ErrInvalidRoleChange.
//
// If the user does not exist, returns a sesc.ErrUserNotFound.
// If the role is invalid, returns a sesc.ErrInvalidRole.
func (s *SESC) SetRole(ctx context.Context, user User, role Role) (User, error) {
	panic("not implemented")
}

// SetDepartment sets the department for Teacher or a Dephead,
// otherwise returns a sesc.ErrInvalidDepartment.
//
// If the user does not exist, returns a sesc.ErrUserNotFound.
// If the department is invalid, returns a sesc.ErrInvalidDepartment.
func (s *SESC) SetDepartment(ctx context.Context, user User, department Department) (User, error) {
	panic("not implemented")
}

// SetUserInfo changes the user's options.
//
// If the user does not exist, returns a sesc.ErrUserNotFound.
func (s *SESC) SetUserInfo(ctx context.Context, user User, opt UserOptions) (User, error) {
	panic("not implemented")
}

// SetProfilePic sets the profile picture for a user.
// This method is an addition to SetUserInfo, because the SetUserInfo is supposed to only
// be used by the system administrator, while SetProfilePic should be available to users.
//
// If the user does not exist, returns a sesc.ErrUserNotFound.
func (s *SESC) SetProfilePic(ctx context.Context, user User, pictureURL string) (User, error) {
	panic("not implemented")
}
