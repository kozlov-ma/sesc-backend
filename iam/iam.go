package iam

import (
	"errors"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrCredentialsAlreadyExist = errors.New("user with similar credentials already exists")
	ErrInvalidToken            = errors.New("invalid token")
	ErrUserNotFound            = errors.New("user not found")
	ErrEmptyUsername           = errors.New("empty username")
	ErrEmptyPassword           = errors.New("empty password")
	ErrInvalidUserID           = errors.New("invalid user ID")
	ErrCredentialsNotFound     = errors.New("credentials not found")
	ErrInvalidRole             = errors.New("invalid role")
	ErrUnauthorized            = errors.New("unauthorized access")
	ErrTokenExpired            = errors.New("token expired")
	ErrInvalidTokenFormat      = errors.New("invalid token format")
	ErrTokenSignature          = errors.New("invalid token signature")
)

type Credentials struct {
	Username string
	Password string
}

type AdminCredentials struct {
	ID UUID
	Credentials
}

func (c Credentials) Validate() error {
	if c.Username == "" {
		return ErrEmptyUsername
	}
	if c.Password == "" {
		return ErrEmptyPassword
	}
	return nil
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Identity struct {
	AuthID uuid.UUID
	Role   Role
	ID     uuid.UUID
}

type UUID = uuid.UUID
