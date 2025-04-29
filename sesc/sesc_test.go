//go:generate mockgen -destination=mocks/mock_db.go . DB
//go:generate mockgen -destination=mocks/mock_iam.go . IAM
package sesc_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/auth"
	"github.com/kozlov-ma/sesc-backend/db"
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
		AuthCredentials: auth.Credentials{
			Username: "ivanovi",
			Password: "password",
		},
	}

	dep := sesc.Department{
		ID:   uuid.Must(uuid.NewV7()),
		Name: "Кафедра математики",
	}

	aid := uuid.Must(uuid.NewV7())

	t.Run("simple", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		iam.EXPECT().Register(gomock.Any(), gomock.Eq(uopt.AuthCredentials)).Return(aid, nil)
		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(nil)

		s := sesc.New(log, db, iam)

		u, err := s.CreateTeacher(context.Background(), uopt, dep)

		require.NoError(t, err)
		assert.Equal(t, aid, u.AuthID)
		assert.Equal(t, false, u.Suspended)
		assert.Equal(t, dep, u.Department)
		assert.Equal(t, sesc.Teacher, u.Role)
	})

	t.Run("duplicate_username", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		iam.EXPECT().Register(gomock.Any(), gomock.Eq(uopt.AuthCredentials)).Return(aid, auth.ErrDuplicateUsername)
		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).MaxTimes(0)

		s := sesc.New(log, db, iam)

		_, err := s.CreateTeacher(context.Background(), uopt, dep)
		assert.ErrorIs(t, err, auth.ErrDuplicateUsername)
	})

	t.Run("db_error", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		e := errors.New("ahh db error")

		iam.EXPECT().Register(gomock.Any(), gomock.Eq(uopt.AuthCredentials)).Return(aid, nil)
		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(e)

		s := sesc.New(log, db, iam)

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
		AuthCredentials: auth.Credentials{
			Username: "ivanovi",
			Password: "password",
		},
	}

	aid := uuid.Must(uuid.NewV7())

	t.Run("simple", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		iam.EXPECT().Register(gomock.Any(), gomock.Eq(uopt.AuthCredentials)).Return(aid, nil)
		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(nil)

		s := sesc.New(log, db, iam)

		u, err := s.CreateUser(t.Context(), uopt, sesc.Dephead)

		require.NoError(t, err)

		assert.Equal(t, sesc.Dephead, u.Role)
		assert.Equal(t, aid, u.AuthID)
		assert.Equal(t, false, u.Suspended)
		assert.Equal(t, "Ivan", u.FirstName)
	})

	t.Run("teacher_role", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		s := sesc.New(log, db, iam)

		_, err := s.CreateUser(t.Context(), uopt, sesc.Teacher)
		assert.ErrorIs(t, err, sesc.ErrInvalidRole)
	})

	t.Run("db_error", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		e := errors.New("ahh db error")

		iam.EXPECT().Register(gomock.Any(), gomock.Eq(uopt.AuthCredentials)).Return(aid, nil)
		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(e)

		s := sesc.New(log, db, iam)

		_, err := s.CreateUser(context.Background(), uopt, sesc.Dephead)
		assert.ErrorIs(t, err, e)
	})

	t.Run("duplicate_username", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		iam.EXPECT().Register(gomock.Any(), gomock.Eq(uopt.AuthCredentials)).Return(aid, auth.ErrDuplicateUsername)
		db.EXPECT().SaveUser(gomock.Any(), gomock.Any()).MaxTimes(0)

		s := sesc.New(log, db, iam)

		_, err := s.CreateUser(context.Background(), uopt, sesc.Dephead)
		assert.ErrorIs(t, err, auth.ErrDuplicateUsername)
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
		iam := mock_sesc.NewMockIAM(ctrl)
		db := mock_sesc.NewMockDB(ctrl)

		db.EXPECT().UserByID(gomock.Any(), gomock.Eq(userId)).Return(user, nil)

		s := sesc.New(log, db, iam)

		u, err := s.User(t.Context(), userId)

		require.NoError(t, err)
		assert.Equal(t, user, u)
	})

	t.Run("not_found", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		mdb := mock_sesc.NewMockDB(ctrl)

		mdb.EXPECT().UserByID(gomock.Any(), gomock.Eq(userId)).Return(sesc.User{}, db.ErrNotFound)

		s := sesc.New(log, mdb, iam)

		_, err := s.User(t.Context(), userId)

		assert.ErrorIs(t, err, sesc.ErrUserNotFound)
	})

	t.Run("db_error", func(t *testing.T) {
		iam := mock_sesc.NewMockIAM(ctrl)
		mdb := mock_sesc.NewMockDB(ctrl)

		e := errors.New("db error")
		mdb.EXPECT().UserByID(gomock.Any(), gomock.Eq(userId)).Return(sesc.User{}, e)

		s := sesc.New(log, mdb, iam)

		_, err := s.User(t.Context(), userId)

		assert.ErrorIs(t, err, e)
	})
}
