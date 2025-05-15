// Package sesc models the SESC employees and relationships between them.
package sesc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type UUID = uuid.UUID

// SESC represents the organization's structure and provides methods to interact with it.
type SESC struct {
	client *ent.Client
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %w", err, rerr)
	}
	return err
}

func convertUser(u *ent.User) (User, error) {
	var dept Department
	dep := u.Edges.Department
	if dep != nil {
		dept = Department{
			ID:          dep.ID,
			Name:        dep.Name,
			Description: dep.Description,
		}
	}

	role, ok := RoleByID(u.RoleID)
	if !ok {
		return User{}, ErrInvalidRole
	}

	return User{
		ID:         u.ID,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		PictureURL: u.PictureURL,
		Suspended:  u.Suspended,
		Department: dept,
		Role:       role,
	}, nil
}

func New(client *ent.Client) *SESC {
	return &SESC{
		client: client,
	}
}

// CreateDepartment creates a new department with auto-generated ID.
// Returns an ErrInvalidDepartment if department already exists.
func (s *SESC) CreateDepartment(
	ctx context.Context,
	name string,
	description string,
) (Department, error) {
	rec := event.Get(ctx).Sub("sesc/create_department")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"name", name,
		"description", description,
	)

	id, err := s.newUUID()
	if err != nil {
		rec.Add(events.Error, err)
		return NoDepartment, err
	}

	rec.Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := s.client.Department.Create().
		SetID(id).
		SetName(name).
		SetDescription(description).
		Save(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsConstraintError(err):
		return NoDepartment, ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't save department: %w", err)
		rec.Add(events.Error, err)
		return NoDepartment, err
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

// DepartmentByID retrieves a department by ID.
// Returns an ErrInvalidDepartment if the department does not exist.
func (s *SESC) DepartmentByID(ctx context.Context, id UUID) (Department, error) {
	rec := event.Get(ctx).Sub("sesc/department_by_id")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := s.client.Department.Get(ctx, id)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		return NoDepartment, ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't get department: %w", err)
		rec.Add(events.Error, err)
		return NoDepartment, err
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
func (s *SESC) Departments(ctx context.Context) ([]Department, error) {
	rec := event.Get(ctx).Sub("sesc/departments")
	statrec := event.Get(ctx).Sub("stats")

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
// Returns an ErrInvalidDepartment if the department does not exist.
func (s *SESC) UpdateDepartment(
	ctx context.Context,
	id UUID,
	name string,
	description string,
) error {
	rec := event.Get(ctx).Sub("sesc/update_department")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"name", name,
		"description", description,
	)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := s.client.Department.UpdateOneID(id).SetName(name).SetDescription(description).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		joinedErr := fmt.Errorf("%w: %w", err, ErrInvalidDepartment)
		rec.Add(events.Error, joinedErr)
		return joinedErr
	case err != nil:
		err := fmt.Errorf("couldn't update department: %w", err)
		rec.Add(events.Error, err)
		return err
	}

	rec.Set("success", true)
	return nil
}

// DeleteDepartment deletes a department.
// Returns an ErrInvalidDepartment if the department does not exist.
// Returns an ErrCannotRemoveDepartment if the department is assigned to a user.
func (s *SESC) DeleteDepartment(ctx context.Context, id UUID) error {
	rec := event.Get(ctx).Sub("sesc/delete_department")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := s.client.Department.DeleteOneID(id).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsConstraintError(err):
		return ErrCannotRemoveDepartment
	case ent.IsNotFound(err):
		return ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't delete department: %w", err)
		rec.Add(events.Error, err)
		return err
	}

	rec.Set("success", true)
	return nil
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
// Returns an ErrUserNotFound if the user does not exist.
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

	// Validate the user exists first before other validations
	_, err := s.UserByID(ctx, id)
	if err != nil {
		return User{}, err
	}

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
	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)

	txStart := time.Now()
	tx, err := s.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		err := fmt.Errorf("couldn't start transaction: %w", err)
		txrec.Add(events.Error, err)
		return User{}, err
	}

	var dept *ent.Department
	if upd.DepartmentID != uuid.Nil {
		statrec.Add(events.PostgresQueries, 1)
		dept, err = tx.Department.Get(ctx, upd.DepartmentID)
		switch {
		case ent.IsNotFound(err):
			txrec.Add(events.Error, ErrInvalidDepartment)
			return User{}, rollback(tx, ErrInvalidDepartment)
		case err != nil:
			err := fmt.Errorf("couldn't query department: %w", err)
			txrec.Add(events.Error, err)
			return User{}, rollback(tx, err)
		}
	}

	statrec.Add(events.PostgresQueries, 1)
	updater := tx.User.UpdateOneID(id).
		SetFirstName(upd.FirstName).
		SetLastName(upd.LastName).
		SetMiddleName(upd.MiddleName).
		SetPictureURL(upd.PictureURL).
		SetSuspended(upd.Suspended).
		SetRoleID(upd.NewRoleID)

	if dept != nil {
		updater = updater.SetDepartmentID(dept.ID)
	} else {
		updater = updater.ClearDepartment()
	}

	_, err = updater.Save(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't update user: %w", err)
		txrec.Add(events.Error, err)
		return User{}, rollback(tx, err)
	}

	statrec.Add(events.PostgresQueries, 1)
	us, err := tx.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't query user after an update: %w", err)
		txrec.Add(events.Error, err)
		return User{}, rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return User{}, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	updated, err := convertUser(us)
	if err != nil {
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
	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)

	txStart := time.Now()
	tx, err := s.client.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		err := fmt.Errorf("couldn't begin transaction: %w", err)
		txrec.Add(events.Error, err)
		return User{}, err
	}

	statrec.Add(events.PostgresQueries, 1)
	var dept *ent.Department
	if opt.DepartmentID != uuid.Nil {
		dept, err = tx.Department.Get(ctx, opt.DepartmentID)
		switch {
		case ent.IsNotFound(err):
			return User{}, rollback(tx, ErrInvalidDepartment)
		case err != nil:
			err := fmt.Errorf("couldn't query department: %w", err)
			txrec.Add(events.Error, err)
			return User{}, rollback(tx, err)
		}
	}

	statrec.Add(events.PostgresQueries, 1)
	cr := tx.User.Create().
		SetFirstName(opt.FirstName).
		SetLastName(opt.LastName).
		SetMiddleName(opt.MiddleName).
		SetPictureURL(opt.PictureURL).
		SetRoleID(opt.NewRoleID)
	if dept != nil {
		cr = cr.SetDepartment(dept)
	}

	res, err := cr.Save(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't save user: %w", err)
		txrec.Add(events.Error, err)
		return User{}, rollback(tx, err)
	}

	statrec.Add(events.PostgresQueries, 1)
	us, err := tx.User.Query().Where(user.ID(res.ID)).WithDepartment().Only(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't query user after saving them: %w", err)
		txrec.Add(events.Error, err)
		return User{}, rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return User{}, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	u, err := convertUser(us)
	if err != nil {
		rec.Add(events.Error, err)
		return User{}, err
	}

	rec.Set("success", true)
	rec.Set("user", u.EventRecord())
	return u, nil
}

