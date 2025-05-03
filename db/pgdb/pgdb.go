package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type DB struct {
	db  *sqlx.DB
	log *slog.Logger
}

var _ sesc.DB = (*DB)(nil)

func Connect(ctx context.Context, log *slog.Logger, sescPGAddress string) (*DB, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", sescPGAddress)
	if err != nil {
		log.DebugContext(ctx, "couldn't connect to postgres", "error", err)
		return nil, fmt.Errorf("couldn't connect to postgres: %w", err)
	}

	return &DB{
		db:  db,
		log: log,
	}, nil
}

func (d *DB) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("couldn't close sqlx db: %w", err)
	}
	return nil
}

// CreateDepartment implements sesc.DB.
func (d *DB) CreateDepartment(
	ctx context.Context,
	id sesc.UUID,
	name string,
	description string,
) (sesc.Department, error) {
	query := `INSERT INTO departments (id, name, description) VALUES ($1, $2, $3)`
	_, err := d.db.ExecContext(ctx, query, id, name, description)
	if err != nil {
		d.log.DebugContext(ctx, "CreateDepartment: failed to insert", "error", err, "id", id, "name", name)
		return sesc.Department{}, fmt.Errorf("create department: %w", err)
	}
	return sesc.Department{ID: id, Name: name, Description: description}, nil
}

// DeleteDepartment implements sesc.DB.
func (d *DB) DeleteDepartment(ctx context.Context, id sesc.UUID) error {
	query := `DELETE FROM departments WHERE id = $1`
	result, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		d.log.DebugContext(ctx, "DeleteDepartment: failed to delete", "error", err, "id", id)
		return fmt.Errorf("delete department: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		d.log.DebugContext(ctx, "DeleteDepartment: no rows affected", "id", id)
		return fmt.Errorf("department not found")
	}
	return nil
}

// DepartmentByID implements sesc.DB.
func (d *DB) DepartmentByID(ctx context.Context, id sesc.UUID) (sesc.Department, error) {
	var department sesc.Department
	query := `SELECT id, name, description FROM departments WHERE id = $1`
	err := d.db.GetContext(ctx, &department, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			d.log.DebugContext(ctx, "DepartmentByID: not found", "id", id)
			return sesc.Department{}, fmt.Errorf("department not found: %w", err)
		}
		d.log.DebugContext(ctx, "DepartmentByID: failed to get", "error", err, "id", id)
		return sesc.Department{}, fmt.Errorf("get department: %w", err)
	}
	return department, nil
}

// Departments implements sesc.DB.
func (d *DB) Departments(ctx context.Context) ([]sesc.Department, error) {
	var departments []sesc.Department
	query := `SELECT id, name, description FROM departments`
	err := d.db.SelectContext(ctx, &departments, query)
	if err != nil {
		d.log.DebugContext(ctx, "Departments: failed to list", "error", err)
		return nil, fmt.Errorf("list departments: %w", err)
	}
	return departments, nil
}

