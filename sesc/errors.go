package sesc

import "errors"

var (
	ErrInvalidRole        = errors.New("invalid role")
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("auth username already taken")
	ErrInvalidDepartment  = errors.New("invalid department")
	ErrInvalidPermission  = errors.New("invalid permission")
	ErrInvalidRoleChange  = errors.New("invalid role change")
	ErrDepartmentRequired = errors.New("department is required")
)
