package iam

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

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

type IAM struct {
	// log *slog.Logger
}

type UUID uuid.UUID

func (i *IAM) Login(ctx context.Context, creds Credentials) (jwt.Token, error) {
	panic("Not implemented")
}

// RegisterCredentials assigns credentials to a user and returns their auth UUID.
//
// If the user already exists, ErrUserAlreadyExists is returned.
// If the credentials are invalid, like username or password being empty, ErrInvalidCredentials is returned.
func (i *IAM) RegisterCredentials(ctx context.Context, userID UUID, creds Credentials) (UUID, error) {
	panic("Not implemented")
}

// ValidateToken validates a JWT token and returns the identity of the user.
//
// If the token is invalid, ErrInvalidToken is returned.
func (i *IAM) ValidateToken(ctx context.Context, token jwt.Token) (Identity, error) {
	panic("Not implemented")
}

// DropCredentials deletes the credentials associated with the given auth ID,
// and deletes the association of the user with the auth ID.
//
// Returns ErrUserNotFound if the user is not found.
func (i *IAM) DropCredentials(ctx context.Context, authID UUID) error {
	panic("Not implemented")
}

// Logout invalidates the given token. If the token is already invalidated, it's a no-op.
//
// If the token itself is invalid, ErrInvalidToken is returned.
func (i *IAM) Logout(ctx context.Context, token jwt.Token) error {
	panic("Not implemented")
}
