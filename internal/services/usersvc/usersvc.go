package usersvc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementgroup"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
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

type USS struct {
	client *ent.Client
}

func New(client *ent.Client) *USS {
	return &USS{
		client: client,
	}
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %w", err, rerr)
	}
	return err
}

func convertUser(u *ent.User) User {
	var dept Department
	dep := u.Edges.Department
	if dep != nil {
		dept = Department{
			ID:          dep.ID,
			Name:        dep.Name,
			Description: dep.Description,
		}
	}

	return User{
		ID:                u.ID,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		MiddleName:        u.MiddleName,
		PictureURL:        u.PictureURL,
		Suspended:         u.Suspended,
		Department:        dept,
		Role:              u.Role,
		Subdivision:       u.Subdivision,
		JobTitle:          u.JobTitle,
		EmploymentRate:    u.EmploymentRate,
		AcademicDegree:    u.AcademicDegree,
		PersonnelCategory: u.PersonnelCategory,
		EmploymentType:    u.EmploymentType,
		AcademicTitle:     u.AcademicTitle,
		Honors:            u.Honors,
		Category:          u.Category,
		DateOfEmployment:  u.DateOfEmployment,
		UnemploymentDate:  u.UnemploymentDate,
	}
}

// UpdateUser updates user with the new fields.
//
// Returns an sesc.ErrInvalidRole if the new role id is invalid.
// Returns an sesc.ErrInvalidName if the first or last name is missing.
// Returns an sesc.ErrUserNotFound if the user does not exist.
func (s *USS) UpdateUser(ctx context.Context, id UUID, upd UserUpdateOptions) (User, error) {
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
		return User{}, err
	}

	// Stage 2: Validate role
	ctx = rec.Sub("validate_role").Wrap(ctx)
	if err := sesc.ValidateRole(upd.NewRole); err != nil {
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
	updated := s.convertUserEntity(ctx, us)

	rec.Set("success", true)
	rec.Set("user", updated.EventRecord())
	return updated, nil
}

// validateUserExists validates that a user exists
func (s *USS) validateUserExists(ctx context.Context, id UUID) error {
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
		rec.Add(events.Error, sesc.ErrInvalidDepartment)
		return nil, sesc.ErrInvalidDepartment
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
		SetSubdivision(upd.Subdivision).
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

// convertUserEntity converts an ent.User to a User domain object
func (s *USS) convertUserEntity(
	ctx context.Context,
	us *ent.User,
) User {
	rec := event.Get(ctx)

	updated := convertUser(us)

	rec.Set("success", true)
	return updated
}

// CreateUser creates a new User with a specified role.
//
// Returns an sesc.ErrInvalidName if the first or last name is missing.
func (s *USS) CreateUser(ctx context.Context, opt UserUpdateOptions) (User, error) {
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
	user := s.convertUserEntity(ctx, us)

	rec.Set("success", true)
	rec.Set("user", user.EventRecord())
	return user, nil
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
		SetSubdivision(opt.Subdivision).
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

// UserByID gets a user by their ID.
// Returns an sesc.ErrUserNotFound if the user does not exist.
func (s *USS) UserByID(ctx context.Context, id UUID) (User, error) {
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
	userObj := s.convertUserFromEntity(ctx, u)

	return userObj, nil
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

// convertUserFromEntity converts an ent.User to a User domain object
func (s *USS) convertUserFromEntity(ctx context.Context, u *ent.User) User {
	rec := event.Get(ctx)

	userObj := convertUser(u)

	rec.Set("success", true)
	return userObj
}

// Users gets all users.
func (s *USS) Users(ctx context.Context) ([]User, error) {
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
	users := s.convertAllUsers(ctx, res)

	return users, nil
}

// queryAllUsers queries all users from the database
func (s *USS) queryAllUsers(ctx context.Context) ([]*ent.User, error) {
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
func (s *USS) convertAllUsers(ctx context.Context, entUsers []*ent.User) []User {
	rec := event.Get(ctx)

	users := make([]User, len(entUsers))
	for i, r := range entUsers {
		users[i] = convertUser(r)
	}

	rec.Set("success", true)
	return users
}

// User returns a User by ID. Alias for UserByID.
// Returns sesc.ErrUserNotFound if the user does not exist.
func (s *USS) User(ctx context.Context, id UUID) (User, error) {
	rec := event.Get(ctx).Sub("sesc/user")

	// Create a wrapped context for UserByID
	ctx = rec.Sub("user_by_id").Wrap(ctx)
	return s.UserByID(ctx, id)
}

// AchievementGroups gets all achievement groups with optional filtering.
func (s *USS) AchievementGroups(
	ctx context.Context,
	options AchievementGroupSearchOptions,
) ([]AchievementGroup, error) {
	rec := event.Get(ctx).Sub("sesc/achievement_groups")

	query := s.client.AchievementGroup.Query()

	// Apply filters
	if !options.ShowInactive {
		query = query.Where(achievementgroup.Active(true))
	}

	if options.Search != "" {
		searchTerm := strings.ToLower(options.Search)
		query = query.Where(achievementgroup.Or(
			achievementgroup.NameContainsFold(searchTerm),
			achievementgroup.DescriptionContainsFold(searchTerm),
		))
	}

	groups, err := query.All(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement groups: %w", err))
		return nil, err
	}

	result := make([]AchievementGroup, 0, len(groups))
	for _, g := range groups {
		result = append(result, AchievementGroup{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Active:      g.Active,
		})
	}

	rec.Add("groups_count", len(result))
	return result, nil
}