// UpdateProfilePicture updates a user's profile picture.
// Returns an ErrUserNotFound if the user does not exist.
func (s *SESC) UpdateProfilePicture(ctx context.Context, id UUID, pictureURL string) error {
	rec := event.Get(ctx).Sub("sesc/update_profile_picture")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"picture_url", pictureURL,
	)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := s.client.User.UpdateOneID(id).SetPictureURL(pictureURL).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		joinedErr := fmt.Errorf("%w: %w", err, ErrUserNotFound)
		rec.Add(events.Error, joinedErr)
		return joinedErr
	case err != nil:
		err := fmt.Errorf("couldn't update user: %w", err)
		rec.Add(events.Error, err)
		return err
	}

	rec.Set("success", true)
	return nil
}

// UserByID gets a user by their ID.
// Returns an ErrUserNotFound if the user does not exist.
func (s *SESC) UserByID(ctx context.Context, id UUID) (User, error) {
	rec := event.Get(ctx).Sub("sesc/user_by_id")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	u, err := s.client.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		rec.Add(events.Error, ErrUserNotFound)
		return User{}, ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't query user: %w", err)
		rec.Add(events.Error, err)
		return User{}, err
	}

	userObj, err := convertUser(u)
	if err != nil {
		rec.Add(events.Error, err)
		return User{}, err
	}

	rec.Set("user", userObj.EventRecord())
	return userObj, nil
}

// Users gets all users.
func (s *SESC) Users(ctx context.Context) ([]User, error) {
	rec := event.Get(ctx).Sub("sesc/users")
	statrec := event.Get(ctx).Sub("stats")

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := s.client.User.Query().WithDepartment().All(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	if err != nil {
		err := fmt.Errorf("couldn't query users: %w", err)
		rec.Add(events.Error, err)
		return nil, err
	}

	users := make([]User, len(res))
	for i, r := range res {
		users[i], err = convertUser(r)
		if err != nil {
			return nil, fmt.Errorf("couldn't convert user: %w", err)
		}
	}

	return users, nil
}

// User returns a User by ID. Alias for UserByID.
// Returns ErrUserNotFound if the user does not exist.
func (s *SESC) User(ctx context.Context, id UUID) (User, error) {
	return s.UserByID(ctx, id)
}
