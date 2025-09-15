package depsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type DES struct {
	client *ent.Client
}

func New(client *ent.Client) *DES {
	return &DES{
		client: client,
	}
}

// CreateDepartment creates a new department with auto-generated ID.
// Returns an sesc.ErrInvalidDepartment if department already exists.
func (s *DES) CreateDepartment(
	ctx context.Context,
	name string,
	description string,
) (*ent.Department, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/create_department")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Sub("params").Set(
		"name", name,
		"description", description,
	)

	ctx = rec.Sub("create_department_record").Wrap(ctx)

	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()
	dep, err := s.client.Department.Create().
		SetName(name).
		SetDescription(description).
		Save(ctx)
	statrec.Add(events.PostgresTime, time.Since(stime))

	if ent.IsValidationError(err) || ent.IsConstraintError(err) {
		return nil, sesc.ErrInvalidDepartmentName
	}
	if err != nil {
		return nil, fmt.Errorf("couldn't create department: %w", err)
	}

	return dep, nil
}

// DepartmentByID retrieves a department by ID.
// Returns an sesc.ErrInvalidDepartment if the department does not exist.
func (s *DES) DepartmentByID(ctx context.Context, id uuid.UUID) (*ent.Department, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/department_by_id")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Sub("params").Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := s.client.Department.Get(ctx, id)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		return nil, sesc.ErrDepartmentNotFound
	case err != nil:
		err := fmt.Errorf("couldn't get department: %w", err)
		rec.Add(events.Error, err)
		return nil, err
	}

	return res, nil
}

// Departments retrieves all departments.
func (s *DES) Departments(ctx context.Context) (ent.Departments, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/departments")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	deps, err := s.client.Department.Query().All(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	if err != nil {
		err := fmt.Errorf("couldn't get all departments: %w", err)
		rec.Add(events.Error, err)
		return nil, err
	}

	return deps, nil
}

// UpdateDepartment updates a department.
// Returns an sesc.ErrInvalidDepartment if the department does not exist.
func (s *DES) UpdateDepartment(
	ctx context.Context,
	id uuid.UUID,
	name string,
	description string,
) (*ent.Department, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/update_department")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"name", name,
		"description", description,
	)

	var dep *ent.Department

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	_, err := s.client.Department.UpdateOneID(id).SetName(name).SetDescription(description).Save(ctx)
	if ent.IsNotFound(err) {
		rec.Set("success", false)
		rec.Add(events.Error, err)
		return nil, fmt.Errorf("couldn't save department: %w", sesc.ErrDepartmentNotFound)
	}
	if err != nil {
		rec.Set("success", false)
		rec.Add(events.Error, err)
		return nil, fmt.Errorf("couldn't save department: %w", err)
	}

	statrec.Add(events.PostgresTime, time.Since(startTime))
	rec.Set("success", true)
	return dep, nil
}

// DeleteDepartment deletes a department by ID.
// Returns an sesc.ErrInvalidDepartment if the department does not exist.
// Returns an sesc.ErrCannotRemoveDepartment if the department has users.
func (s *DES) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/delete_department")

	rec.Sub("params").Set("id", id)
	statsRec := event.Root(ctx).Sub("stats")

	startTime := time.Now()

	statsRec.Add(events.PostgresQueries, 1)
	err := s.client.Department.DeleteOneID(id).Exec(ctx)

	statsRec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsConstraintError(err):
		rec.Add(events.Error, sesc.ErrCannotRemoveDepartment)
		rec.Set("success", false)
		return sesc.ErrCannotRemoveDepartment
	case ent.IsNotFound(err):
		rec.Add(events.Error, sesc.ErrDepartmentNotFound)
		rec.Set("success", false)
		return sesc.ErrDepartmentNotFound
	case err != nil:
		err := fmt.Errorf("couldn't delete department: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	return nil
}
