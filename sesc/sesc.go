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
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/create_department")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Sub("params").Set(
		"name", name,
		"description", description,
	)

	// Stage 1: Generate UUID
	ctx = rec.Sub("generate_department_id").Wrap(ctx)
	id, err := s.generateDepartmentID(ctx)
	if err != nil {
		return NoDepartment, err
	}

	// Stage 2: Create department record
	ctx = rec.Sub("create_department_record").Wrap(ctx)
	department, err := s.createDepartmentRecord(ctx, statrec, id, name, description)
	if ent.IsValidationError(err) {
		return NoDepartment, ErrInvalidName
	}
	if err != nil {
		return NoDepartment, err
	}

	return department, nil
}

// generateDepartmentID generates a UUID for a new department
func (s *SESC) generateDepartmentID(ctx context.Context) (UUID, error) {
	rec := event.Get(ctx)

	id, err := s.newUUID()
	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return UUID{}, err
	}

	rec.Set("success", true)
	rec.Set("id", id)
	return id, nil
}

// createDepartmentRecord creates a department record in the database
func (s *SESC) createDepartmentRecord(
	ctx context.Context,
	statrec *event.Record,
	id UUID,
	name string,
	description string,
) (Department, error) {
	rec := event.Get(ctx)
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
		rec.Set("success", false)
		rec.Add(events.Error, ErrInvalidDepartment)
		return NoDepartment, ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't save department: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return NoDepartment, err
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
// Returns an ErrInvalidDepartment if the department does not exist.
func (s *SESC) DepartmentByID(ctx context.Context, id UUID) (Department, error) {
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
// Returns an ErrInvalidDepartment if the department does not exist.
func (s *SESC) UpdateDepartment(
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
func (s *SESC) updateDepartmentRecord(
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
		joinedErr := fmt.Errorf("%w: %w", err, ErrInvalidDepartment)
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
// Returns an ErrInvalidDepartment if the department does not exist.
// Returns an ErrCannotRemoveDepartment if the department has users.
func (s *SESC) DeleteDepartment(ctx context.Context, id UUID) error {
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
func (s *SESC) deleteDepartmentRecord(ctx context.Context, id UUID) error {
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
		rec.Add(events.Error, ErrCannotRemoveDepartment)
		rec.Set("success", false)
		return ErrCannotRemoveDepartment
	case ent.IsNotFound(err):
		rec.Add(events.Error, ErrInvalidDepartment)
		rec.Set("success", false)
		return ErrInvalidDepartment
	case err != nil:
		err := fmt.Errorf("couldn't delete department: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
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
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/update_user")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

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

	// Stage 1: Validate user exists
	ctx = rec.Sub("validate_user_exists").Wrap(ctx)
	if err := s.validateUserExists(ctx, id); err != nil {
		return User{}, err
	}

	// Stage 2: Validate role
	ctx = rec.Sub("validate_role").Wrap(ctx)
	if err := s.validateRole(ctx, upd.NewRoleID); err != nil {
		return User{}, err
	}

	// Stage 3: Validate name
	ctx = rec.Sub("validate_name").Wrap(ctx)
	if err := s.validateName(ctx, upd.FirstName, upd.LastName); err != nil {
		return User{}, err
	}

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

	// Stage 4: Check and get department if needed
	ctx = rec.Sub("check_department").Wrap(ctx)
	dept, err := s.checkAndGetDepartment(ctx, statrec, tx, upd.DepartmentID)
	if err != nil {
		return User{}, rollback(tx, err)
	}

	// Stage 5: Update user
	ctx = rec.Sub("update_user_record").Wrap(ctx)
	if err := s.updateUserRecord(ctx, statrec, tx, id, upd, dept); err != nil {
		return User{}, rollback(tx, err)
	}

	// Stage 6: Query updated user
	ctx = rec.Sub("query_updated_user").Wrap(ctx)
	us, err := s.queryUpdatedUser(ctx, statrec, tx, id)
	if err != nil {
		return User{}, rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return User{}, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	// Stage 7: Convert user entity to domain object
	ctx = rec.Sub("convert_user").Wrap(ctx)
	updated, err := s.convertUserEntity(ctx, us)
	if err != nil {
		return User{}, err
	}

	rec.Set("success", true)
	rec.Set("user", updated.EventRecord())
	return updated, nil
}

// validateUserExists validates that a user exists
func (s *SESC) validateUserExists(ctx context.Context, id UUID) error {
	rec := event.Get(ctx)
	rec.Set("user_id", id)

	_, err := s.UserByID(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("exists", false)
		return err
	}

	rec.Set("exists", true)
	return nil
}

// validateRole validates the role ID
func (s *SESC) validateRole(ctx context.Context, roleID int32) error {
	rec := event.Get(ctx)
	rec.Set("role_id", roleID)

	if roleID == 0 {
		rec.Set("valid", true)
		return nil
	}

	_, ok := RoleByID(roleID)
	if !ok {
		rec.Set("valid", false)
		return ErrInvalidRole
	}

	rec.Set("valid", true)
	return nil
}

// validateName validates that the name is not empty
func (s *SESC) validateName(ctx context.Context, firstName, lastName string) error {
	rec := event.Get(ctx)
	rec.Set(
		"first_name", firstName,
		"last_name", lastName,
	)

	if firstName == "" || lastName == "" {
		rec.Set("valid", false)
		return ErrInvalidName
	}

	rec.Set("valid", true)
	return nil
}

// checkAndGetDepartment checks if the department exists and returns it
func (s *SESC) checkAndGetDepartment(
	ctx context.Context,
	statrec *event.Record,
	tx *ent.Tx,
	departmentID UUID,
) (*ent.Department, error) {
	rec := event.Get(ctx)
	rec.Set("department_id", departmentID)

	if departmentID == uuid.Nil {
		rec.Set("required", false)
		//nolint:nilnil // department should be deleted or should not be set.
		return nil, nil
	}

	rec.Set("required", true)
	statrec.Add(events.PostgresQueries, 1)

	dept, err := tx.Department.Get(ctx, departmentID)
	switch {
	case ent.IsNotFound(err):
		rec.Set("exists", false)
		rec.Add(events.Error, ErrInvalidDepartment)
		return nil, ErrInvalidDepartment
	case err != nil:
		rec.Set("exists", false)
		err := fmt.Errorf("couldn't query department: %w", err)
		rec.Add(events.Error, err)
		return nil, err
	}

	rec.Set(
		"exists", true,
		"id", dept.ID,
		"name", dept.Name,
	)
	return dept, nil
}

// updateUserRecord updates the user record in the database
func (s *SESC) updateUserRecord(
	ctx context.Context,
	statrec *event.Record,
	tx *ent.Tx,
	id UUID,
	upd UserUpdateOptions,
	dept *ent.Department,
) error {
	rec := event.Get(ctx)
	rec.Set("user_id", id)

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

	_, err := updater.Save(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't update user: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	return nil
}

// queryUpdatedUser queries the updated user from the database
func (s *SESC) queryUpdatedUser(
	ctx context.Context,
	statrec *event.Record,
	tx *ent.Tx,
	id UUID,
) (*ent.User, error) {
	rec := event.Get(ctx)
	rec.Set("user_id", id)

	statrec.Add(events.PostgresQueries, 1)
	us, err := tx.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't query user after an update: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return nil, err
	}

	rec.Set("success", true)
	return us, nil
}

// convertUserEntity converts an ent.User to a User domain object
func (s *SESC) convertUserEntity(
	ctx context.Context,
	us *ent.User,
) (User, error) {
	rec := event.Get(ctx)

	updated, err := convertUser(us)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return User{}, err
	}

	rec.Set("success", true)
	return updated, nil
}

// CreateUser creates a new User with a specified role.
//
// Returns an ErrInvalidName if the first or last name is missing.
func (s *SESC) CreateUser(ctx context.Context, opt UserUpdateOptions) (User, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/create_user")
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Sub("params").Set(
		"first_name", opt.FirstName,
		"last_name", opt.LastName,
		"middle_name", opt.MiddleName,
		"picture_url", opt.PictureURL,
		"suspended", opt.Suspended,
		"department_id", opt.DepartmentID,
		"new_role_id", opt.NewRoleID,
	)

	// Stage 1: Validate input
	ctx = rec.Sub("validate_create_input").Wrap(ctx)
	if err := s.validateCreateInput(ctx, opt); err != nil {
		return User{}, err
	}

	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)

	txStart := time.Now()
	tx, err := s.client.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		err := fmt.Errorf("couldn't begin transaction: %w", err)
		txrec.Add(events.Error, err)
		return User{}, err
	}

	// Stage 2: Check and get department if needed
	ctx = rec.Sub("check_department").Wrap(ctx)
	dept, err := s.checkAndGetDepartment(ctx, statrec, tx, opt.DepartmentID)
	if err != nil {
		return User{}, rollback(tx, err)
	}

	// Stage 3: Create user record
	ctx = rec.Sub("create_user_record").Wrap(ctx)
	userID, err := s.createUserRecord(ctx, statrec, tx, opt, dept)
	if err != nil {
		return User{}, rollback(tx, err)
	}

	// Stage 4: Query created user
	ctx = rec.Sub("query_created_user").Wrap(ctx)
	us, err := s.queryCreatedUser(ctx, statrec, tx, userID)
	if err != nil {
		return User{}, rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return User{}, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	// Stage 5: Convert user entity to domain object
	ctx = rec.Sub("convert_user").Wrap(ctx)
	user, err := s.convertUserEntity(ctx, us)
	if err != nil {
		return User{}, err
	}

	rec.Set("success", true)
	rec.Set("user", user.EventRecord())
	return user, nil
}

// validateCreateInput validates the create user input
func (s *SESC) validateCreateInput(ctx context.Context, opt UserUpdateOptions) error {
	rec := event.Get(ctx)

	if err := opt.Validate(); err != nil {
		rec.Add(events.Error, err)
		rec.Set("valid", false)
		return err
	}

	rec.Set("valid", true)
	return nil
}

// createUserRecord creates a new user record in the database
func (s *SESC) createUserRecord(
	ctx context.Context,
	statrec *event.Record,
	tx *ent.Tx,
	opt UserUpdateOptions,
	dept *ent.Department,
) (UUID, error) {
	rec := event.Get(ctx)

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
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return UUID{}, err
	}

	rec.Set("success", true)
	rec.Set("user_id", res.ID)
	return res.ID, nil
}

// queryCreatedUser queries the newly created user from the database
func (s *SESC) queryCreatedUser(
	ctx context.Context,
	statrec *event.Record,
	tx *ent.Tx,
	id UUID,
) (*ent.User, error) {
	rec := event.Get(ctx)
	rec.Set("user_id", id)

	statrec.Add(events.PostgresQueries, 1)
	us, err := tx.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't query user after saving them: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return nil, err
	}

	rec.Set("success", true)
	return us, nil
}

// UpdateProfilePicture updates a user's profile picture.
// Returns an ErrUserNotFound if the user does not exist.
func (s *SESC) UpdateProfilePicture(ctx context.Context, id UUID, pictureURL string) error {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/update_profile_picture")

	rec.Sub("params").Set(
		"id", id,
		"picture_url", pictureURL,
	)

	// Stage 1: Update profile picture
	ctx = rec.Sub("update_profile_picture_record").Wrap(ctx)
	if err := s.updateProfilePictureRecord(ctx, id, pictureURL); err != nil {
		return err
	}

	rec.Set("success", true)
	return nil
}

// updateProfilePictureRecord updates a user's profile picture in the database
func (s *SESC) updateProfilePictureRecord(ctx context.Context, id UUID, pictureURL string) error {
	rec := event.Get(ctx)
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Set("id", id)
	rec.Set("picture_url", pictureURL)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	err := s.client.User.UpdateOneID(id).SetPictureURL(pictureURL).Exec(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		joinedErr := fmt.Errorf("%w: %w", err, ErrUserNotFound)
		rec.Add(events.Error, joinedErr)
		rec.Set("success", false)
		return joinedErr
	case err != nil:
		err := fmt.Errorf("couldn't update user: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	return nil
}

// UserByID gets a user by their ID.
// Returns an ErrUserNotFound if the user does not exist.
func (s *SESC) UserByID(ctx context.Context, id UUID) (User, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/user_by_id")

	rec.Sub("params").Set("id", id)

	// Stage 1: Query user by ID
	ctx = rec.Sub("query_user_by_id").Wrap(ctx)
	u, err := s.getUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	// Stage 2: Convert user entity
	ctx = rec.Sub("convert_user_entity").Wrap(ctx)
	userObj, err := s.convertUserFromEntity(ctx, u)
	if err != nil {
		return User{}, err
	}

	return userObj, nil
}

// getUserByID queries a user by ID from the database
func (s *SESC) getUserByID(ctx context.Context, id UUID) (*ent.User, error) {
	rec := event.Get(ctx)
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Set("id", id)

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	u, err := s.client.User.Query().Where(user.ID(id)).WithDepartment().Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		rec.Add(events.Error, ErrUserNotFound)
		rec.Set("success", false)
		return nil, ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't query user: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return nil, err
	}

	rec.Set("success", true)
	return u, nil
}

// convertUserFromEntity converts an ent.User to a User domain object
func (s *SESC) convertUserFromEntity(ctx context.Context, u *ent.User) (User, error) {
	rec := event.Get(ctx)

	userObj, err := convertUser(u)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return User{}, err
	}

	rec.Set("success", true)
	return userObj, nil
}

// Users gets all users.
func (s *SESC) Users(ctx context.Context) ([]User, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/users")

	// Stage 1: Query all users
	ctx = rec.Sub("query_all_users").Wrap(ctx)
	res, err := s.queryAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Stage 2: Convert all users
	ctx = rec.Sub("convert_all_users").Wrap(ctx)
	users, err := s.convertAllUsers(ctx, res)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// queryAllUsers queries all users from the database
func (s *SESC) queryAllUsers(ctx context.Context) ([]*ent.User, error) {
	rec := event.Get(ctx)
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	startTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	res, err := s.client.User.Query().WithDepartment().All(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	if err != nil {
		err := fmt.Errorf("couldn't query users: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return nil, err
	}

	rec.Set("success", true)
	return res, nil
}

// convertAllUsers converts all ent.User objects to User domain objects
func (s *SESC) convertAllUsers(ctx context.Context, entUsers []*ent.User) ([]User, error) {
	rec := event.Get(ctx)

	users := make([]User, len(entUsers))
	for i, r := range entUsers {
		var err error
		users[i], err = convertUser(r)
		if err != nil {
			rec.Add(events.Error, err)
			rec.Set("success", false)
			return nil, fmt.Errorf("couldn't convert user: %w", err)
		}
	}

	rec.Set("success", true)
	return users, nil
}

// User returns a User by ID. Alias for UserByID.
// Returns ErrUserNotFound if the user does not exist.
func (s *SESC) User(ctx context.Context, id UUID) (User, error) {
	rec := event.Get(ctx).Sub("sesc/user")

	// Create a wrapped context for UserByID
	ctx = rec.Sub("user_by_id").Wrap(ctx)
	return s.UserByID(ctx, id)
}
