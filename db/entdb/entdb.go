package entdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type DB struct {
	c *ent.Client
}

func New(c *ent.Client) *DB {
	return &DB{
		c: c,
	}
}

// CreateDepartment implements sesc.DB.
func (d *DB) CreateDepartment(
	ctx context.Context,
	id sesc.UUID,
	name string,
	description string,
) (sesc.Department, error) {
	rec := event.Get(ctx).Sub("entdb/create_department")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"name", name,
		"description", description,
	)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := d.c.Department.Create().
		SetID(id).
		SetName(name).
		SetDescription(description).
		Save(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsConstraintError(err):
		return sesc.NoDepartment, sesc.ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't save department: %w", err)
		rec.Add(events.Error, err)
		return sesc.NoDepartment, err
	}

	rec.Sub("department").Set(
		"id", res.ID,
		"name", res.Name,
		"description", res.Description,
	)

	return sesc.Department{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
	}, nil
}

// DeleteDepartment implements sesc.DB.
func (d *DB) DeleteDepartment(ctx context.Context, id sesc.UUID) error {
	rec := event.Get(ctx).Sub("entdb/delete_department")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := d.c.Department.DeleteOneID(id).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsConstraintError(err):
		return sesc.ErrCannotRemoveDepartment
	case ent.IsNotFound(err):
		return sesc.ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't delete department: %w", err)
		rec.Add(events.Error, err)
		return err
	}

	rec.Set("success", true)
	return nil
}

// DepartmentByID implements sesc.DB.
func (d *DB) DepartmentByID(ctx context.Context, id sesc.UUID) (sesc.Department, error) {
	rec := event.Get(ctx).Sub("entdb/department_by_id")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := d.c.Department.Get(ctx, id)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		return sesc.NoDepartment, sesc.ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't get department: %w", err)
		rec.Add(events.Error, err)
		return sesc.NoDepartment, err
	}

	rec.Sub("department").Set(
		"id", res.ID,
		"name", res.Name,
		"description", res.Description,
	)

	return sesc.Department{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
	}, nil
}

