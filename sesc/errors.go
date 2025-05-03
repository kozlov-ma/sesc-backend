package sesc

import "errors"

var (
	ErrInvalidRole            = errors.New("invalid role")
	ErrUserNotFound           = errors.New("user not found")
	ErrCannotRemoveDepartment = errors.New("cannot remove department")
	ErrInvalidDepartment      = errors.New("invalid department")
	ErrInvalidPermission      = errors.New("invalid permission")
	ErrInvalidRoleChange      = errors.New("invalid role change")
	ErrInvalidName            = errors.New("invalid name (first or last name missing)")
)
