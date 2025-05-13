package sesc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// In this file are SESC methods that contain no business logic and just forward the DB methods,
// optionally logging some conditions.
//
// These methods don't really need to be tested, though unit tests are still welcome.

// CreateDepartment creates a new Department with the given name and description.
//
// Return a sesc.DepartmentAlreadyExists if the department already exists.
func (s *SESC) CreateDepartment(ctx context.Context, name, description string) (Department, error) {
	id, err := s.newUUID()
	if err != nil {
		return Department{}, err
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

func (s *SESC) UpdateDepartment(ctx context.Context, id UUID, name, description string) error {
	err := s.db.UpdateDepartment(ctx, id, name, description)
	if errors.Is(err, ErrInvalidDepartment) {
		s.log.DebugContext(ctx, "department already exists", slog.Any("department", id))
		return err
	} else if err != nil {
		s.log.ErrorContext(ctx, "got a db error when saving department", slog.Any("error", err))
		return fmt.Errorf("couldn't save department: %w", err)
	}

	s.log.InfoContext(ctx, "changed department", slog.Any("department", id))
	return nil
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

// Departments returns all the departments currently registered within the system.
func (s *SESC) Departments(ctx context.Context) ([]Department, error) {
	return s.db.Departments(ctx)
}

func (s *SESC) DepartmentByID(ctx context.Context, id UUID) (Department, error) {
	return s.db.DepartmentByID(ctx, id)
}

func (s *SESC) DeleteDepartment(ctx context.Context, id UUID) error {
	err := s.db.DeleteDepartment(ctx, id)
	switch {
	case errors.Is(err, ErrInvalidDepartment):
		s.log.DebugContext(ctx, "department id not found", slog.Any("id", id))
		return err
	case errors.Is(err, ErrCannotRemoveDepartment):
		s.log.InfoContext(ctx, "tried to delete department which still has users", slog.Any("department_id", id))
		return err
	case err != nil:
		s.log.ErrorContext(
			ctx,
			"couldn't delete department because of db error",
			slog.Any("department_id", id),
			slog.Any("error", err),
		)
		return fmt.Errorf("couldn't delete department: %w", err)
	}

	s.log.InfoContext(ctx, "deleted department", slog.Any("department", id))
	return nil
}

func (s *SESC) UpdateProfilePicture(ctx context.Context, id UUID, pictureURL string) error {
	err := s.db.UpdateProfilePicture(ctx, id, pictureURL)
	if errors.Is(err, ErrUserNotFound) {
		s.log.DebugContext(ctx, "user id not found", slog.Any("id", id))
		return err
	} else if err != nil {
		s.log.ErrorContext(ctx, "couldn't update profile picture because of db error", slog.Any("user_id", id), slog.Any("error", err))
		return fmt.Errorf("couldn't update profile picture: %w", err)
	}

	s.log.InfoContext(ctx, "updated profile picture", slog.Any("user", id))
	return nil
}

func (s *SESC) Users(ctx context.Context) ([]User, error) {
	return s.db.Users(ctx)
}
