package sescsvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// validUserUpdateOptions returns a valid UserUpdateOptions for testing
func validUserUpdateOptions() UserUpdateOptions {
	return UserUpdateOptions{
		FirstName:         "Test",
		LastName:          "User",
		NewRoleID:         1,
		Subdivision:       "Test Subdivision",
		JobTitle:          "Test Position",
		EmploymentRate:    1.0,
		PersonnelCategory: 1,
		EmploymentType:    1,
		DateOfEmployment:  time.Now(),
	}
}

func requireDepartmentMatches(t *testing.T, expected, actual Department) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "Department ID mismatch")
	require.Equal(t, expected.Name, actual.Name, "Department name mismatch")
	require.Equal(t, expected.Description, actual.Description, "Department description mismatch")
}

func requireUserMatches(t *testing.T, expected, actual User) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "User ID mismatch")
	require.Equal(t, expected.FirstName, actual.FirstName, "User FirstName mismatch")
	require.Equal(t, expected.LastName, actual.LastName, "User LastName mismatch")

	// Only check department if expected has one
	if expected.Department.ID != uuid.Nil {
		require.Equal(
			t,
			expected.Department.ID,
			actual.Department.ID,
			"User Department.ID mismatch",
		)
	}

	if expected.Role.ID != 0 {
		require.Equal(t, expected.Role.ID, actual.Role.ID, "User Role.ID mismatch")
	}

	if expected.PictureURL != "" {
		require.Equal(t, expected.PictureURL, actual.PictureURL, "User PictureURL mismatch")
	}
}

func setupSESC(t *testing.T) *SESC {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		_ = client.Close()
	})
	return New(client)
}

func TestCreateDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")

		svc = setupSESC(t)
		return ctx, svc
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc := setup(t)

		name := "HR"
		desc := "Human Resources"

		dep, err := svc.CreateDepartment(ctx, name, desc)
		expected := Department{ID: dep.ID, Name: name, Description: desc}
		require.NoError(t, err, "CreateDepartment failed")
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("duplicate id", func(t *testing.T) {
		ctx, svc := setup(t)

		_, _ = svc.CreateDepartment(ctx, "IT", "IT Dept")
		// Trying to create another department with the same name
		_, err := svc.CreateDepartment(ctx, "IT", "Duplicate Dept")
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestDeleteDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, id UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		dep, _ := svc.CreateDepartment(ctx, "Test", "Test Dept")
		return ctx, svc, dep.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, id := setup(t)

		err := svc.DeleteDepartment(ctx, id)
		require.NoError(t, err, "DeleteDepartment failed")

		_, err = svc.DepartmentByID(ctx, id)
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		err := svc.DeleteDepartment(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})

	t.Run("with dependent users", func(t *testing.T) {
		ctx, svc, depID := setup(t)

		// Create a user with this department
		opt := validUserUpdateOptions()
		opt.DepartmentID = depID
		_, err := svc.CreateUser(ctx, opt)
		require.NoError(t, err)

		err = svc.DeleteDepartment(ctx, depID)
		require.ErrorIs(t, err, sesc.ErrCannotRemoveDepartment)
	})
}

func TestDepartmentByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, id UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		name := "Test"
		desc := "Test Dept"
		dep, _ := svc.CreateDepartment(ctx, name, desc)
		return ctx, svc, dep.ID
	}

	t.Run("existing department", func(t *testing.T) {
		ctx, svc, id := setup(t)

		dep, err := svc.DepartmentByID(ctx, id)
		require.NoError(t, err, "DepartmentByID failed")

		expected := Department{ID: id, Name: "Test", Description: "Test Dept"}
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		_, err := svc.DepartmentByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestGetAllDepartments(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		return ctx, svc
	}

	t.Run("empty", func(t *testing.T) {
		ctx, svc := setup(t)

		deps, err := svc.Departments(ctx)
		require.NoError(t, err, "Departments failed")
		require.Empty(t, deps, "Expected 0 departments")
	})

	t.Run("multiple departments", func(t *testing.T) {
		ctx, svc := setup(t)

		// Create departments
		expectedDeps := make([]Department, 2)
		for i := range expectedDeps {
			name := fmt.Sprintf("Dep %d", i)
			desc := "Desc"
			dep, err := svc.CreateDepartment(ctx, name, desc)
			require.NoError(t, err)
			expectedDeps[i] = dep
		}

		deps, err := svc.Departments(ctx)
		require.NoError(t, err, "Departments failed")
		require.Len(t, deps, len(expectedDeps), "Unexpected number of departments")

		// Verify that each created department exists in the result
		for _, expected := range expectedDeps {
			found := false
			for _, actual := range deps {
				if actual.ID == expected.ID {
					requireDepartmentMatches(t, expected, actual)
					found = true
					break
				}
			}
			require.True(t, found, "Created department not found in results")
		}
	})
}

func TestCreateUser(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, depID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		dep, _ := svc.CreateDepartment(ctx, "Dep", "Dep")
		depID = dep.ID
		return ctx, svc, depID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, depID := setup(t)

		opts := validUserUpdateOptions()
		opts.FirstName = "John"
		opts.LastName = "Doe"
		opts.DepartmentID = depID

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err, "CreateUser failed")

		expected := User{
			ID:         user.ID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: Department{ID: depID},
			Role:       Role{ID: 1},
		}
		requireUserMatches(t, expected, user)

		savedUser, err := svc.UserByID(ctx, user.ID)
		require.NoError(t, err)
		requireUserMatches(t, expected, savedUser)

		us, err := svc.Users(ctx)
		require.NoError(t, err)
		require.Len(t, us, 1)
	})

	t.Run("without_department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		opts := validUserUpdateOptions()
		opts.FirstName = "Jane"
		opts.LastName = "Smith"
		opts.NewRoleID = 2

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err, "CreateUser failed")

		expected := User{
			ID:         user.ID, // Use the ID returned by CreateUser
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: Department{},
			Role:       Role{ID: opts.NewRoleID},
		}
		requireUserMatches(t, expected, user)

		// Verify user is retrievable
		savedUser, err := svc.UserByID(ctx, user.ID)
		require.NoError(t, err)
		requireUserMatches(t, expected, savedUser)
	})

	t.Run("invalid department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		opts := validUserUpdateOptions()
		opts.FirstName = "Jane"
		opts.LastName = "Doe"
		opts.DepartmentID = uuid.Must(uuid.NewV7())

		_, err := svc.CreateUser(ctx, opts)
		require.Error(t, err)
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestUpdateDepartment(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, id UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		dep, _ := svc.CreateDepartment(ctx, "Old", "Old Desc")
		id = dep.ID
		return ctx, svc, id
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, id := setup(t)

		newName, newDesc := "New", "New Desc"
		err := svc.UpdateDepartment(ctx, id, newName, newDesc)
		require.NoError(t, err, "UpdateDepartment failed")

		dep, err := svc.DepartmentByID(ctx, id)
		require.NoError(t, err)

		expected := Department{ID: id, Name: newName, Description: newDesc}
		requireDepartmentMatches(t, expected, dep)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		err := svc.UpdateDepartment(ctx, uuid.Must(uuid.NewV7()), "Name", "Desc")
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})
}

func TestUpdateProfilePicture(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, userID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create a user
		opts := validUserUpdateOptions()
		opts.FirstName = "John"
		opts.LastName = "Doe"

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err)

		return ctx, svc, user.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, userID := setup(t)

		newURL := "http://example.com/new.jpg"
		err := svc.UpdateProfilePicture(ctx, userID, newURL)
		require.NoError(t, err, "UpdateProfilePicture failed")

		user, err := svc.UserByID(ctx, userID)
		require.NoError(t, err)

		expected := User{
			ID:         userID,
			FirstName:  "John",
			LastName:   "Doe",
			Role:       Role{ID: 1},
			PictureURL: newURL,
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		err := svc.UpdateProfilePicture(ctx, uuid.Must(uuid.NewV7()), "url")
		require.ErrorIs(t, err, sesc.ErrUserNotFound)
	})
}

