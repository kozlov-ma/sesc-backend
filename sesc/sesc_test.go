//go:generate mockgen -destination=mocks/mock_db.go . DB
//go:generate mockgen -destination=mocks/mock_iam.go . IAM
package sesc_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/auth"
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
		if err != nil {
			t.Fatal(err)
		}

		require.NoError(t, err)
		assert.Equal(t, aid, u.AuthID)
		assert.Equal(t, false, u.Suspended)
		assert.Equal(t, dep, u.Department)
		assert.Equal(t, sesc.Teacher, u.Role)
	})
}
