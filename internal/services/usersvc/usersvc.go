package usersvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/department"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	UUID                             = uuid.UUID
	Role                             = sesc.Role
	UserUpdateOptions                = sesc.UserUpdateOptions
	AchievementGroupCreateOptions    = achievement.GroupCreateOptions
	AchievementGroupUpdateOptions    = achievement.GroupUpdateOptions
	AchievementGroupSearchOptions    = achievement.GroupSearchOptions
	AchievementTemplateCreateOptions = achievement.TemplateCreateOptions
	AchievementTemplateUpdateOptions = achievement.TemplateUpdateOptions
	AchievementTemplateSearchOptions = achievement.TemplateSearchOptions
)

type USS struct {
	client *ent.Client
}

func New(client *ent.Client) *USS {
	return &USS{
		client: client,
	}
}

// UpdateUser updates user with the new fields.
//
// Returns an sesc.ErrInvalidRole if the new role id is invalid.
// Returns an sesc.ErrInvalidName if the first or last name is missing.
// Returns an sesc.ErrUserNotFound if the user does not exist.
func (s *USS) UpdateUser(ctx context.Context, id UUID, upd UserUpdateOptions) (*ent.User, error) {
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
		"new_role_id", upd.NewRole,
	)

	// Stage 1: Validate user exists
	ctx = rec.Sub("validate_user_exists").Wrap(ctx)
	if err := s.validateUserExists(ctx, id); err != nil {
		return nil, err
	}

	// Stage 2: Validate role
	ctx = rec.Sub("validate_role").Wrap(ctx)
	if err := sesc.ValidateRole(upd.NewRole); err != nil {
		return nil, err
	}

	// Stage 3: Validate name
	ctx = rec.Sub("validate_name").Wrap(ctx)
	if err := s.validateName(ctx, upd.FirstName, upd.LastName); err != nil {
		return nil, err
	}

	var us *ent.User
	txStart := time.Now()

	err := txwrapper.WithTx(ctx, s.client, sql.LevelSerializable, rec, func(tx *ent.Tx) error {
		// Stage 4: Check and get department if needed
		ctx := rec.Sub("check_department").Wrap(ctx)
		dept, err := s.checkAndGetDepartment(ctx, statrec, tx, upd.DepartmentID)
		if err != nil {
			return err
		}
		// Stage 5: Update user
		ctx = rec.Sub("update_user_record").Wrap(ctx)
		if err := s.updateUserRecord(ctx, statrec, tx, id, upd, dept); err != nil {
			return err
		}

		// Stage 6: Query updated user
		ctx = rec.Sub("query_updated_user").Wrap(ctx)
		us, err = s.queryUpdatedUser(ctx, statrec, tx, id)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))
	rec.Set("success", true)
	rec.Set("user", us)
	return us, nil
}

// validateUserExists validates that a user exists
func (s *USS) validateUserExists(ctx context.Context, id UUID) error {
	rec := event.Get(ctx)
	rec.Set("user_id", id)

	_, err := s.User(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("exists", false)
		return err
	}

	rec.Set("exists", true)
	return nil
}

// validateName validates that the name is not empty
func (s *USS) validateName(ctx context.Context, firstName, lastName string) error {
	rec := event.Get(ctx)
	rec.Set(
		"first_name", firstName,
		"last_name", lastName,
	)

	if firstName == "" || lastName == "" {
		rec.Set("valid", false)
		return sesc.ErrInvalidUserName
	}

	rec.Set("valid", true)
	return nil
}