// Departments implements sesc.DB.
func (d *DB) Departments(ctx context.Context) ([]sesc.Department, error) {
	rec := event.Get(ctx).Sub("entdb/departments")
	statrec := event.Get(ctx).Sub("stats")

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := d.c.Department.Query().All(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	if err != nil {
		err := fmt.Errorf("couldn't get all departments: %w", err)
		rec.Add(events.Error, err)
		return nil, err
	}

	deps := make([]sesc.Department, len(res))
	for i, r := range res {
		deps[i] = sesc.Department{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		}
	}

	return deps, nil
}

// SaveUser implements sesc.DB.
func (d *DB) SaveUser(ctx context.Context, opt sesc.UserUpdateOptions) (sesc.User, error) {
	rec := event.Get(ctx).Sub("entdb/save_user")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"first_name", opt.FirstName,
		"last_name", opt.LastName,
		"middle_name", opt.MiddleName,
		"picture_url", opt.PictureURL,
		"department_id", opt.DepartmentID,
		"new_role_id", opt.NewRoleID,
	)

	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)

	txStart := time.Now()
	tx, err := d.c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		err := fmt.Errorf("couldn't begin transaction: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, err
	}

	statrec.Add(events.PostgresQueries, 1)
	var dept *ent.Department
	if opt.DepartmentID != uuid.Nil {
		dept, err = tx.Department.Get(ctx, opt.DepartmentID)
		switch {
		case ent.IsNotFound(err):
			return sesc.User{}, rollback(tx, sesc.ErrInvalidDepartment)
		case err != nil:
			err := fmt.Errorf("couldn't query department: %w", err)
			txrec.Add(events.Error, err)
			return sesc.User{}, rollback(tx, err)
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
		return sesc.User{}, rollback(tx, err)
	}

	statrec.Add(events.PostgresQueries, 1)
	us, err := tx.User.Query().Where(user.ID(res.ID)).WithDepartment().Only(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't query user after saving them: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	user, err := convertUser(us)
	if err != nil {
		rec.Add(events.Error, err)
		return sesc.User{}, err
	}

	rec.Set("user", user.EventRecord())
	return user, nil
}

// UpdateDepartment implements sesc.DB.
func (d *DB) UpdateDepartment(
	ctx context.Context,
	id sesc.UUID,
	name string,
	description string,
) error {
	rec := event.Get(ctx).Sub("entdb/update_department")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"name", name,
		"description", description,
	)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := d.c.Department.UpdateOneID(id).SetName(name).SetDescription(description).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		joinedErr := errors.Join(err, sesc.ErrInvalidDepartment)
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

// UpdateProfilePicture implements sesc.DB.
func (d *DB) UpdateProfilePicture(ctx context.Context, id sesc.UUID, pictureURL string) error {
	rec := event.Get(ctx).Sub("entdb/update_profile_picture")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"picture_url", pictureURL,
	)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := d.c.User.UpdateOneID(id).SetPictureURL(pictureURL).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		joinedErr := errors.Join(err, sesc.ErrUserNotFound)
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

// UpdateUser implements sesc.DB.
func (d *DB) UpdateUser(
	ctx context.Context,
	id sesc.UUID,
	opt sesc.UserUpdateOptions,
) (sesc.User, error) {
	rec := event.Get(ctx).Sub("entdb/update_user")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"id", id,
		"first_name", opt.FirstName,
		"last_name", opt.LastName,
		"middle_name", opt.MiddleName,
		"picture_url", opt.PictureURL,
		"suspended", opt.Suspended,
		"department_id", opt.DepartmentID,
		"new_role_id", opt.NewRoleID,
	)

	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)

	txStart := time.Now()
	tx, err := d.c.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		err := fmt.Errorf("couldn't start transaction: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, err
	}

	statrec.Add(events.PostgresQueries, 1)
	us, err := tx.User.Get(ctx, id)
	switch {
	case ent.IsNotFound(err):
		txrec.Add(events.Error, sesc.ErrUserNotFound)
		return sesc.User{}, rollback(tx, sesc.ErrUserNotFound)
	case err != nil:
		err := fmt.Errorf("couldn't query user: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, rollback(tx, err)
	}

	var dept *ent.Department
	if opt.DepartmentID != uuid.Nil {
		statrec.Add(events.PostgresQueries, 1)
		dept, err = tx.Department.Get(ctx, opt.DepartmentID)
		switch {
		case ent.IsNotFound(err):
			txrec.Add(events.Error, sesc.ErrInvalidDepartment)
			return sesc.User{}, rollback(tx, sesc.ErrInvalidDepartment)
		case err != nil:
			err := fmt.Errorf("couldn't query department: %w", err)
			txrec.Add(events.Error, err)
			return sesc.User{}, rollback(tx, err)
		}
	}

	upd := us.Update().
		SetFirstName(opt.FirstName).
		SetLastName(opt.LastName).
		SetMiddleName(opt.MiddleName).
		SetPictureURL(opt.PictureURL).
		SetSuspended(opt.Suspended).
		SetRoleID(opt.NewRoleID)

	if dept != nil {
		upd = upd.SetDepartment(dept)
	} else {
		upd = upd.ClearDepartment()
	}

	statrec.Add(events.PostgresQueries, 1)
	_, err = upd.Save(ctx)

	if err != nil {
		err := fmt.Errorf("couldn't update user: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, rollback(tx, err)
	}

	statrec.Add(events.PostgresQueries, 1)
	us, err = tx.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't query user after an update: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return sesc.User{}, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	user, err := convertUser(us)
	if err != nil {
		rec.Add(events.Error, err)
		return sesc.User{}, err
	}

	rec.Set("user", user.EventRecord())
	return user, nil
}

// UserByID implements sesc.DB.
func (d *DB) UserByID(ctx context.Context, id sesc.UUID) (sesc.User, error) {
	rec := event.Get(ctx).Sub("entdb/user_by_id")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	u, err := d.c.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		rec.Add(events.Error, sesc.ErrUserNotFound)
		return sesc.User{}, sesc.ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't query user: %w", err)
		rec.Add(events.Error, err)
		return sesc.User{}, err
	}

	user, err := convertUser(u)
	if err != nil {
		rec.Add(events.Error, err)
		return sesc.User{}, err
	}

	rec.Set("user", user.EventRecord())
	return user, nil
}

// Users implements sesc.DB.
func (d *DB) Users(ctx context.Context) ([]sesc.User, error) {
	rec := event.Get(ctx).Sub("entdb/users")
	statrec := event.Get(ctx).Sub("stats")

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := d.c.User.Query().WithDepartment().All(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	if err != nil {
		err := fmt.Errorf("couldn't query users: %w", err)
		rec.Add(events.Error, err)
		return nil, err
	}

	users := make([]sesc.User, len(res))
	for i, r := range res {
		users[i], err = convertUser(r)
		if err != nil {
			return nil, fmt.Errorf("couldn't conver user: %w", err)
		}
	}

	return users, nil
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %w", err, rerr)
	}
	return err
}

func convertUser(u *ent.User) (sesc.User, error) {
	var dept sesc.Department
	dep := u.Edges.Department
	if dep != nil {
		dept = sesc.Department{
			ID:          dep.ID,
			Name:        dep.Name,
			Description: dep.Description,
		}
	}

	role, ok := sesc.RoleByID(u.RoleID)
	if !ok {
		return sesc.User{}, sesc.ErrInvalidRole
	}

	return sesc.User{
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

var _ sesc.DB = (*DB)(nil)
