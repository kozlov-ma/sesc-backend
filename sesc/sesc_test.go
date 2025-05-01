//go:generate mockgen -destination=mocks/mocks.go . DB
package sesc_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
	mock_sesc "github.com/kozlov-ma/sesc-backend/sesc/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSESC_CreateTeacher(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	log := slog.New(slog.DiscardHandler)

	uopt := sesc.UserOptions{
		FirstName:  "Ivan",
		LastName:   "Ivanov",
		MiddleName: "Ivanovich",
		PictureURL: "https://example.com/avatar.jpg",
	}

	dep := sesc.Department{
		ID:   uuid.Must(uuid.NewV7()),
		Name: "Кафедра математики",
	}

	t.Run("simple", func(t *testing.T) {
		mockdb := mock_sesc.NewMockDB(ctrl)

		mockdb.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(nil)

		s := sesc.New(log, mockdb)

		u, err := s.CreateTeacher(context.Background(), uopt, dep)

		require.NoError(t, err)
		assert.Equal(t, false, u.Suspended)
		assert.Equal(t, dep, u.Department)
		assert.Equal(t, sesc.Teacher, u.Role)
	})

	t.Run("db_error", func(t *testing.T) {
		db := mock_sesc.NewMockDB(ctrl)

		e := errors.New("ahh db error")

		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(e)

		s := sesc.New(log, db)

		_, err := s.CreateTeacher(context.Background(), uopt, dep)
		assert.ErrorIs(t, err, e)
	})
}

func TestSESC_CreateUser(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	log := slog.New(slog.DiscardHandler)

	uopt := sesc.UserOptions{
		FirstName:  "Ivan",
		LastName:   "Ivanov",
		MiddleName: "Ivanovich",
		PictureURL: "https://example.com/avatar.jpg",
	}

	t.Run("simple", func(t *testing.T) {
		db := mock_sesc.NewMockDB(ctrl)

		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(nil)

		s := sesc.New(log, db)

		u, err := s.CreateUser(t.Context(), uopt, sesc.Dephead)

		require.NoError(t, err)

		assert.Equal(t, sesc.Dephead, u.Role)
		assert.Equal(t, false, u.Suspended)
		assert.Equal(t, "Ivan", u.FirstName)
	})

	t.Run("teacher_role", func(t *testing.T) {
		db := mock_sesc.NewMockDB(ctrl)

		s := sesc.New(log, db)

		_, err := s.CreateUser(t.Context(), uopt, sesc.Teacher)
		assert.ErrorIs(t, err, sesc.ErrInvalidRole)
	})

	t.Run("db_error", func(t *testing.T) {
		db := mock_sesc.NewMockDB(ctrl)

		e := errors.New("ahh db error")

		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(e)

		s := sesc.New(log, db)

		_, err := s.CreateUser(context.Background(), uopt, sesc.Dephead)
		assert.ErrorIs(t, err, e)
	})
}

func TestSESC_User(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	log := slog.New(slog.DiscardHandler)

	userId := uuid.Must(uuid.NewV7())

	user := sesc.User{
		ID:        userId,
		FirstName: "Ivan",
	}

	t.Run("simple", func(t *testing.T) {
		db := mock_sesc.NewMockDB(ctrl)

		db.EXPECT().UserByID(gomock.Any(), gomock.Eq(userId)).Return(user, nil)

		s := sesc.New(log, db)

		u, err := s.User(t.Context(), userId)

		require.NoError(t, err)
		assert.Equal(t, user, u)
	})

	t.Run("not_found", func(t *testing.T) {
		mdb := mock_sesc.NewMockDB(ctrl)

		mdb.EXPECT().UserByID(gomock.Any(), gomock.Eq(userId)).Return(sesc.User{}, sesc.ErrUserNotFound)

		s := sesc.New(log, mdb)

		_, err := s.User(t.Context(), userId)

		assert.ErrorIs(t, err, sesc.ErrUserNotFound)
	})

	t.Run("db_error", func(t *testing.T) {
		mdb := mock_sesc.NewMockDB(ctrl)

		e := errors.New("db error")
		mdb.EXPECT().UserByID(gomock.Any(), gomock.Eq(userId)).Return(sesc.User{}, e)

		s := sesc.New(log, mdb)

		_, err := s.User(t.Context(), userId)

		assert.ErrorIs(t, err, e)
	})
}