// UpdateDepartment implements sesc.DB.
func (d *DB) UpdateDepartment(ctx context.Context, id sesc.UUID, name string, description string) error {
	query := `UPDATE departments SET name = $1, description = $2 WHERE id = $3`
	result, err := d.db.ExecContext(ctx, query, name, description, id)
	if err != nil {
		d.log.DebugContext(ctx, "UpdateDepartment: failed to update", "error", err, "id", id)
		return fmt.Errorf("update department: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		d.log.DebugContext(ctx, "UpdateDepartment: no rows affected", "id", id)
		return fmt.Errorf("department not found")
	}
	return nil
}

// SaveUser implements sesc.DB.
func (d *DB) SaveUser(ctx context.Context, user sesc.User) error {
	var departmentID *sesc.UUID
	if user.Department != sesc.NoDepartment {
		departmentID = &user.Department.ID
	}
	query := `
        INSERT INTO users (
            id, first_name, last_name, middle_name, picture_url, suspended,
            department_id, role_id, auth_id
        ) VALUES (
            :id, :first_name, :last_name, :middle_name, :picture_url, :suspended,
            :department_id, :role_id, :auth_id
        )`
	args := map[string]interface{}{
		"id":            user.ID,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"middle_name":   user.MiddleName,
		"picture_url":   user.PictureURL,
		"suspended":     user.Suspended,
		"department_id": departmentID,
		"role_id":       user.Role.ID,
		"auth_id":       user.AuthID,
	}
	_, err := d.db.NamedExecContext(ctx, query, args)
	if err != nil {
		d.log.DebugContext(ctx, "SaveUser: failed to insert", "error", err, "user_id", user.ID)
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

// UpdateProfilePicture implements sesc.DB.
func (d *DB) UpdateProfilePicture(ctx context.Context, userID sesc.UUID, url string) error {
	query := `UPDATE users SET picture_url = $1 WHERE id = $2`
	result, err := d.db.ExecContext(ctx, query, url, userID)
	if err != nil {
		d.log.DebugContext(ctx, "UpdateProfilePicture: failed", "error", err, "user_id", userID)
		return fmt.Errorf("update profile picture: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		d.log.DebugContext(ctx, "UpdateProfilePicture: no rows affected", "user_id", userID)
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdateUser implements sesc.DB.
func (d *DB) UpdateUser(ctx context.Context, id sesc.UUID, opts sesc.UserUpdateOptions) (sesc.User, error) {
	if _, ok := sesc.RoleByID(opts.NewRoleID); !ok {
		d.log.DebugContext(ctx, "UpdateUser: invalid role", "role_id", opts.NewRoleID)
		return sesc.User{}, fmt.Errorf("invalid role ID %d", opts.NewRoleID)
	}
	var departmentID *sesc.UUID
	if opts.DepartmentID != (sesc.UUID{}) {
		departmentID = &opts.DepartmentID
	}
	query := `
        UPDATE users SET
            first_name = :first_name,
            last_name = :last_name,
            middle_name = :middle_name,
            picture_url = :picture_url,
            suspended = :suspended,
            department_id = :department_id,
            role_id = :role_id,
            auth_id = :auth_id
        WHERE id = :id`
	args := map[string]any{
		"first_name":    opts.FirstName,
		"last_name":     opts.LastName,
		"middle_name":   opts.MiddleName,
		"picture_url":   opts.PictureURL,
		"suspended":     opts.Suspended,
		"department_id": departmentID,
		"role_id":       opts.NewRoleID,
		"auth_id":       opts.AuthID,
		"id":            id,
	}
	_, err := d.db.NamedExecContext(ctx, query, args)
	if err != nil {
		d.log.DebugContext(ctx, "UpdateUser: failed to update", "error", err, "user_id", id)
		return sesc.User{}, fmt.Errorf("update user: %w", err)
	}
	return d.UserByID(ctx, id)
}

// UserByID implements sesc.DB.
func (d *DB) UserByID(ctx context.Context, id sesc.UUID) (sesc.User, error) {
	var row struct {
		ID         sesc.UUID  `db:"id"`
		FirstName  string     `db:"first_name"`
		LastName   string     `db:"last_name"`
		MiddleName string     `db:"middle_name"`
		PictureURL string     `db:"picture_url"`
		Suspended  bool       `db:"suspended"`
		RoleID     int32      `db:"role_id"`
		AuthID     sesc.UUID  `db:"auth_id"`
		DepID      *sesc.UUID `db:"dep_id"`
		DepName    *string    `db:"dep_name"`
		DepDesc    *string    `db:"dep_description"`
	}

	query := `
        SELECT
            u.id,
            u.first_name,
            u.last_name,
            u.middle_name,
            u.picture_url,
            u.suspended,
            u.role_id,
            u.auth_id,
            d.id AS dep_id,
            d.name AS dep_name,
            d.description AS dep_description
        FROM users u
        LEFT JOIN departments d ON u.department_id = d.id
        WHERE u.id = $1`

	err := d.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			d.log.DebugContext(ctx, "UserByID: not found", "id", id)
			return sesc.User{}, fmt.Errorf("user not found: %w", err)
		}
		d.log.DebugContext(ctx, "UserByID: failed to get", "error", err, "id", id)
		return sesc.User{}, fmt.Errorf("get user: %w", err)
	}

	// Handle department data
	var department sesc.Department
	if row.DepID != nil {
		department = sesc.Department{
			ID:          *row.DepID,
			Name:        *row.DepName,
			Description: *row.DepDesc,
		}
	} else {
		department = sesc.NoDepartment
	}

	role, ok := sesc.RoleByID(row.RoleID)
	if !ok {
		d.log.ErrorContext(ctx, "UserByID: invalid role",
			"user_id", id, "role_id", row.RoleID)
		return sesc.User{}, fmt.Errorf("invalid role ID %d", row.RoleID)
	}

	return sesc.User{
		ID:         row.ID,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		MiddleName: row.MiddleName,
		PictureURL: row.PictureURL,
		Suspended:  row.Suspended,
		Department: department,
		Role:       role,
		AuthID:     row.AuthID,
	}, nil
}

// Users implements sesc.DB.
func (d *DB) Users(ctx context.Context) ([]sesc.User, error) {
	var users []struct {
		ID         sesc.UUID `db:"id"`
		FirstName  string    `db:"first_name"`
		LastName   string    `db:"last_name"`
		MiddleName string    `db:"middle_name"`
		PictureURL string    `db:"picture_url"`
		Suspended  bool      `db:"suspended"`
		RoleID     int32     `db:"role_id"`
		AuthID     sesc.UUID `db:"auth_id"`
		Department sesc.Department
	}
	query := `
        SELECT
            u.id, u.first_name, u.last_name, u.middle_name, u.picture_url, u.suspended,
            u.role_id, u.auth_id,
            d.id AS "department.id", d.name AS "department.name", d.description AS "department.description"
        FROM users u
        LEFT JOIN departments d ON u.department_id = d.id`
	err := d.db.SelectContext(ctx, &users, query)
	if err != nil {
		d.log.DebugContext(ctx, "Users: failed to list", "error", err)
		return nil, fmt.Errorf("list users: %w", err)
	}
	result := make([]sesc.User, 0, len(users))
	for _, u := range users {
		role, ok := sesc.RoleByID(u.RoleID)
		if !ok {
			d.log.ErrorContext(ctx, "got an invalid role in the database", "user_id", u.ID, "role_id", u.RoleID)
			return nil, fmt.Errorf("invalid role ID %d for user %s", u.RoleID, u.ID)
		}
		result = append(result, sesc.User{
			ID:         u.ID,
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			MiddleName: u.MiddleName,
			PictureURL: u.PictureURL,
			Suspended:  u.Suspended,
			Department: u.Department,
			Role:       role,
			AuthID:     u.AuthID,
		})
	}
	return result, nil
}
