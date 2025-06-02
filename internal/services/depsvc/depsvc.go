package depsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	UUID                             = uuid.UUID
	User                             = sesc.User
	Department                       = sesc.Department
	Role                             = sesc.Role
	UserUpdateOptions                = sesc.UserUpdateOptions
	AchievementGroup                 = achievement.Group
	AchievementTemplate              = achievement.Template
	AchievementGroupCreateOptions    = achievement.GroupCreateOptions
	AchievementGroupUpdateOptions    = achievement.GroupUpdateOptions
	AchievementGroupSearchOptions    = achievement.GroupSearchOptions
	AchievementTemplateCreateOptions = achievement.TemplateCreateOptions
	AchievementTemplateUpdateOptions = achievement.TemplateUpdateOptions
	AchievementTemplateSearchOptions = achievement.TemplateSearchOptions
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
) (Department, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/create_department")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Sub("params").Set(
		"name", name,
		"description", description,
	)

	ctx = rec.Sub("create_department_record").Wrap(ctx)
	department, err := s.createDepartmentRecord(ctx, statrec, name, description)
	if ent.IsValidationError(err) {
		return Department{}, sesc.ErrInvalidDepartmentName
	}
	if err != nil {
		return Department{}, err
	}

	return department, nil
}

// createDepartmentRecord creates a department record in the database
func (s *DES) createDepartmentRecord(
	ctx context.Context,
	statrec *event.Record,
	name string,
	description string,
) (Department, error) {
	rec := event.Get(ctx)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := s.client.Department.Create().
		SetName(name).
		SetDescription(description).
		Save(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsConstraintError(err):
		rec.Set("success", false)
		rec.Add(events.Error, sesc.ErrInvalidDepartment)
		return Department{}, sesc.ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't save department: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return Department{}, err
	}

	rec.Set("success", true)
	rec.Set(
		"id", res.ID,
		"name", res.Name,
		"description", res.Description,
	)

	return Department{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
	}, nil
}

// DepartmentByID retrieves a department by ID.
// Returns an sesc.ErrInvalidDepartment if the department does not exist.
func (s *DES) DepartmentByID(ctx context.Context, id UUID) (Department, error) {
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
		return Department{}, sesc.ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't get department: %w", err)
		rec.Add(events.Error, err)
		return Department{}, err
	}

	rec.Sub("department").Set(
		"id", res.ID,
		"name", res.Name,
		"description", res.Description,
	)

	return Department{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
	}, nil
}

// Departments retrieves all departments.
func (s *DES) Departments(ctx context.Context) ([]Department, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/departments")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := s.client.Department.Query().All(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	if err != nil {
		err := fmt.Errorf("couldn't get all departments: %w", err)
		rec.Add(events.Error, err)
		return nil, err
	}

	deps := make([]Department, len(res))
	for i, r := range res {
		deps[i] = Department{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		}
	}

	return deps, nil
}

// UpdateDepartment updates a department.
// Returns an sesc.ErrInvalidDepartment if the department does not exist.
func (s *DES) UpdateDepartment(
	ctx context.Context,
	id UUID,
	name string,
	description string,
) error {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/update_department")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"name", name,
		"description", description,
	)

	// Stage 1: Update department record
	ctx = rec.Sub("update_department_record").Wrap(ctx)
	if err := s.updateDepartmentRecord(ctx, statrec, id, name, description); err != nil {
		return err
	}

	rec.Set("success", true)
	return nil
}

// updateDepartmentRecord updates a department record in the database
func (s *DES) updateDepartmentRecord(
	ctx context.Context,
	statrec *event.Record,
	id UUID,
	name string,
	description string,
) error {
	rec := event.Get(ctx)
	rec.Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := s.client.Department.UpdateOneID(id).SetName(name).SetDescription(description).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		joinedErr := fmt.Errorf("%w: %w", err, sesc.ErrInvalidDepartment)
		rec.Add(events.Error, joinedErr)
		rec.Set("success", false)
		return joinedErr
	case err != nil:
		err := fmt.Errorf("couldn't update department: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	return nil
}

// DeleteDepartment deletes a department by ID.
// Returns an sesc.ErrInvalidDepartment if the department does not exist.
// Returns an sesc.ErrCannotRemoveDepartment if the department has users.
func (s *DES) DeleteDepartment(ctx context.Context, id UUID) error {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/delete_department")

	rec.Sub("params").Set("id", id)

	// Stage 1: Delete department record
	ctx = rec.Sub("delete_department_record").Wrap(ctx)
	if err := s.deleteDepartmentRecord(ctx, id); err != nil {
		return err
	}

	rec.Set("success", true)
	return nil
}

// deleteDepartmentRecord deletes a department record from the database
func (s *DES) deleteDepartmentRecord(ctx context.Context, id UUID) error {
	rec := event.Get(ctx)
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := s.client.Department.DeleteOneID(id).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsConstraintError(err):
		rec.Add(events.Error, sesc.ErrCannotRemoveDepartment)
		rec.Set("success", false)
		return sesc.ErrCannotRemoveDepartment
	case ent.IsNotFound(err):
		rec.Add(events.Error, sesc.ErrInvalidDepartment)
		rec.Set("success", false)
		return sesc.ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't delete department: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	return nil
}