func TestUpdateUser(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, depID UUID, userID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create department
		dep, _ := svc.CreateDepartment(ctx, "Dep", "Dep")
		depID = dep.ID

		// Create user
		opts := validUserUpdateOptions()
		opts.FirstName = "Original"
		opts.LastName = "User"
		opts.DepartmentID = depID

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err)

		return ctx, svc, depID, user.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, depID, userID := setup(t)
		opts := validUserUpdateOptions()
		opts.FirstName = "Updated"
		opts.LastName = "User"
		opts.DepartmentID = depID
		opts.NewRoleID = 2

		user, err := svc.UpdateUser(ctx, userID, opts)
		require.NoError(t, err, "UpdateUser failed")

		expected := User{
			ID:         userID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: Department{ID: depID},
			Role:       Role{ID: opts.NewRoleID},
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, svc, _, _ := setup(t)
		_, err := svc.UpdateUser(ctx, uuid.Must(uuid.NewV7()), validUserUpdateOptions())
		require.ErrorIs(t, err, sesc.ErrUserNotFound)
	})

	t.Run("invalid department", func(t *testing.T) {
		ctx, svc, _, userID := setup(t)
		opts := validUserUpdateOptions()
		opts.FirstName = "Updated"
		opts.LastName = "User"
		opts.DepartmentID = uuid.Must(uuid.NewV7())
		_, err := svc.UpdateUser(ctx, userID, opts)
		require.ErrorIs(t, err, sesc.ErrInvalidDepartment)
	})

	t.Run("remove department", func(t *testing.T) {
		ctx, svc, _, userID := setup(t)
		opts := validUserUpdateOptions()
		opts.FirstName = "Updated"
		opts.LastName = "User"
		opts.NewRoleID = 2
		res, err := svc.UpdateUser(ctx, userID, opts)
		require.NoError(t, err)

		expected := User{
			ID:         userID,
			FirstName:  opts.FirstName,
			LastName:   opts.LastName,
			Department: Department{},
			Role:       Role{ID: opts.NewRoleID},
		}
		requireUserMatches(t, expected, res)
	})

	t.Run("invalid role", func(t *testing.T) {
		ctx, svc, _, userID := setup(t)
		opts := validUserUpdateOptions()
		opts.FirstName = "Updated"
		opts.LastName = "User"
		opts.NewRoleID = 999
		_, err := svc.UpdateUser(ctx, userID, opts)
		require.ErrorIs(t, err, sesc.ErrInvalidRole)
	})
}

func TestUserByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, userID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create user
		opts := validUserUpdateOptions()
		opts.FirstName = "John"
		opts.LastName = "Doe"

		user, err := svc.CreateUser(ctx, opts)
		require.NoError(t, err)

		return ctx, svc, user.ID
	}

	t.Run("existing user", func(t *testing.T) {
		ctx, svc, userID := setup(t)

		user, err := svc.UserByID(ctx, userID)
		require.NoError(t, err, "UserByID failed")

		expected := User{
			ID:        userID,
			FirstName: "John",
			LastName:  "Doe",
			Role:      Role{ID: 1},
		}
		requireUserMatches(t, expected, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		_, err := svc.UserByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, sesc.ErrUserNotFound)
	})
}

func TestGetAllUsers(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create some users
		for i := range 2 {
			opts := validUserUpdateOptions()
			opts.FirstName = fmt.Sprintf("User%d", i+1)
			opts.LastName = fmt.Sprintf("User%d", i+1)
			_, err := svc.CreateUser(ctx, opts)
			require.NoError(t, err)
		}

		return ctx, svc
	}

	t.Run("fetch all users", func(t *testing.T) {
		ctx, svc := setup(t)

		users, err := svc.Users(ctx)
		require.NoError(t, err, "Users failed")
		require.Len(t, users, 2, "Expected 2 users")

		// Verify user fields
		for _, user := range users {
			require.NotEqual(t, uuid.Nil, user.ID, "User ID should not be nil")
			require.NotEmpty(t, user.FirstName, "User FirstName should not be empty")
			require.NotEmpty(t, user.LastName, "User LastName should not be empty")
			require.Equal(t, int32(1), user.Role.ID, "User Role.ID should be 1")
		}
	})
}

func requireAchievementGroupMatches(t *testing.T, expected, actual AchievementGroup) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "AchievementGroup ID mismatch")
	require.Equal(t, expected.Name, actual.Name, "AchievementGroup name mismatch")
	require.Equal(t, expected.Description, actual.Description, "AchievementGroup description mismatch")
	require.Equal(t, expected.Active, actual.Active, "AchievementGroup active mismatch")
}

func requireAchievementTemplateMatches(t *testing.T, expected, actual AchievementTemplate) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "AchievementTemplate ID mismatch")
	require.Equal(t, expected.Name, actual.Name, "AchievementTemplate name mismatch")
	require.Equal(t, expected.Description, actual.Description, "AchievementTemplate description mismatch")
	require.Equal(t, expected.PointsLimit, actual.PointsLimit, "AchievementTemplate points limit mismatch")
	require.Equal(t, expected.GroupID, actual.GroupID, "AchievementTemplate group ID mismatch")
	require.Equal(t, expected.Active, actual.Active, "AchievementTemplate active mismatch")
	require.Equal(t, expected.Kind, actual.Kind, "AchievementTemplate kind mismatch")
}

