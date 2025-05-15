package sesc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// In this file are SESC methods that contain no business logic and just forward the DB methods,
// optionally logging some conditions.
//
// These methods don't really need to be tested, though unit tests are still welcome.

// CreateDepartment creates a new Department with the given name and description.
//
// Return a sesc.DepartmentAlreadyExists if the department already exists.
func (s *SESC) CreateDepartment(ctx context.Context, name, description string) (Department, error) {
	rec := event.Get(ctx).Sub("sesc/create_department")

	rec.Sub("params").Set(
		"name", name,
		"description", description,
	)
	statrec := event.Get(ctx).Sub("stats")

	id, err := s.newUUID()
	if err != nil {
		rec.Add(events.Error, err)
		return Department{}, err
	}

	rec.Set("id", id)

	dbStart := time.Now()
	d, err := s.db.CreateDepartment(ctx, id, name, description)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	if errors.Is(err, ErrInvalidDepartment) {
		rec.Set("department_already_exists", true)
		return Department{}, err
	} else if err != nil {
		err := fmt.Errorf("couldn't save department: %w", err)
		rec.Add(events.Error, err)
		return Department{}, err
	}

	rec.Set("success", true)
	rec.Sub("department").Set(
		"id", d.ID,
		"name", d.Name,
		"description", d.Description,
	)
	return d, nil
}

func (s *SESC) UpdateDepartment(ctx context.Context, id UUID, name, description string) error {
	rec := event.Get(ctx).Sub("sesc/update_department")

	rec.Sub("params").Set(
		"id", id,
		"name", name,
		"description", description,
	)
	statrec := event.Get(ctx).Sub("stats")

	dbStart := time.Now()
	err := s.db.UpdateDepartment(ctx, id, name, description)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	if errors.Is(err, ErrInvalidDepartment) {
		rec.Add(events.Error, err)
		rec.Set("department", nil)
		return err
	} else if err != nil {
		err := fmt.Errorf("couldn't save department: %w", err)
		rec.Add(events.Error, err)
		return err
	}

	return nil
}

// User returns a User by ID. If the user does not exist, returns a sesc.ErrUserNotFound.
func (s *SESC) User(ctx context.Context, id UUID) (User, error) {
	rec := event.Get(ctx).Sub("sesc/user_by_id")

	rec.Sub("params").Set("id", id)
	statrec := event.Get(ctx).Sub("stats")
	rec.Set("user", nil)

	dbStart := time.Now()
	u, err := s.db.UserByID(ctx, id)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	if errors.Is(err, ErrUserNotFound) {
		return u, err
	} else if err != nil {
		rec.Add(events.Error, err)
		return u, fmt.Errorf("couldn't get user: %w", err)
	}

	rec.Set("user", u.EventRecord())
	return u, nil
}

// Departments returns all the departments currently registered within the system.
func (s *SESC) Departments(ctx context.Context) ([]Department, error) {
	rec := event.Get(ctx).Sub("sesc/departments")
	statrec := event.Get(ctx).Sub("stats")

	dbStart := time.Now()
	deps, err := s.db.Departments(ctx)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	rec.Set("count", len(deps))
	return deps, nil
}

func (s *SESC) DepartmentByID(ctx context.Context, id UUID) (Department, error) {
	rec := event.Get(ctx).Sub("sesc/department_by_id")

	rec.Sub("params").Set("id", id)
	statrec := event.Get(ctx).Sub("stats")

	dbStart := time.Now()
	dept, err := s.db.DepartmentByID(ctx, id)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	if err != nil {
		rec.Add(events.Error, err)
		if errors.Is(err, ErrInvalidDepartment) {
			rec.Set("department_exists", false)
		}
		return dept, err
	}

	rec.Set("department_exists", true)
	rec.Sub("department").Set(
		"id", dept.ID,
		"name", dept.Name,
		"description", dept.Description,
	)
	return dept, nil
}

func (s *SESC) DeleteDepartment(ctx context.Context, id UUID) error {
	rec := event.Get(ctx).Sub("sesc/delete_department")

	rec.Sub("params").Set("id", id)
	statrec := event.Get(ctx).Sub("stats")

	dbStart := time.Now()
	err := s.db.DeleteDepartment(ctx, id)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	switch {
	case errors.Is(err, ErrInvalidDepartment):
		rec.Set("department_exists", false)
		return err
	case errors.Is(err, ErrCannotRemoveDepartment):
		rec.Set("has_users", true)
		return err
	case err != nil:
		rec.Add(events.Error, err)
		return fmt.Errorf("couldn't delete department: %w", err)
	}

	rec.Set("success", true)
	return nil
}

func (s *SESC) UpdateProfilePicture(ctx context.Context, id UUID, pictureURL string) error {
	rec := event.Get(ctx).Sub("sesc/update_profile_picture")

	rec.Sub("params").Set(
		"id", id,
		"picture_url", pictureURL,
	)
	statrec := event.Get(ctx).Sub("stats")

	dbStart := time.Now()
	err := s.db.UpdateProfilePicture(ctx, id, pictureURL)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	if errors.Is(err, ErrUserNotFound) {
		rec.Set("user_exists", false)
		return err
	} else if err != nil {
		rec.Add(events.Error, err)
		return fmt.Errorf("couldn't update profile picture: %w", err)
	}

	rec.Set("success", true)
	return nil
}

func (s *SESC) Users(ctx context.Context) ([]User, error) {
	rec := event.Get(ctx).Sub("sesc/users")
	statrec := event.Get(ctx).Sub("stats")

	dbStart := time.Now()
	users, err := s.db.Users(ctx)
	statrec.Add(events.PostgresTime, time.Since(dbStart))
	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	rec.Set("count", len(users))
	return users, nil
}