func TestSESC_CreateDepartment(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	log := slog.New(slog.DiscardHandler)

	dep := sesc.Department{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "Кафедра математики",
		Description: "Самая пацанская кафедра",
	}

	t.Run("simple", func(t *testing.T) {
		db := mock_sesc.NewMockDB(ctrl)
		db.EXPECT().CreateDepartment(gomock.Any(), gomock.Any(), dep.Name, dep.Description).Return(dep, nil)

		s := sesc.New(log, db)

		d, err := s.CreateDepartment(context.Background(), dep.Name, dep.Description)

		require.NoError(t, err)
		assert.Equal(t, dep, d)
	})

	t.Run("error department already exists", func(t *testing.T) {
		mockdb := mock_sesc.NewMockDB(ctrl)
		mockdb.EXPECT().
			CreateDepartment(gomock.Any(), gomock.Any(), dep.Name, dep.Description).
			Return(sesc.NoDepartment, sesc.ErrInvalidDepartment)

		s := sesc.New(log, mockdb)

		d, err := s.CreateDepartment(context.Background(), dep.Name, dep.Description)

		require.Error(t, err)
		assert.ErrorIs(t, err, sesc.ErrInvalidDepartment)
		assert.Equal(t, sesc.NoDepartment, d)
	})

	t.Run("strange error", func(t *testing.T) {
		mockdb := mock_sesc.NewMockDB(ctrl)
		mockdb.EXPECT().
			CreateDepartment(gomock.Any(), gomock.Any(), dep.Name, dep.Description).
			Return(sesc.NoDepartment, fmt.Errorf("dinahu"))

		s := sesc.New(log, mockdb)

		d, err := s.CreateDepartment(context.Background(), dep.Name, dep.Description)

		require.Error(t, err)
		assert.Equal(t, sesc.NoDepartment, d)
	})
}

func TestSESC_GrantExtraPermissions(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	log := slog.New(slog.DiscardHandler)

	user := sesc.User{
		ID: uuid.Must(uuid.NewV7()),
	}

	perm1 := sesc.Permissions[0]
	perm2 := sesc.Permissions[1]

	t.Run("simple", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		updatedUser := user
		updatedUser.ExtraPermissions = []sesc.Permission{perm1, perm2}

		dbMock.EXPECT().
			GrantExtraPermissions(gomock.Any(), user, perm1, perm2).
			Return(updatedUser, nil)

		s := sesc.New(log, dbMock)

		u, err := s.GrantExtraPermissions(context.Background(), user, perm1, perm2)
		require.NoError(t, err)
		assert.Equal(t, []sesc.Permission{perm1, perm2}, u.ExtraPermissions)
	})

	t.Run("user_not_found", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		dbMock.EXPECT().
			GrantExtraPermissions(gomock.Any(), user, perm1).
			Return(sesc.User{}, sesc.ErrUserNotFound)

		s := sesc.New(log, dbMock)

		_, err := s.GrantExtraPermissions(context.Background(), user, perm1)
		assert.ErrorIs(t, err, sesc.ErrUserNotFound)
	})

	t.Run("invalid_permission", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		dbMock.EXPECT().
			GrantExtraPermissions(gomock.Any(), user, perm1).
			Return(sesc.User{}, sesc.ErrInvalidPermission)

		s := sesc.New(log, dbMock)

		_, err := s.GrantExtraPermissions(context.Background(), user, perm1)
		assert.ErrorIs(t, err, sesc.ErrInvalidPermission)
	})

	t.Run("db_error", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		e := errors.New("some db error")

		dbMock.EXPECT().
			GrantExtraPermissions(gomock.Any(), user, perm1).
			Return(sesc.User{}, e)

		s := sesc.New(log, dbMock)

		_, err := s.GrantExtraPermissions(context.Background(), user, perm1)
		assert.ErrorIs(t, err, e)
	})
}

func TestSESC_RevokeExtraPermissions(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	log := slog.New(slog.DiscardHandler)

	user := sesc.User{
		ID: uuid.Must(uuid.NewV7()),
		ExtraPermissions: []sesc.Permission{
			sesc.Permissions[0],
		},
	}

	perm1 := sesc.Permissions[0]
	perm2 := sesc.Permissions[1]
	t.Run("simple", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		updatedUser := user
		updatedUser.ExtraPermissions = nil

		dbMock.EXPECT().
			RevokeExtraPermissions(gomock.Any(), user, perm1).
			Return(updatedUser, nil)

		s := sesc.New(log, dbMock)

		u, err := s.RevokeExtraPermissions(context.Background(), user, perm1)
		require.NoError(t, err)
		assert.Empty(t, u.ExtraPermissions)
	})

	t.Run("user_not_found", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		dbMock.EXPECT().
			RevokeExtraPermissions(gomock.Any(), user, perm2).
			Return(sesc.User{}, sesc.ErrUserNotFound)

		s := sesc.New(log, dbMock)

		_, err := s.RevokeExtraPermissions(context.Background(), user, perm2)
		assert.ErrorIs(t, err, sesc.ErrUserNotFound)
	})

	t.Run("invalid_permission", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		dbMock.EXPECT().
			RevokeExtraPermissions(gomock.Any(), user, perm2).
			Return(sesc.User{}, sesc.ErrInvalidPermission)

		s := sesc.New(log, dbMock)

		_, err := s.RevokeExtraPermissions(context.Background(), user, perm2)
		assert.ErrorIs(t, err, sesc.ErrInvalidPermission)
	})

	t.Run("db_error", func(t *testing.T) {
		dbMock := mock_sesc.NewMockDB(ctrl)

		e := errors.New("some db error")

		dbMock.EXPECT().
			RevokeExtraPermissions(gomock.Any(), user, perm1).
			Return(sesc.User{}, e)

		s := sesc.New(log, dbMock)

		_, err := s.RevokeExtraPermissions(context.Background(), user, perm1)
		assert.ErrorIs(t, err, e)
	})
}