// checkAndGetDepartment checks if the department exists and returns it
func (s *USS) checkAndGetDepartment(
	ctx context.Context,
	statrec *event.Record,
	tx *ent.Tx,
	departmentID *UUID,
) (*ent.Department, error) {
	rec := event.Get(ctx)
	rec.Set("department_id", departmentID)

	if departmentID == nil {
		rec.Set("required", false)
		//nolint:nilnil // department should be deleted or should not be set.
		return nil, nil
	}

	rec.Set("required", true)
	statrec.Add(events.PostgresQueries, 1)
	// todo uncomment if replace sqlite with postgres(linked with locks in department service)
	dept, err := tx.Department.Query().
		Where(department.ID(*departmentID)).
		// ForShare().
		Only(ctx)
	switch {
	case ent.IsNotFound(err):
		rec.Set("exists", false)
		rec.Add(events.Error, sesc.ErrDepartmentNotFound)
		return nil, sesc.ErrDepartmentNotFound
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
func (s *USS) updateUserRecord(
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
		SetRole(upd.NewRole).
		SetJobTitle(upd.JobTitle).
		SetEmploymentRate(upd.EmploymentRate).
		SetPersonnelCategory(upd.PersonnelCategory).
		SetEmploymentType(upd.EmploymentType).
		SetDateOfEmployment(upd.DateOfEmployment)

	updater = updater.SetAcademicDegree(upd.AcademicDegree).
		SetAcademicTitle(upd.AcademicTitle).
		SetHonors(upd.Honors).
		SetCategory(upd.Category).
		SetUnemploymentDate(upd.UnemploymentDate)

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
func (s *USS) queryUpdatedUser(
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

// CreateUser creates a new User with a specified role.
//
// Returns an sesc.ErrInvalidName if the first or last name is missing.
func (s *USS) CreateUser(ctx context.Context, opt UserUpdateOptions) (*ent.User, error) {
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
		"new_role_id", opt.NewRole,
	)

	// Stage 1: Validate input
	ctx = rec.Sub("validate_create_input").Wrap(ctx)
	if err := s.validateCreateInput(ctx, opt); err != nil {
		return nil, err
	}
	var us *ent.User

	txStart := time.Now()
	err := txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		// Stage 2: Check and get department if needed
		ctx := rec.Sub("check_department").Wrap(ctx)
		dept, err := s.checkAndGetDepartment(ctx, statrec, tx, opt.DepartmentID)
		if err != nil {
			return err
		}

		// Stage 3: Create user record
		ctx = rec.Sub("create_user_record").Wrap(ctx)
		userID, err := s.createUserRecord(ctx, statrec, tx, opt, dept)
		if err != nil {
			return err
		}

		// Stage 4: Query created user
		ctx = rec.Sub("query_created_user").Wrap(ctx)
		us, err = s.queryCreatedUser(ctx, statrec, tx, userID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return nil, err
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	rec.Set("success", true)
	rec.Set("user", us)
	return us, nil
}

// validateCreateInput validates the create user input
func (s *USS) validateCreateInput(ctx context.Context, opt UserUpdateOptions) error {
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
func (s *USS) createUserRecord(
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
		SetRole(opt.NewRole).
		SetJobTitle(opt.JobTitle).
		SetEmploymentRate(opt.EmploymentRate).
		SetPersonnelCategory(opt.PersonnelCategory).
		SetEmploymentType(opt.EmploymentType).
		SetDateOfEmployment(opt.DateOfEmployment)

	cr = cr.SetAcademicDegree(opt.AcademicDegree).
		SetAcademicTitle(opt.AcademicTitle).
		SetHonors(opt.Honors).
		SetCategory(opt.Category).
		SetUnemploymentDate(opt.UnemploymentDate)
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
func (s *USS) queryCreatedUser(
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
// Returns an sesc.ErrUserNotFound if the user does not exist.
func (s *USS) UpdateProfilePicture(ctx context.Context, id UUID, pictureURL string) error {
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
func (s *USS) updateProfilePictureRecord(ctx context.Context, id UUID, pictureURL string) error {
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
		joinedErr := fmt.Errorf("%w: %w", err, sesc.ErrUserNotFound)
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

// User gets a user by their ID.
// Returns an sesc.ErrUserNotFound if the user does not exist.
func (s *USS) User(ctx context.Context, id UUID) (*ent.User, error) {
	// Caller should create the record and use Wrap to add it to the context
	rec := event.Get(ctx).Sub("sesc/user_by_id")

	rec.Sub("params").Set("id", id)

	// Stage 1: Query user by ID
	ctx = rec.Sub("query_user_by_id").Wrap(ctx)
	u, err := s.getUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// getUserByID queries a user by ID from the database
func (s *USS) getUserByID(ctx context.Context, id UUID) (*ent.User, error) {
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
		rec.Add(events.Error, sesc.ErrUserNotFound)
		rec.Set("success", false)
		return nil, sesc.ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't query user: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return nil, err
	}

	rec.Set("success", true)
	return u, nil
}