func TestCreateAchievementGroup(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		return ctx, svc
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc := setup(t)

		opts := AchievementGroupCreateOptions{
			Name:        "Scientific Activities",
			Description: "Research and scientific work",
		}

		group, err := svc.CreateAchievementGroup(ctx, opts)
		require.NoError(t, err, "CreateAchievementGroup failed")

		expected := AchievementGroup{
			ID:          group.ID,
			Name:        opts.Name,
			Description: opts.Description,
			Active:      true,
		}
		requireAchievementGroupMatches(t, expected, group)
	})
}

func TestAchievementGroupByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, groupID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		opts := AchievementGroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}
		group, _ := svc.CreateAchievementGroup(ctx, opts)
		return ctx, svc, group.ID
	}

	t.Run("existing group", func(t *testing.T) {
		ctx, svc, groupID := setup(t)

		group, err := svc.AchievementGroupByID(ctx, groupID)
		require.NoError(t, err, "AchievementGroupByID failed")

		expected := AchievementGroup{
			ID:          groupID,
			Name:        "Test Group",
			Description: "Test Description",
			Active:      true,
		}
		requireAchievementGroupMatches(t, expected, group)
	})

	t.Run("non-existent group", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		_, err := svc.AchievementGroupByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, achievement.ErrAchievementGroupNotFound)
	})
}

func TestAchievementGroups(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		return ctx, svc
	}

	t.Run("empty", func(t *testing.T) {
		ctx, svc := setup(t)

		options := AchievementGroupSearchOptions{}
		groups, err := svc.AchievementGroups(ctx, options)
		require.NoError(t, err, "AchievementGroups failed")
		require.Empty(t, groups, "Expected 0 groups")
	})

	t.Run("multiple groups", func(t *testing.T) {
		ctx, svc := setup(t)

		// Create groups
		expectedGroups := make([]AchievementGroup, 2)
		for i := range expectedGroups {
			opts := AchievementGroupCreateOptions{
				Name:        fmt.Sprintf("Group %d", i+1),
				Description: fmt.Sprintf("Description %d", i+1),
			}
			group, err := svc.CreateAchievementGroup(ctx, opts)
			require.NoError(t, err)
			expectedGroups[i] = group
		}

		options := AchievementGroupSearchOptions{}
		groups, err := svc.AchievementGroups(ctx, options)
		require.NoError(t, err, "AchievementGroups failed")
		require.Len(t, groups, len(expectedGroups), "Unexpected number of groups")

		// Verify that each created group exists in the result
		for _, expected := range expectedGroups {
			found := false
			for _, actual := range groups {
				if actual.ID == expected.ID {
					requireAchievementGroupMatches(t, expected, actual)
					found = true
					break
				}
			}
			require.True(t, found, "Created group not found in results")
		}
	})

	t.Run("search filter", func(t *testing.T) {
		ctx, svc := setup(t)

		// Create groups with different names
		opts1 := AchievementGroupCreateOptions{Name: "Scientific Research", Description: "Science"}
		opts2 := AchievementGroupCreateOptions{Name: "Sports Activities", Description: "Sports"}
		group1, _ := svc.CreateAchievementGroup(ctx, opts1)
		_, _ = svc.CreateAchievementGroup(ctx, opts2)

		// Search for "scientific"
		options := AchievementGroupSearchOptions{Search: "scientific"}
		groups, err := svc.AchievementGroups(ctx, options)
		require.NoError(t, err)
		require.Len(t, groups, 1)
		require.Equal(t, group1.ID, groups[0].ID)
	})
}

