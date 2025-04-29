package auth

import (
	"errors"

	"github.com/gofrs/uuid/v5"
)

type ID = uuid.UUID

type Credentials struct {
	Username string
	Password string
}

var (
	ErrDuplicateUsername = errors.New("duplicate username")
)
