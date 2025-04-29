package sesc

import "errors"

var (
	ErrInvalidRole  = errors.New("invalid role")
	ErrUserNotFound = errors.New("user not found")
)