func TestUpdateAchievementGroup(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, groupID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		opts := AchievementGroupCreateOptions{
			Name:        "Original Name",
			Description: "Original Description",
		}
		group, _ := svc.CreateAchievementGroup(ctx, opts)
		return ctx, svc, group.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, groupID := setup(t)

		newName := "Updated Name"
		newDesc := "Updated Description"
		newActive := false

		opts := AchievementGroupUpdateOptions{
			Name:        &newName,
			Description: &newDesc,
			Active:      &newActive,
		}

		group, err := svc.UpdateAchievementGroup(ctx, groupID, opts)
		require.NoError(t, err, "UpdateAchievementGroup failed")

		expected := AchievementGroup{
			ID:          groupID,
			Name:        newName,
			Description: newDesc,
			Active:      newActive,
		}
		requireAchievementGroupMatches(t, expected, group)
	})

	t.Run("non-existent group", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		opts := AchievementGroupUpdateOptions{}
		_, err := svc.UpdateAchievementGroup(ctx, uuid.Must(uuid.NewV7()), opts)
		require.ErrorIs(t, err, achievement.ErrAchievementGroupNotFound)
	})
}

func TestCreateAchievementTemplate(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, groupID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		groupOpts := AchievementGroupCreateOptions{
			Name:        "Test Group",
			Description: "Test Description",
		}
		group, _ := svc.CreateAchievementGroup(ctx, groupOpts)
		return ctx, svc, group.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, groupID := setup(t)

		opts := AchievementTemplateCreateOptions{
			Name:        "Publication in Journal",
			Description: "Scientific publication",
			PointsLimit: 10,
			GroupID:     groupID,
			Kind:        achievement.Scientific,
		}

		template, err := svc.CreateAchievementTemplate(ctx, opts)
		require.NoError(t, err, "CreateAchievementTemplate failed")

		expected := AchievementTemplate{
			ID:          template.ID,
			Name:        opts.Name,
			Description: opts.Description,
			PointsLimit: opts.PointsLimit,
			GroupID:     opts.GroupID,
			Active:      true,
			Kind:        opts.Kind,
		}
		requireAchievementTemplateMatches(t, expected, template)
	})

	t.Run("invalid group", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		opts := AchievementTemplateCreateOptions{
			Name:        "Template",
			Description: "Description",
			PointsLimit: 5,
			GroupID:     uuid.Must(uuid.NewV7()),
			Kind:        achievement.Scientific,
		}

		_, err := svc.CreateAchievementTemplate(ctx, opts)
		require.ErrorIs(t, err, achievement.ErrAchievementGroupNotFound)
	})

	t.Run("invalid kind", func(t *testing.T) {
		ctx, svc, groupID := setup(t)

		opts := AchievementTemplateCreateOptions{
			Name:        "Template",
			Description: "Description",
			PointsLimit: 5,
			GroupID:     groupID,
			Kind:        achievement.Kind("invalid"),
		}

		_, err := svc.CreateAchievementTemplate(ctx, opts)
		require.ErrorIs(t, err, achievement.ErrInvalidAchievementKind)
	})
}

func TestAchievementTemplateByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, templateID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create group first
		groupOpts := AchievementGroupCreateOptions{Name: "Group", Description: "Group"}
		group, _ := svc.CreateAchievementGroup(ctx, groupOpts)

		// Create template
		templateOpts := AchievementTemplateCreateOptions{
			Name:        "Test Template",
			Description: "Test Description",
			PointsLimit: 15,
			GroupID:     group.ID,
			Kind:        achievement.Olympiad,
		}
		template, _ := svc.CreateAchievementTemplate(ctx, templateOpts)
		return ctx, svc, template.ID
	}

	t.Run("existing template", func(t *testing.T) {
		ctx, svc, templateID := setup(t)

		template, err := svc.AchievementTemplateByID(ctx, templateID)
		require.NoError(t, err, "AchievementTemplateByID failed")

		require.Equal(t, templateID, template.ID)
		require.Equal(t, "Test Template", template.Name)
		require.Equal(t, "Test Description", template.Description)
		require.Equal(t, 15, template.PointsLimit)
		require.Equal(t, achievement.Olympiad, template.Kind)
		require.True(t, template.Active)
	})

	t.Run("non-existent template", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		_, err := svc.AchievementTemplateByID(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, achievement.ErrAchievementTemplateNotFound)
	})
}

