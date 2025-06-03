package iam

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/pkg/apperr"
)

var (
	ErrCredentialsAlreadyExist = apperr.New(
		"пользователь с такими учетными данными уже существует",
		apperr.KindConflict,
	)
	ErrInvalidToken        = apperr.New("некорректный токен", apperr.KindUnauthorized)
	ErrUserNotFound        = apperr.New("пользователь не существует", apperr.KindNotFound)
	ErrEmptyUsername       = apperr.New("имя пользователя не должно быть пустым", apperr.KindValidation)
	ErrEmptyPassword       = apperr.New("пароль не должен быть пустым", apperr.KindValidation)
	ErrCredentialsNotFound = apperr.New("неверное имя пользователя или пароль", apperr.KindNotFound)
	ErrUnauthorized        = apperr.New("требуется войти в систему", apperr.KindUnauthorized)
	ErrForbidden           = apperr.New("недостаточно прав", apperr.KindForbidden)
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
