package companyservice

import (
	"context"
	"testing"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testUsers() []company.User {
	return []company.User{
		{ID: "user1", FullName: "Alice Smith", DepartmentID: "dept1", Roles: []company.Role{company.Teacher}},
		{ID: "user2", FullName: "Bob Jones", DepartmentID: "dept1", Roles: []company.Role{company.Dephead}},
		{ID: "user3", FullName: "Charlie Brown", DepartmentID: "dept2", Roles: []company.Role{company.Teacher, company.Admin}},
		{ID: "user4", FullName: "Diana Prince", DepartmentID: "dept2", Roles: []company.Role{company.ScientificDeputy}},
	}
}

func testDepartments() []company.Department {
	return []company.Department{
		{ID: "dept1", Name: "Department One", Description: "First department"},
		{ID: "dept2", Name: "Department Two", Description: "Second department"},
	}
}

func TestStorage_NewStorage(t *testing.T) {
	t.Run("creates storage with users and departments", func(t *testing.T) {
		users := testUsers()
		depts := testDepartments()

		s := newStorage(users, depts)

		require.NotNil(t, s)
		assert.Len(t, s.users, len(users))
		assert.Len(t, s.departments, len(depts))
		assert.Len(t, s.usersByID, len(users))
		assert.Len(t, s.deptsByID, len(depts))
	})

	t.Run("creates storage with nil slices", func(t *testing.T) {
		s := newStorage(nil, nil)

		require.NotNil(t, s)
		assert.Empty(t, s.users)
		assert.Empty(t, s.departments)
		assert.Empty(t, s.usersByID)
		assert.Empty(t, s.deptsByID)
	})

	t.Run("creates storage with empty slices", func(t *testing.T) {
		s := newStorage([]company.User{}, []company.Department{})

		require.NotNil(t, s)
		assert.Empty(t, s.users)
		assert.Empty(t, s.departments)
	})
}

func TestStorage_GetUserByID(t *testing.T) {
	s := newStorage(testUsers(), testDepartments())

	t.Run("returns existing user", func(t *testing.T) {
		user, ok := s.getUserByID("user1")

		assert.True(t, ok)
		assert.Equal(t, "user1", user.ID)
		assert.Equal(t, "Alice Smith", user.FullName)
	})

	t.Run("returns false for non-existing user", func(t *testing.T) {
		user, ok := s.getUserByID("nonexistent")

		assert.False(t, ok)
		assert.Empty(t, user.ID)
	})

	t.Run("returns false for empty ID", func(t *testing.T) {
		user, ok := s.getUserByID("")

		assert.False(t, ok)
		assert.Empty(t, user.ID)
	})
}

func TestStorage_GetDepartmentByID(t *testing.T) {
	s := newStorage(testUsers(), testDepartments())

	t.Run("returns existing department", func(t *testing.T) {
		dept, ok := s.getDepartmentByID("dept1")

		assert.True(t, ok)
		assert.Equal(t, "dept1", dept.ID)
		assert.Equal(t, "Department One", dept.Name)
	})

	t.Run("returns false for non-existing department", func(t *testing.T) {
		dept, ok := s.getDepartmentByID("nonexistent")

		assert.False(t, ok)
		assert.Empty(t, dept.ID)
	})

	t.Run("returns false for empty ID", func(t *testing.T) {
		dept, ok := s.getDepartmentByID("")

		assert.False(t, ok)
		assert.Empty(t, dept.ID)
	})
}

func TestStorage_UsersWithIDs(t *testing.T) {
	s := newStorage(testUsers(), testDepartments())
	ctx := context.Background()

	t.Run("returns users in order of IDs", func(t *testing.T) {
		ids := []string{"user3", "user1", "user4", "user2"}

		users, err := s.usersWithIDs(ctx, ids)

		require.NoError(t, err)
		require.Len(t, users, 4)
		assert.Equal(t, "user3", users[0].ID)
		assert.Equal(t, "user1", users[1].ID)
		assert.Equal(t, "user4", users[2].ID)
		assert.Equal(t, "user2", users[3].ID)
	})

	t.Run("returns ExEmployee for missing users", func(t *testing.T) {
		ids := []string{"user1", "missing1", "user2", "missing2"}

		users, err := s.usersWithIDs(ctx, ids)

		require.NoError(t, err)
		require.Len(t, users, 4)
		assert.Equal(t, "user1", users[0].ID)
		assert.Equal(t, "Alice Smith", users[0].FullName)
		assert.Equal(t, "missing1", users[1].ID)
		assert.Equal(t, "Бывший Сотрудник", users[1].FullName)
		assert.Equal(t, "user2", users[2].ID)
		assert.Equal(t, "Bob Jones", users[2].FullName)
		assert.Equal(t, "missing2", users[3].ID)
		assert.Equal(t, "Бывший Сотрудник", users[3].FullName)
	})

	t.Run("returns all ExEmployees for all missing users", func(t *testing.T) {
		ids := []string{"missing1", "missing2", "missing3"}

		users, err := s.usersWithIDs(ctx, ids)

		require.NoError(t, err)
		require.Len(t, users, 3)
		for i, user := range users {
			assert.Equal(t, ids[i], user.ID)
			assert.Equal(t, "Бывший Сотрудник", user.FullName)
		}
	})

	t.Run("returns empty slice for empty IDs", func(t *testing.T) {
		users, err := s.usersWithIDs(ctx, []string{})

		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("returns empty slice for nil IDs", func(t *testing.T) {
		users, err := s.usersWithIDs(ctx, nil)

		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ids := []string{"user1", "user2"}

		_, err := s.usersWithIDs(ctx, ids)

		assert.Error(t, err)
	})

	t.Run("handles duplicate IDs", func(t *testing.T) {
		ids := []string{"user1", "user1", "user1"}

		users, err := s.usersWithIDs(ctx, ids)

		require.NoError(t, err)
		require.Len(t, users, 3)
		for _, user := range users {
			assert.Equal(t, "user1", user.ID)
			assert.Equal(t, "Alice Smith", user.FullName)
		}
	})
}

func TestStorage_QueryUser(t *testing.T) {
	s := newStorage(testUsers(), testDepartments())
	ctx := context.Background()

	t.Run("returns user by ID without password verification", func(t *testing.T) {
		user, err := s.queryUser(ctx, companyquery.User{ID: "user1"}, nil)

		require.NoError(t, err)
		assert.Equal(t, "user1", user.ID)
		assert.Equal(t, "Alice Smith", user.FullName)
	})

	t.Run("returns error for non-existing user", func(t *testing.T) {
		_, err := s.queryUser(ctx, companyquery.User{ID: "nonexistent"}, nil)

		assert.ErrorIs(t, err, company.ErrUserNotFound)
	})

	t.Run("verifies password when verifyPassword is provided", func(t *testing.T) {
		verifyPassword := func(userID, password string) error {
			if userID == "user1" && password == "correct" {
				return nil
			}
			return company.ErrUserNotFound
		}

		user, err := s.queryUser(ctx, companyquery.User{ID: "user1", Password: "correct"}, verifyPassword)

		require.NoError(t, err)
		assert.Equal(t, "user1", user.ID)
	})

	t.Run("returns error for incorrect password", func(t *testing.T) {
		verifyPassword := func(userID, password string) error {
			return company.ErrUserNotFound
		}

		_, err := s.queryUser(ctx, companyquery.User{ID: "user1", Password: "wrong"}, verifyPassword)

		assert.ErrorIs(t, err, company.ErrUserNotFound)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.queryUser(ctx, companyquery.User{ID: "user1"}, nil)

		assert.Error(t, err)
	})
}

func TestStorage_QueryUsers(t *testing.T) {
	s := newStorage(testUsers(), testDepartments())
	ctx := context.Background()

	t.Run("returns all users for empty query", func(t *testing.T) {
		users, err := s.queryUsers(ctx, companyquery.Users{})

		require.NoError(t, err)
		assert.Len(t, users, 4)
	})

	t.Run("filters by department ID", func(t *testing.T) {
		users, err := s.queryUsers(ctx, companyquery.Users{DepartmentID: "dept1"})

		require.NoError(t, err)
		require.Len(t, users, 2)
		for _, u := range users {
			assert.Equal(t, "dept1", u.DepartmentID)
		}
	})

	t.Run("filters by role ID", func(t *testing.T) {
		users, err := s.queryUsers(ctx, companyquery.Users{RoleID: string(company.Teacher)})

		require.NoError(t, err)
		require.Len(t, users, 2)
		for _, u := range users {
			assert.True(t, u.HasRole(company.Teacher))
		}
	})

	t.Run("filters by full name substring", func(t *testing.T) {
		users, err := s.queryUsers(ctx, companyquery.Users{FullName: "alice"})

		require.NoError(t, err)
		require.Len(t, users, 1)
		assert.Equal(t, "Alice Smith", users[0].FullName)
	})

	t.Run("returns empty slice when no matches", func(t *testing.T) {
		users, err := s.queryUsers(ctx, companyquery.Users{DepartmentID: "nonexistent"})

		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.queryUsers(ctx, companyquery.Users{})

		assert.Error(t, err)
	})
}

func TestStorage_QueryDepartment(t *testing.T) {
	s := newStorage(testUsers(), testDepartments())
	ctx := context.Background()

	t.Run("returns existing department", func(t *testing.T) {
		dept, err := s.queryDepartment(ctx, companyquery.Department{ID: "dept1"})

		require.NoError(t, err)
		assert.Equal(t, "dept1", dept.ID)
		assert.Equal(t, "Department One", dept.Name)
	})

	t.Run("returns error for non-existing department", func(t *testing.T) {
		_, err := s.queryDepartment(ctx, companyquery.Department{ID: "nonexistent"})

		assert.ErrorIs(t, err, company.ErrDepartmentNotFound)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.queryDepartment(ctx, companyquery.Department{ID: "dept1"})

		assert.Error(t, err)
	})
}

func TestStorage_QueryDepartments(t *testing.T) {
	s := newStorage(testUsers(), testDepartments())
	ctx := context.Background()

	t.Run("returns all departments for empty name", func(t *testing.T) {
		depts, err := s.queryDepartments(ctx, companyquery.Departments{Name: ""})

		require.NoError(t, err)
		assert.Len(t, depts, 2)
	})

	t.Run("filters by name substring", func(t *testing.T) {
		depts, err := s.queryDepartments(ctx, companyquery.Departments{Name: "one"})

		require.NoError(t, err)
		require.Len(t, depts, 1)
		assert.Equal(t, "Department One", depts[0].Name)
	})

	t.Run("filters by name case-insensitive", func(t *testing.T) {
		depts, err := s.queryDepartments(ctx, companyquery.Departments{Name: "ONE"})

		require.NoError(t, err)
		require.Len(t, depts, 1)
		assert.Equal(t, "Department One", depts[0].Name)
	})

	t.Run("returns empty slice when no matches", func(t *testing.T) {
		depts, err := s.queryDepartments(ctx, companyquery.Departments{Name: "nonexistent"})

		require.NoError(t, err)
		assert.Empty(t, depts)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.queryDepartments(ctx, companyquery.Departments{Name: ""})

		assert.Error(t, err)
	})
}
