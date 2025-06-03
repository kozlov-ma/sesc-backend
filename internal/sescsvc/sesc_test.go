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
		NewRole:           1,
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
		require.ErrorIs(t, err, sesc.ErrInvalidDepartmentName)
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
		require.ErrorIs(t, err, sesc.ErrDepartmentNotFound)
	})

	t.Run("non-existent department", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		err := svc.DeleteDepartment(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, sesc.ErrDepartmentNotFound)
	})

	t.Run("with dependent users", func(t *testing.T) {
		ctx, svc, depID := setup(t)

		// Create a user with this department
		opt := validUserUpdateOptions()
		opt.DepartmentID = &depID
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
		require.ErrorIs(t, err, sesc.ErrDepartmentNotFound)
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
		require.ErrorIs(t, err, sesc.ErrDepartmentNotFound)
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