func TestAchievementTemplates(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, groupID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create group
		groupOpts := AchievementGroupCreateOptions{Name: "Group", Description: "Group"}
		group, _ := svc.CreateAchievementGroup(ctx, groupOpts)
		return ctx, svc, group.ID
	}

	t.Run("empty", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		options := AchievementTemplateSearchOptions{}
		templates, err := svc.AchievementTemplates(ctx, options)
		require.NoError(t, err, "AchievementTemplates failed")
		require.Empty(t, templates, "Expected 0 templates")
	})

	t.Run("multiple templates", func(t *testing.T) {
		ctx, svc, groupID := setup(t)

		// Create templates
		expectedTemplates := make([]AchievementTemplate, 2)
		for i := range expectedTemplates {
			opts := AchievementTemplateCreateOptions{
				Name:        fmt.Sprintf("Template %d", i+1),
				Description: fmt.Sprintf("Description %d", i+1),
				PointsLimit: i + 5,
				GroupID:     groupID,
				Kind:        achievement.Scientific,
			}
			template, err := svc.CreateAchievementTemplate(ctx, opts)
			require.NoError(t, err)
			expectedTemplates[i] = template
		}

		options := AchievementTemplateSearchOptions{}
		templates, err := svc.AchievementTemplates(ctx, options)
		require.NoError(t, err, "AchievementTemplates failed")
		require.Len(t, templates, len(expectedTemplates), "Unexpected number of templates")

		// Verify that each created template exists in the result
		for _, expected := range expectedTemplates {
			found := false
			for _, actual := range templates {
				if actual.ID == expected.ID {
					requireAchievementTemplateMatches(t, expected, actual)
					found = true
					break
				}
			}
			require.True(t, found, "Created template not found in results")
		}
	})

	t.Run("search filter", func(t *testing.T) {
		ctx, svc, groupID := setup(t)

		// Create templates with different names
		opts1 := AchievementTemplateCreateOptions{
			Name: "Research Publication", Description: "Research", PointsLimit: 10,
			GroupID: groupID, Kind: achievement.Scientific,
		}
		opts2 := AchievementTemplateCreateOptions{
			Name: "Development Project", Description: "Dev", PointsLimit: 8,
			GroupID: groupID, Kind: achievement.Development,
		}
		template1, _ := svc.CreateAchievementTemplate(ctx, opts1)
		_, _ = svc.CreateAchievementTemplate(ctx, opts2)

		// Search for "research"
		options := AchievementTemplateSearchOptions{Search: "research"}
		templates, err := svc.AchievementTemplates(ctx, options)
		require.NoError(t, err)
		require.Len(t, templates, 1)
		require.Equal(t, template1.ID, templates[0].ID)
	})
}

func TestUpdateAchievementTemplate(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, templateID UUID, groupID UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create group
		groupOpts := AchievementGroupCreateOptions{Name: "Group", Description: "Group"}
		group, _ := svc.CreateAchievementGroup(ctx, groupOpts)

		// Create template
		templateOpts := AchievementTemplateCreateOptions{
			Name:        "Original Template",
			Description: "Original Description",
			PointsLimit: 10,
			GroupID:     group.ID,
			Kind:        achievement.Scientific,
		}
		template, _ := svc.CreateAchievementTemplate(ctx, templateOpts)
		return ctx, svc, template.ID, group.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, templateID, groupID := setup(t)

		newName := "Updated Template"
		newDesc := "Updated Description"
		newPoints := 20
		newActive := false
		newKind := achievement.Development

		opts := AchievementTemplateUpdateOptions{
			Name:        &newName,
			Description: &newDesc,
			PointsLimit: &newPoints,
			Active:      &newActive,
			Kind:        &newKind,
		}

		template, err := svc.UpdateAchievementTemplate(ctx, templateID, opts)
		require.NoError(t, err, "UpdateAchievementTemplate failed")

		expected := AchievementTemplate{
			ID:          templateID,
			Name:        newName,
			Description: newDesc,
			PointsLimit: newPoints,
			GroupID:     groupID,
			Active:      newActive,
			Kind:        newKind,
		}
		requireAchievementTemplateMatches(t, expected, template)
	})

	t.Run("non-existent template", func(t *testing.T) {
		ctx, svc, _, _ := setup(t)

		opts := AchievementTemplateUpdateOptions{}
		_, err := svc.UpdateAchievementTemplate(ctx, uuid.Must(uuid.NewV7()), opts)
		require.ErrorIs(t, err, achievement.ErrAchievementTemplateNotFound)
	})

	t.Run("invalid kind", func(t *testing.T) {
		ctx, svc, templateID, _ := setup(t)

		invalidKind := achievement.Kind("invalid")
		opts := AchievementTemplateUpdateOptions{
			Kind: &invalidKind,
		}

		_, err := svc.UpdateAchievementTemplate(ctx, templateID, opts)
		require.ErrorIs(t, err, achievement.ErrInvalidAchievementKind)
	})
}
