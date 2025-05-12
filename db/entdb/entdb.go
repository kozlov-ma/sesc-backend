package entdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type DB struct {
	log *slog.Logger
	c   *ent.Client
}

func New(log *slog.Logger, c *ent.Client) *DB {
	return &DB{
		log: log,
		c:   c,
	}
}

// CreateDepartment implements sesc.DB.
func (d *DB) CreateDepartment(
	ctx context.Context,
	id sesc.UUID,
	name string,
	description string,
) (sesc.Department, error) {
	res, err := d.c.Department.Create().SetID(id).SetName(name).SetDescription(description).Save(ctx)
	switch {
	case ent.IsConstraintError(err):
		return sesc.NoDepartment, sesc.ErrInvalidDepartment
	case err != nil:
		return sesc.NoDepartment, fmt.Errorf("couldn't save department: %w", err)
	}

	return sesc.Department{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
	}, nil
}

// DeleteDepartment implements sesc.DB.
func (d *DB) DeleteDepartment(ctx context.Context, id sesc.UUID) error {
	err := d.c.Department.DeleteOneID(id).Exec(ctx)
	switch {
	case ent.IsConstraintError(err):
		return sesc.ErrCannotRemoveDepartment
	case ent.IsNotFound(err):
		return sesc.ErrInvalidDepartment
	case err != nil:
		return fmt.Errorf("couldn't delete department: %w", err)
	}
	return nil
}

// DepartmentByID implements sesc.DB.
func (d *DB) DepartmentByID(ctx context.Context, id sesc.UUID) (sesc.Department, error) {
	res, err := d.c.Department.Get(ctx, id)
	switch {
	case ent.IsNotFound(err):
		return sesc.NoDepartment, sesc.ErrInvalidDepartment
	case err != nil:
		return sesc.NoDepartment, fmt.Errorf("couldn't get department: %w", err)
	}
	return sesc.Department{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
	}, nil
}

// Departments implements sesc.DB.
func (d *DB) Departments(ctx context.Context) ([]sesc.Department, error) {
	res, err := d.c.Department.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get all departments: %w", err)
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
func (d *DB) SaveUser(ctx context.Context, u sesc.User) error {
	_, err := d.c.User.Create().
		SetID(u.ID).
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetMiddleName(u.MiddleName).
		SetPictureURL(u.PictureURL).
		SetDepartmentID(u.Department.ID).
		SetRoleID(u.Role.ID).
		Save(ctx)

	switch {
	case ent.IsNotFound(err):
		return errors.Join(err, sesc.ErrInvalidDepartment)
	case err != nil:
		return fmt.Errorf("couldn't save user: %w", err)
	}

	return nil
}

// UpdateDepartment implements sesc.DB.
func (d *DB) UpdateDepartment(ctx context.Context, id sesc.UUID, name string, description string) error {
	err := d.c.Department.UpdateOneID(id).SetName(name).SetDescription(description).Exec(ctx)
	switch {
	case ent.IsNotFound(err):
		return errors.Join(err, sesc.ErrInvalidDepartment)
	case err != nil:
		return fmt.Errorf("couldn't update department: %w", err)
	}
	return nil
}

// UpdateProfilePicture implements sesc.DB.
func (d *DB) UpdateProfilePicture(ctx context.Context, id sesc.UUID, pictureURL string) error {
	err := d.c.User.UpdateOneID(id).SetPictureURL(pictureURL).Exec(ctx)
	switch {
	case ent.IsNotFound(err):
		return errors.Join(err, sesc.ErrUserNotFound)
	case err != nil:
		return fmt.Errorf("couldn't update user: %w", err)
	}
	return nil
}

// UpdateUser implements sesc.DB.
func (d *DB) UpdateUser(ctx context.Context, id sesc.UUID, opt sesc.UserUpdateOptions) (sesc.User, error) {
	tx, err := d.c.Tx(ctx)
	if err != nil {
		return sesc.User{}, fmt.Errorf("couldn't start transaction: %w", err)
	}

	user, err := tx.User.Query().ForUpdate().Where(user.ID(id)).WithDepartment().Only(ctx)
	switch {
	case ent.IsNotFound(err):
		return sesc.User{}, rollback(tx, sesc.ErrUserNotFound)
	case err != nil:
		return sesc.User{}, rollback(tx, fmt.Errorf("couldn't query user: %w", err))
	}

	res, err := user.Update().
		SetFirstName(opt.FirstName).
		SetLastName(opt.LastName).
		SetMiddleName(opt.MiddleName).
		SetPictureURL(opt.PictureURL).
		SetSuspended(opt.Suspended).
		SetDepartmentID(opt.DepartmentID).
		SetRoleID(opt.NewRoleID).
		Save(ctx)

	if err != nil {
		return sesc.User{}, rollback(tx, fmt.Errorf("couldn't update user: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return sesc.User{}, fmt.Errorf("couldn't commit transaction: %w", err)
	}

	return convertUser(res)
}

// UserByID implements sesc.DB.
func (d *DB) UserByID(ctx context.Context, id sesc.UUID) (sesc.User, error) {
	user, err := d.c.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	switch {
	case ent.IsNotFound(err):
		return sesc.User{}, sesc.ErrUserNotFound
	case err != nil:
		return sesc.User{}, fmt.Errorf("couldn't query user: %w", err)
	}

	return convertUser(user)
}

// Users implements sesc.DB.
func (d *DB) Users(ctx context.Context) ([]sesc.User, error) {
	res, err := d.c.User.Query().WithDepartment().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't query users: %w", err)
	}

	users := make([]sesc.User, len(res))
	for i, r := range res {
		var err error
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
		err = fmt.Errorf("%w: %v", err, rerr)
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
