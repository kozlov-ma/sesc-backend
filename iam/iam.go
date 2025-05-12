package iam

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/authuser"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

// JWT secret key (should be in config)
var jwtKey = []byte("dinahu")

type Credentials struct {
	Username string
	Password string
}

func (c Credentials) Validate() error {
	if c.Username == "" || c.Password == "" {
		return ErrInvalidCredentials
	}
	return nil
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Identity struct {
	ID   uuid.UUID
	Role Role
}

// IAM handles authentication using Ent for persistence.
type IAM struct {
	log    *slog.Logger
	client *ent.Client
}

// New creates a new IAM with the given Ent client.
func New(log *slog.Logger, client *ent.Client) *IAM {
	return &IAM{log: log, client: client}
}

type UUID uuid.UUID

// RegisterCredentials stores username/password linked to userID, returns authID.
// Returns ErrUserAlreadyExists if username exists, ErrInvalidCredentials if creds invalid.
func (i *IAM) RegisterCredentials(ctx context.Context, userID UUID, creds Credentials) (UUID, error) {
	if err := creds.Validate(); err != nil {
		return UUID{}, err
	}

	exists, err := i.client.AuthUser.
		Query().
		Where(authuser.UsernameEQ(creds.Username)).
		Exist(ctx)
	if err != nil {
		return UUID{}, err
	}
	if exists {
		return UUID{}, ErrUserAlreadyExists
	}
	authID := uuid.Must(uuid.NewV7())
	_, err = i.client.AuthUser.
		Create().
		SetUsername(creds.Username).
		SetPassword(creds.Password).
		SetAuthID(authID).
		SetUserID(uuid.UUID(userID)).
		Save(ctx)
	if err != nil {
		return UUID{}, err
	}
	return UUID(authID), nil
}

// Login verifies credentials and returns signed JWT token string.
func (i *IAM) Login(ctx context.Context, creds Credentials) (string, error) {
	if err := creds.Validate(); err != nil {
		return "", err
	}
	authRec, err := i.client.AuthUser.
		Query().
		Where(authuser.UsernameEQ(creds.Username)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return "", ErrInvalidCredentials
	} else if err != nil {
		return "", err
	}
	if authRec.Password != creds.Password {
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"auth_id": authRec.AuthID.String(),
		"user_id": authRec.UserID.String(),
		"role":    string(RoleUser),
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return signed, nil
}

// ImWatermelon parses tokenString, returns Identity or ErrInvalidToken.
func (i *IAM) ImWatermelon(ctx context.Context, tokenString string) (Identity, error) {
	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return jwtKey, nil
	})
	if err != nil || !parsed.Valid {
		return Identity{}, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, ErrInvalidToken
	}
	userIDstr, ok1 := claims["user_id"].(string)
	roleStr, ok2 := claims["role"].(string)
	if !ok1 || !ok2 {
		return Identity{}, ErrInvalidToken
	}
	uid, err := uuid.FromString(userIDstr)
	if err != nil {
		return Identity{}, ErrInvalidToken
	}
	return Identity{ID: uid, Role: Role(roleStr)}, nil
}

// Logout is a no-op for stateless JWT.
func (i *IAM) Logout(ctx context.Context, tokenString string) error {
	return nil
}

// DropCredentials deletes credentials by authID; returns ErrUserNotFound if missing.
func (i *IAM) DropCredentials(ctx context.Context, authID UUID) error {
	res, err := i.client.AuthUser.
		Delete().
		Where(authuser.AuthIDEQ(uuid.UUID(authID))).
		Exec(ctx)
	if err != nil {
		return err
	}
	if res == 0 {
		return ErrUserNotFound
	}
	return nil
}
