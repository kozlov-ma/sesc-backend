package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/lib/pq"
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

// AssignHeadOfDepartment implements sesc.DB.
//
// Returns a sesc.ErrUserNotFound if user does not exist.
// Returns a sesc.ErrInvalidDepartment if department does not exist.
func (d *DB) AssignHeadOfDepartment(ctx context.Context, departmentID sesc.UUID, userID sesc.UUID) (rerr error) {
	tx, err := d.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
			rerr = errors.Join(rerr, tx.Rollback())
		}
	}()

	var deptExists bool
	err = tx.GetContext(ctx, &deptExists, "SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1)", departmentID)
	if err != nil {
		return fmt.Errorf("check department exists: %w", err)
	}
	if !deptExists {
		return sesc.ErrInvalidDepartment
	}

	// Check if user exists
	var userExists bool
	err = tx.GetContext(ctx, &userExists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID)
	if err != nil {
		return fmt.Errorf("check user exists: %w", err)
	}
	if !userExists {
		return sesc.ErrUserNotFound
	}

	// Update department's head
	_, err = tx.ExecContext(ctx, "UPDATE departments SET head_user_id = $1 WHERE id = $2", userID, departmentID)
	if err != nil {
		return fmt.Errorf("update department head: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// CreateDepartment implements sesc.DB.
//
// Returns a sesc.ErrInvalidDepartment if department with this name already exists.
func (d *DB) CreateDepartment(
	ctx context.Context,
	id sesc.UUID,
	name string,
	description string,
) (sesc.Department, error) {
	// Check for existing department with same name
	var exists bool
	err := d.db.GetContext(ctx, &exists,
		"SELECT EXISTS(SELECT 1 FROM departments WHERE name = $1)", name)
	if err != nil {
		return sesc.Department{}, fmt.Errorf("department existence check failed: %w", err)
	}
	if exists {
		return sesc.Department{}, sesc.ErrInvalidDepartment
	}

	// Insert new department without head user
	_, err = d.db.ExecContext(ctx, `
        INSERT INTO departments
            (id, name, description, head_user_id)
        VALUES
            ($1, $2, $3, NULL)
    `, id, name, description)

	if err != nil {
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == "23505" {
			return sesc.Department{}, sesc.ErrInvalidDepartment
		}
		return sesc.Department{}, fmt.Errorf("department creation failed: %w", err)
	}

	return sesc.Department{
		ID:          id,
		Name:        name,
		Description: description,
		Head:        nil,
	}, nil
}

// GrantExtraPermissions implements sesc.DB.
//
// If the user already has the extra permission, or it is granted by a role, it is a no-op.
// If the user does not exist, it returns a sesc.ErrUserNotFound.
// If the permission is not valid, it returns a sesc.ErrInvalidPermission.
func (d *DB) GrantExtraPermissions(
	ctx context.Context,
	user sesc.User,
	permissions ...sesc.Permission,
) (ruser sesc.User, rerr error) {
	// Check user exists
	var exists bool
	err := d.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", user.ID)
	if err != nil {
		return sesc.User{}, fmt.Errorf("couldn't check user exists: %w", err)
	}
	if !exists {
		return sesc.User{}, sesc.ErrUserNotFound
	}

	// Check permissions exist
	permIDs := make([]int32, len(permissions))
	for i, p := range permissions {
		permIDs[i] = p.ID
	}
	query, args, err := sqlx.In("SELECT COUNT(*) FROM permissions WHERE id IN (?)", permIDs)
	if err != nil {
		return sesc.User{}, fmt.Errorf("build permission check query: %w", err)
	}
	query = d.db.Rebind(query)
	var count int
	err = d.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return sesc.User{}, fmt.Errorf("check permissions exist: %w", err)
	}
	if count != len(permIDs) {
		return sesc.User{}, sesc.ErrInvalidPermission
	}

	tx, err := d.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return sesc.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
			rerr = errors.Join(rerr, tx.Rollback())
		}
	}()

	var roleID *sesc.UUID
	err = tx.GetContext(ctx, &roleID, "SELECT role_id FROM users WHERE id = $1", user.ID)
	if err != nil {
		return sesc.User{}, fmt.Errorf("get user role: %w", err)
	}

	for _, perm := range permissions {
		// Check if role grants permission
		var roleHasPerm bool
		if roleID != nil {
			err = tx.GetContext(ctx, &roleHasPerm,
				"SELECT EXISTS(SELECT 1 FROM permissions_roles WHERE role_id = $1 AND permission_id = $2)",
				roleID, perm.ID,
			)
			if err != nil {
				return sesc.User{}, fmt.Errorf("check role permission: %w", err)
			}
		}
		if roleHasPerm {
			continue
		}

		// Check if already has extra permission
		var hasExtra bool
		err = tx.GetContext(ctx, &hasExtra,
			"SELECT EXISTS(SELECT 1 FROM users_extra_permissions WHERE user_id = $1 AND permission_id = $2)",
			user.ID, perm.ID,
		)
		if err != nil {
			return sesc.User{}, fmt.Errorf("check extra permission: %w", err)
		}
		if hasExtra {
			continue
		}

		// Grant permission
		_, err = tx.ExecContext(ctx,
			"INSERT INTO users_extra_permissions (user_id, permission_id) VALUES ($1, $2)",
			user.ID, perm.ID,
		)
		if err != nil {
			return sesc.User{}, fmt.Errorf("grant extra permission: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return sesc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return d.UserByID(ctx, user.ID)
}

// RevokeExtraPermissions implements sesc.DB.
//
// If the user does not have the permission, it is a no-op.
// If the user does not exist, it returns a sesc.ErrUserNotFound.
// If the permission is not valid, it returns a sesc.ErrInvalidPermission.
//
// Permissions granted by roles are not affected by this operation.
func (d *DB) RevokeExtraPermissions(
	ctx context.Context,
	user sesc.User,
	permissions ...sesc.Permission,
) (ruser sesc.User, rerr error) {
	// Check user exists
	var exists bool
	err := d.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", user.ID)
	if err != nil {
		return sesc.User{}, fmt.Errorf("check user exists: %w", err)
	}
	if !exists {
		return sesc.User{}, sesc.ErrUserNotFound
	}

	// Check permissions exist
	permIDs := make([]int32, len(permissions))
	for i, p := range permissions {
		permIDs[i] = p.ID
	}
	query, args, err := sqlx.In("SELECT COUNT(*) FROM permissions WHERE id IN (?)", permIDs)
	if err != nil {
		return sesc.User{}, fmt.Errorf("build permission check query: %w", err)
	}
	query = d.db.Rebind(query)
	var count int
	err = d.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return sesc.User{}, fmt.Errorf("check permissions exist: %w", err)
	}
	if count != len(permIDs) {
		return sesc.User{}, sesc.ErrInvalidPermission
	}

	tx, err := d.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return sesc.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
			rerr = errors.Join(rerr, tx.Rollback())
		}
	}()

	var roleID *sesc.UUID
	err = tx.GetContext(ctx, &roleID, "SELECT role_id FROM users WHERE id = $1", user.ID)
	if err != nil {
		return sesc.User{}, fmt.Errorf("get user role: %w", err)
	}

	for _, perm := range permissions {
		// Check if role grants permission
		var roleHasPerm bool
		if roleID != nil {
			err = tx.GetContext(ctx, &roleHasPerm,
				"SELECT EXISTS(SELECT 1 FROM permissions_roles WHERE role_id = $1 AND permission_id = $2)",
				roleID, perm.ID,
			)
			if err != nil {
				return sesc.User{}, fmt.Errorf("check role permission: %w", err)
			}
		}
		if roleHasPerm {
			continue
		}

		_, err = tx.ExecContext(ctx,
			"DELETE FROM users_extra_permissions WHERE user_id = $1 AND permission_id = $2",
			user.ID, perm.ID,
		)
		if err != nil {
			return sesc.User{}, fmt.Errorf("revoke extra permission: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return sesc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return d.UserByID(ctx, user.ID)
}

// SaveUser implements sesc.DB.
func (d *DB) SaveUser(ctx context.Context, user sesc.User) (rerr error) {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
			rerr = errors.Join(rerr, tx.Rollback())
		}
	}()

	// Check existence within transaction
	var exists bool
	err = tx.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", user.ID)
	if err != nil {
		return fmt.Errorf("check user existence: %w", err)
	}

	// Prepare nullable department/role IDs
	var deptID *sesc.UUID
	if user.Department != sesc.NoDepartment {
		deptID = &user.Department.ID
	}

	var roleID *int32
	if user.Role.ID != 0 {
		roleID = &user.Role.ID
	}

	// Execute appropriate operation
	if exists {
		_, err = tx.ExecContext(ctx, `
            UPDATE users SET
                first_name = $1,
                last_name = $2,
                middle_name = $3,
                picture_url = $4,
                suspended = $5,
                department_id = $6,
                role_id = $7,
                auth_id = $8
            WHERE id = $9`,
			user.FirstName,
			user.LastName,
			user.MiddleName,
			user.PictureURL,
			user.Suspended,
			deptID,
			roleID,
			user.AuthID,
			user.ID,
		)
	} else {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO users (
                id, first_name, last_name, middle_name,
                picture_url, suspended, department_id,
                role_id, auth_id
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			user.ID,
			user.FirstName,
			user.LastName,
			user.MiddleName,
			user.PictureURL,
			user.Suspended,
			deptID,
			roleID,
			user.AuthID,
		)
	}

	if err != nil {
		return fmt.Errorf("save operation failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// UserByID implements sesc.DB.
//
// Returns a db.ErrUserNotFound if user does not exist.
func (d *DB) UserByID(ctx context.Context, id sesc.UUID) (sesc.User, error) {
	var userRow struct {
		ID           sesc.UUID  `db:"id"`
		FirstName    string     `db:"first_name"`
		LastName     string     `db:"last_name"`
		MiddleName   string     `db:"middle_name"`
		PictureURL   string     `db:"picture_url"`
		Suspended    bool       `db:"suspended"`
		DepartmentID *sesc.UUID `db:"department_id"`
		RoleID       *int32     `db:"role_id"`
		AuthID       sesc.UUID  `db:"auth_id"`
		DeptName     *string    `db:"dept_name"`
		DeptDesc     *string    `db:"dept_desc"`
		DeptHeadID   *sesc.UUID `db:"dept_head_id"`
		RoleName     *string    `db:"role_name"`
	}

	query := `
        SELECT u.*, d.name AS dept_name, d.description AS dept_desc, d.head_user_id AS dept_head_id, r.name AS role_name
        FROM users u
        LEFT JOIN departments d ON u.department_id = d.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE u.id = $1
    `
	err := d.db.GetContext(ctx, &userRow, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return sesc.User{}, sesc.ErrUserNotFound
		}
		return sesc.User{}, fmt.Errorf("query user: %w", err)
	}

	user := sesc.User{
		ID:         userRow.ID,
		FirstName:  userRow.FirstName,
		LastName:   userRow.LastName,
		MiddleName: userRow.MiddleName,
		PictureURL: userRow.PictureURL,
		Suspended:  userRow.Suspended,
		AuthID:     userRow.AuthID,
	}

	// Populate department
	if userRow.DepartmentID != nil {
		dept := sesc.Department{
			ID:          *userRow.DepartmentID,
			Name:        *userRow.DeptName,
			Description: *userRow.DeptDesc,
		}
		if userRow.DeptHeadID != nil {
			headUser, err := d.UserByID(ctx, *userRow.DeptHeadID)
			if err != nil {
				d.log.Error("failed to fetch department head", "error", err)
			} else {
				dept.Head = &headUser
			}
		}
		user.Department = dept
	}

	// Populate role
	if userRow.RoleID != nil {
		role := sesc.Role{
			ID:   *userRow.RoleID,
			Name: *userRow.RoleName,
		}
		var perms []sesc.Permission
		query = `
            SELECT p.id, p.name, p.description
            FROM permissions_roles pr
            JOIN permissions p ON pr.permission_id = p.id
            WHERE pr.role_id = $1
        `
		err = d.db.SelectContext(ctx, &perms, query, role.ID)
		if err != nil {
			return sesc.User{}, fmt.Errorf("fetch role permissions: %w", err)
		}
		role.Permissions = perms
		user.Role = role
	}

	// Populate extra permissions
	var extraPerms []sesc.Permission
	query = `
        SELECT p.id, p.name, p.description
        FROM users_extra_permissions uep
        JOIN permissions p ON uep.permission_id = p.id
        WHERE uep.user_id = $1
    `
	err = d.db.SelectContext(ctx, &extraPerms, query, id)
	if err != nil {
		return sesc.User{}, fmt.Errorf("fetch extra permissions: %w", err)
	}
	user.ExtraPermissions = extraPerms

	return user, nil
}

// Departments implements sesc.DB.
func (d *DB) Departments(ctx context.Context) ([]sesc.Department, error) {
	var deptRows []struct {
		ID          uuid.UUID  `db:"id"`
		Name        string     `db:"name"`
		Description string     `db:"description"`
		HeadUserID  *uuid.UUID `db:"head_user_id"`
	}

	err := d.db.SelectContext(ctx, &deptRows,
		`SELECT id, name, description, head_user_id FROM departments`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch departments: %w", err)
	}
	if len(deptRows) == 0 {
		return []sesc.Department{}, nil
	}

	// Collect head user IDs
	headUserIDs := make([]uuid.UUID, len(deptRows))
	for i, row := range deptRows {
		if row.HeadUserID != nil {
			headUserIDs[i] = *row.HeadUserID
		}
	}

	// Second query: batch fetch head users with basic info
	var userRows []struct {
		ID           uuid.UUID  `db:"id"`
		FirstName    string     `db:"first_name"`
		LastName     string     `db:"last_name"`
		MiddleName   string     `db:"middle_name"`
		PictureURL   string     `db:"picture_url"`
		Suspended    bool       `db:"suspended"`
		DepartmentID *uuid.UUID `db:"department_id"`
		RoleID       *int32     `db:"role_id"`
		AuthID       uuid.UUID  `db:"auth_id"`
	}

	query, args, err := sqlx.In(
		`SELECT
            id, first_name, last_name, middle_name,
            picture_url, suspended, department_id,
            role_id, auth_id
         FROM users
         WHERE id IN (?)`,
		headUserIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build user query: %w", err)
	}

	err = d.db.SelectContext(ctx, &userRows, d.db.Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch head users: %w", err)
	}

	userMap := make(map[uuid.UUID]sesc.User)
	for _, ur := range userRows {
		user := sesc.User{
			ID:         ur.ID,
			FirstName:  ur.FirstName,
			LastName:   ur.LastName,
			MiddleName: ur.MiddleName,
			PictureURL: ur.PictureURL,
			Suspended:  ur.Suspended,
			AuthID:     ur.AuthID,
		}

		if ur.DepartmentID != nil {
			user.Department = sesc.Department{ID: *ur.DepartmentID}
		}

		if ur.RoleID != nil {
			user.Role = sesc.Role{ID: *ur.RoleID}
		}

		userMap[ur.ID] = user
	}

	departments := make([]sesc.Department, 0, len(deptRows))
	for _, row := range deptRows {
		if row.HeadUserID == nil {
			departments = append(departments, sesc.Department{
				ID:          row.ID,
				Name:        row.Name,
				Description: row.Description,
			})
			continue
		}

		headUser, exists := userMap[*row.HeadUserID]
		if !exists {
			d.log.ErrorContext(ctx,
				"department head user not found",
				"department_id", row.ID,
				"head_user_id", row.HeadUserID,
			)
			continue
		}

		departments = append(departments, sesc.Department{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			Head:        &headUser,
		})
	}

	return departments, nil
}

// InsertDefaultPermissions implements sesc.DB.
func (d *DB) InsertDefaultPermissions(ctx context.Context, permissions []sesc.Permission) error {
	query := `
        INSERT INTO permissions (id, name, description)
        VALUES (:id, :name, :description)
        ON CONFLICT (id) DO NOTHING`

	permMaps := make([]map[string]any, 0, len(permissions))
	for _, p := range permissions {
		permMaps = append(permMaps, map[string]any{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
		})
	}

	_, err := d.db.NamedExecContext(ctx, query, permMaps)
	if err != nil {
		return fmt.Errorf("failed to bulk insert permissions: %w", err)
	}
	return nil
}

// InsertDefaultRoles implements sesc.DB.
func (d *DB) InsertDefaultRoles(ctx context.Context, roles []sesc.Role) (rerr error) {
	tx, err := d.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
			rerr = errors.Join(rerr, tx.Rollback())
		}
	}()

	roleQuery := `
        INSERT INTO roles (id, name)
        VALUES (:id, :name)
        ON CONFLICT (id) DO NOTHING`

	roleMaps := make([]map[string]any, 0, len(roles))
	for _, r := range roles {
		roleMaps = append(roleMaps, map[string]any{
			"id":   r.ID,
			"name": r.Name,
		})
	}

	_, err = tx.NamedExecContext(ctx, roleQuery, roleMaps)
	if err != nil {
		return fmt.Errorf("failed to bulk insert roles: %w", err)
	}

	// Prepare permission-role relationships
	type permRole struct {
		PermissionID int32 `db:"permission_id"`
		RoleID       int32 `db:"role_id"`
	}

	var relationships []permRole
	for _, role := range roles {
		for _, perm := range role.Permissions {
			relationships = append(relationships, permRole{
				PermissionID: perm.ID,
				RoleID:       role.ID,
			})
		}
	}

	relQuery := `
        INSERT INTO permissions_roles (permission_id, role_id)
        VALUES (:permission_id, :role_id)
        ON CONFLICT (permission_id, role_id) DO NOTHING`

	_, err = tx.NamedExecContext(ctx, relQuery, relationships)
	if err != nil {
		return fmt.Errorf("failed to bulk insert role permissions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *DB) DepartmentByID(ctx context.Context, id sesc.UUID) (sesc.Department, error) {
	var row struct {
		ID          uuid.UUID  `db:"id"`
		Name        string     `db:"name"`
		Description string     `db:"description"`
		HeadUserID  *uuid.UUID `db:"head_user_id"`
	}

	err := d.db.GetContext(ctx, &row, `
        SELECT id, name, description, head_user_id
        FROM departments
        WHERE id = $1
    `, id)
	if err != nil {
		if err == sql.ErrNoRows {
			d.log.DebugContext(ctx, "department not found", "department_id", id)
			return sesc.Department{}, sesc.ErrInvalidDepartment
		}
		return sesc.Department{}, fmt.Errorf("failed to query department: %w", err)
	}

	department := sesc.Department{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
	}

	if row.HeadUserID != nil {
		headUser, err := d.UserByID(ctx, *row.HeadUserID)
		switch {
		case errors.Is(err, sesc.ErrUserNotFound):
			d.log.WarnContext(ctx,
				"department head user not found",
				"department_id", id,
				"head_user_id", *row.HeadUserID,
			)
		case err != nil:
			return sesc.Department{}, fmt.Errorf("failed to resolve head user: %w", err)
		default:
			department.Head = &headUser
		}
	}

	return department, nil
}
