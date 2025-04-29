package sesc

import "errors"

var (
	ErrInvalidRole             = errors.New("invalid role")
	ErrUserNotFound            = errors.New("user not found")
	ErrUsernameTaken           = errors.New("auth username already taken")
	ErrDepartmentAlreadyExists = errors.New("department already exists")
)
