package db

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDepartmentNotFound = errors.New("department not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidPermission  = errors.New("invalid permission")
)
