package iam

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func setupIAM(t *testing.T) *IAM {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		_ = client.Close()
	})

	return New(
		slog.New(slog.DiscardHandler),
		client,
		time.Hour,
		[]string{"admin-token"},
	)
}

func createTestUser(ctx context.Context, t *testing.T, client *ent.Client) uuid.UUID {
	t.Helper()
	userID := uuid.Must(uuid.NewV7())
	user, err := client.User.Create().
		SetID(userID).
		SetFirstName("Test").
		SetLastName("User").
		SetRoleID(1).
		Save(ctx)
	require.NoError(t, err)
	return user.ID
}

func TestRegisterCredentials(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, iam *IAM, userID uuid.UUID) {
		ctx = t.Context()
		iam = setupIAM(t)
		userID = createTestUser(ctx, t, iam.client)
		return ctx, iam, userID
	}

	t.Run("success", func(t *testing.T) {
		ctx, iam, userID := setup(t)

		creds := Credentials{
			Username: "testuser",
			Password: "password123",
		}

		authID, err := iam.RegisterCredentials(ctx, userID, creds)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, authID)

		savedCreds, err := iam.Credentials(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, creds.Username, savedCreds.Username)
		require.Equal(t, creds.Password, savedCreds.Password)
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		ctx, iam, userID := setup(t)

		_, err := iam.RegisterCredentials(ctx, userID, Credentials{})
		require.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("non_existent_user", func(t *testing.T) {
		ctx, iam, _ := setup(t)

		nonExistentID := uuid.Must(uuid.NewV7())
		_, err := iam.RegisterCredentials(ctx, nonExistentID, Credentials{
			Username: "testuser2",
			Password: "password123",
		})
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("duplicate_username", func(t *testing.T) {
		ctx, iam, userID := setup(t)

		creds := Credentials{
			Username: "duplicate_user",
			Password: "password123",
		}

		_, err := iam.RegisterCredentials(ctx, userID, creds)
		require.NoError(t, err)

		anotherUserID := createTestUser(ctx, t, iam.client)

		_, err = iam.RegisterCredentials(ctx, anotherUserID, creds)
		require.ErrorIs(t, err, ErrUserAlreadyExists)
	})
}

func TestLogin(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, iam *IAM, creds Credentials) {
		ctx = t.Context()
		iam = setupIAM(t)
		userID := createTestUser(ctx, t, iam.client)
		creds = Credentials{
			Username: "logintest",
			Password: "password123",
		}
		_, err := iam.RegisterCredentials(ctx, userID, creds)
		require.NoError(t, err)
		return ctx, iam, creds
	}

	t.Run("success", func(t *testing.T) {
		ctx, iam, creds := setup(t)

		token, err := iam.Login(ctx, creds)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		identity, err := iam.ImWatermelon(ctx, token)
		require.NoError(t, err)
		require.Equal(t, RoleUser, identity.Role)
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		ctx, iam, _ := setup(t)

		_, err := iam.Login(ctx, Credentials{
			Username: "logintest",
			Password: "wrongpassword",
		})
		require.ErrorIs(t, err, ErrUserNotFound)

		_, err = iam.Login(ctx, Credentials{
			Username: "nonexistent",
			Password: "password123",
		})
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestLoginAdmin(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, iam *IAM) {
		ctx = t.Context()
		iam = setupIAM(t)
		return ctx, iam
	}

	t.Run("success", func(t *testing.T) {
		ctx, iam := setup(t)

		token, err := iam.LoginAdmin(ctx, "admin-token")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		identity, err := iam.ImWatermelon(ctx, token)
		require.NoError(t, err)
		require.Equal(t, RoleAdmin, identity.Role)
	})

	t.Run("invalid_token", func(t *testing.T) {
		ctx, iam := setup(t)

		_, err := iam.LoginAdmin(ctx, "invalid-token")
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestDropCredentials(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, iam *IAM, userID uuid.UUID) {
		ctx = t.Context()
		iam = setupIAM(t)
		userID = createTestUser(ctx, t, iam.client)
		creds := Credentials{
			Username: "droptest",
			Password: "password123",
		}
		_, err := iam.RegisterCredentials(ctx, userID, creds)
		require.NoError(t, err)
		return ctx, iam, userID
	}

	t.Run("success", func(t *testing.T) {
		ctx, iam, userID := setup(t)

		err := iam.DropCredentials(ctx, userID)
		require.NoError(t, err)

		_, err = iam.Credentials(ctx, userID)
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("non_existent_user", func(t *testing.T) {
		ctx, iam, _ := setup(t)

		nonExistentID := uuid.Must(uuid.NewV7())
		err := iam.DropCredentials(ctx, nonExistentID)
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("no_credentials", func(t *testing.T) {
		ctx, iam, _ := setup(t)

		userID := createTestUser(ctx, t, iam.client)

		err := iam.DropCredentials(ctx, userID)
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestImWatermelon(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, iam *IAM, userID uuid.UUID, token string) {
		ctx = t.Context()
		iam = setupIAM(t)
		userID = createTestUser(ctx, t, iam.client)
		creds := Credentials{
			Username: "watermelon",
			Password: "password123",
		}
		_, err := iam.RegisterCredentials(ctx, userID, creds)
		require.NoError(t, err)
		token, err = iam.Login(ctx, creds)
		require.NoError(t, err)
		return ctx, iam, userID, token
	}

	t.Run("success", func(t *testing.T) {
		ctx, iam, userID, token := setup(t)

		identity, err := iam.ImWatermelon(ctx, token)
		require.NoError(t, err)
		require.Equal(t, userID, identity.ID)
		require.Equal(t, RoleUser, identity.Role)
	})

	t.Run("invalid_token", func(t *testing.T) {
		ctx, iam, _, _ := setup(t)

		_, err := iam.ImWatermelon(ctx, "invalid-token")
		require.ErrorIs(t, err, ErrInvalidToken)
	})
}

func TestCredentials(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, iam *IAM, userID uuid.UUID, originalCreds Credentials) {
		ctx = t.Context()
		iam = setupIAM(t)
		userID = createTestUser(ctx, t, iam.client)
		originalCreds = Credentials{
			Username: "getcreds",
			Password: "password123",
		}
		_, err := iam.RegisterCredentials(ctx, userID, originalCreds)
		require.NoError(t, err)
		return ctx, iam, userID, originalCreds
	}

	t.Run("success", func(t *testing.T) {
		ctx, iam, userID, originalCreds := setup(t)

		creds, err := iam.Credentials(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, originalCreds.Username, creds.Username)
		require.Equal(t, originalCreds.Password, creds.Password)
	})

	t.Run("non_existent_user", func(t *testing.T) {
		ctx, iam, _, _ := setup(t)

		nonExistentID := uuid.Must(uuid.NewV7())
		_, err := iam.Credentials(ctx, nonExistentID)
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("no_credentials", func(t *testing.T) {
		ctx, iam, _, _ := setup(t)

		userID := createTestUser(ctx, t, iam.client)

		_, err := iam.Credentials(ctx, userID)
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}
