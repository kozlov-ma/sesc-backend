package sesc

import "errors"

var (
	ErrInvalidRole            = errors.New("invalid role")
	ErrUserNotFound           = errors.New("user not found")
	ErrCannotRemoveDepartment = errors.New("cannot remove department")
	ErrInvalidDepartment      = errors.New("invalid department")
	ErrInvalidPermission      = errors.New("invalid permission")
	ErrInvalidRoleChange      = errors.New("invalid role change")
	ErrInvalidUserName        = errors.New("invalid or missing user name")
	ErrInvalidDepartmentName  = errors.New("invalid or missing department name")
	ErrEmptyDepartment        = errors.New("department is empty")
	ErrDepartmentNotFound     = errors.New("department not found")
	ErrInvalidUserID          = errors.New("invalid user ID")
	ErrInvalidDepartmentID    = errors.New("invalid department ID")
)
