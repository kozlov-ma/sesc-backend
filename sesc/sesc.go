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

func (s *SESC) Init(ctx context.Context) error {
	errr := s.db.InsertDefaultPermissions(ctx, Permissions)
	errp := s.db.InsertDefaultRoles(ctx, []Role{
		Teacher,
		Dephead,
		ScientificDeputy,
		DevelopmentDeputy,
		ContestDeputy,
	})

	return errors.Join(errr, errp)
}

// Return a sesc.DepartmentAlreadyExists if the department already exists
func (s *SESC) CreateDepartment(ctx context.Context, name, description string) (Department, error) {
	id, err := uuid.NewV7()
	if err != nil {
		s.log.ErrorContext(ctx, "couldn't create uuid", slog.Any("error", err))
		return Department{}, fmt.Errorf("couldn't create uuid: %w", err)
	}

	d, err := s.db.CreateDepartment(ctx, id, name, description)
	if errors.Is(err, ErrInvalidDepartment) {
		s.log.DebugContext(ctx, "department already exists", slog.Any("department", id))
		return Department{}, err
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

	u := User{
		ID:         id,
		FirstName:  opt.FirstName,
		LastName:   opt.LastName,
		MiddleName: opt.MiddleName,
		PictureURL: opt.PictureURL,
		Role:       Teacher,
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

	u := User{
		ID:         id,
		PictureURL: opt.PictureURL,
		Role:       role,
		FirstName:  opt.FirstName,
		LastName:   opt.LastName,
		MiddleName: opt.MiddleName,
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
	if errors.Is(err, ErrUserNotFound) {
		s.log.DebugContext(ctx, "user id not found", slog.Any("id", id))
		return u, err
	} else if err != nil {
		s.log.ErrorContext(ctx, "couldn't get user because of db error", slog.Any("user_id", id), slog.Any("error", err))
		return u, fmt.Errorf("couldn't get user: %w", err)
	}

	s.log.DebugContext(ctx, "user id found", slog.Any("id", id))
	return u, nil
}

// GrantExtraPermissions grants extra permissions to a user.
// If the user does not exist, returns a sesc.ErrUserNotFound.
// If one of the permissions is invalid, returns a sesc.ErrInvalidPermission.
//
// THIS IS NOT THREAD SAFE FOR THE SAME USER.
func (s *SESC) GrantExtraPermissions(ctx context.Context, user User, permission ...Permission) (User, error) {
	u, err := s.db.GrantExtraPermissions(ctx, user, permission...)
	if errors.Is(err, ErrUserNotFound) {
		s.log.DebugContext(
			ctx,
			"tried to add extra permissions to a non-existent user",
			"user_id",
			user.ID,
			"permissions",
			permission,
		)
		return u, err
	} else if errors.Is(err, ErrInvalidPermission) {
		s.log.DebugContext(ctx, "tried to add an invalid permission to a non-existent user", "user_id", user.ID, "permissions", permission)
		return u, err
	} else if err != nil {
		s.log.ErrorContext(ctx, "couldn't grant extra permissions because of a db error", "user_id", user.ID, "permissions", permission, "error", err)
		return u, fmt.Errorf("couldn't grant extra permissions: %w", err)
	}

	s.log.InfoContext(ctx, "granted extra permissions", "user_id", user.ID, "permissions", permission)
	return u, nil
}

// RevokeExtraPermissions revokes extra permissions from a user.
// It does not affect the permissions that belong to the user's role.
// If the user does not exist, returns a sesc.ErrUserNotFound.
// If one of the permissions is invalid, or the User does not have it, returns a sesc.ErrInvalidPermission.
func (s *SESC) RevokeExtraPermissions(ctx context.Context, user User, permission ...Permission) (User, error) {
	u, err := s.db.RevokeExtraPermissions(ctx, user, permission...)
	if errors.Is(err, ErrUserNotFound) {
		s.log.DebugContext(
			ctx,
			"tried to revoke extra permissions from a non-existent user",
			"user_id",
			user.ID,
			"permissions",
			permission,
		)
		return u, ErrUserNotFound
	} else if errors.Is(err, ErrInvalidPermission) {
		s.log.DebugContext(ctx, "tried to revoke an invalid permission from a non-existent user", "user_id", user.ID, "permissions", permission)
		return u, ErrInvalidPermission
	} else if err != nil {
		s.log.ErrorContext(ctx, "couldn't revoke extra permissions because of a db error", "user_id", user.ID, "permissions", permission, "error", err)
		return u, fmt.Errorf("couldn't revoke extra permissions: %w", err)
	}

	s.log.InfoContext(ctx, "revoked extra permissions", "user_id", user.ID, "permissions", permission)
	return u, nil
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
	// Validate role against predefined roles
	validRoles := []Role{Teacher, Dephead, ContestDeputy, ScientificDeputy, DevelopmentDeputy}
	validRole := false
	for _, r := range validRoles {
		if role.ID == r.ID {
			validRole = true
			break
		}
	}
	if !validRole {
		s.log.DebugContext(ctx, "invalid role provided", slog.Any("role", role))
		return User{}, ErrInvalidRole
	}

	currentUser, err := s.User(ctx, user.ID)
	if err != nil {
		return User{}, err
	}

	// Check Teacher role requirements
	if role.ID == Teacher.ID && currentUser.Department == NoDepartment {
		s.log.DebugContext(ctx, "cannot assign Teacher role without department",
			slog.Any("user", currentUser.ID))
		return User{}, ErrInvalidRoleChange
	}

	// Check if role is unchanged
	if currentUser.Role.ID == role.ID {
		s.log.DebugContext(ctx, "role already set",
			slog.Any("user", currentUser.ID), slog.Any("role", role))
		return currentUser, nil
	}

	currentUser.Role = role
	if err := s.db.SaveUser(ctx, currentUser); err != nil {
		s.log.ErrorContext(ctx, "failed to save user role",
			slog.Any("error", err), slog.Any("user", currentUser.ID))
		return User{}, fmt.Errorf("failed to save user role: %w", err)
	}

	s.log.InfoContext(ctx, "user role updated",
		slog.Any("user", currentUser.ID), slog.Any("new_role", role))
	return currentUser, nil
}

// SetDepartment sets the department for Teacher or a Dephead,
// otherwise returns a sesc.ErrInvalidDepartment.
//
// If the user does not exist, returns a sesc.ErrUserNotFound.
// If the department is invalid, returns a sesc.ErrInvalidDepartment.
func (s *SESC) SetDepartment(ctx context.Context, user User, department Department) (User, error) {
	currentUser, err := s.User(ctx, user.ID)
	if err != nil {
		return User{}, err
	}

	// Validate allowed roles
	if currentUser.Role.ID != Teacher.ID && currentUser.Role.ID != Dephead.ID {
		s.log.DebugContext(ctx, "invalid role for department assignment",
			slog.Any("user", currentUser.ID), slog.Any("role", currentUser.Role))
		return User{}, ErrInvalidDepartment
	}

	// Check if department is unchanged
	if currentUser.Department.ID == department.ID {
		s.log.DebugContext(ctx, "department already set",
			slog.Any("user", currentUser.ID), slog.Any("department", department))
		return currentUser, nil
	}

	currentUser.Department = department
	if err := s.db.SaveUser(ctx, currentUser); err != nil {
		s.log.ErrorContext(ctx, "failed to set department",
			slog.Any("error", err), slog.Any("department", department.ID))
		return User{}, ErrInvalidDepartment
	}

	s.log.InfoContext(ctx, "department updated",
		slog.Any("user", currentUser.ID), slog.Any("department", department))
	return currentUser, nil
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

// Departments returns all the departments currently registered within the system.
func (s *SESC) Departments(ctx context.Context) ([]Department, error) {
	return s.db.Departments(ctx)
}

// Roles returns all the roles currently registered within the system.
//
// As it is not possible to register new roles at the moment, it returns all the pre-defined roles.
func (s *SESC) Roles(ctx context.Context) ([]Role, error) {
	return []Role{
		Teacher,
		Dephead,
		ContestDeputy,
		ScientificDeputy,
		DevelopmentDeputy,
	}, nil
}
