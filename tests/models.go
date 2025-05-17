package tests

import (
	"github.com/gofrs/uuid/v5"
)

// API request/response models for use in tests

// LoginRequest is used for authentication
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse contains the JWT token from a successful login
type LoginResponse struct {
	Token string `json:"token"`
}

// User represents a user in the system
type User struct {
	ID         uuid.UUID  `json:"id"`
	FirstName  string     `json:"firstName"`
	LastName   string     `json:"lastName"`
	MiddleName string     `json:"middleName,omitempty"`
	PictureURL string     `json:"pictureUrl"`
	Role       Role       `json:"role"`
	Suspended  bool       `json:"suspended"`
	Department Department `json:"department,omitempty"`
}

// CreateUserRequest is used to create a new user
type CreateUserRequest struct {
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	MiddleName   string    `json:"middleName,omitempty"`
	RoleID       int32     `json:"roleId"`
	PictureURL   string    `json:"pictureUrl,omitempty"`
	DepartmentID uuid.UUID `json:"departmentId,omitempty"`
}

// PatchUserRequest is used to update a user
type PatchUserRequest struct {
	FirstName    *string    `json:"firstName,omitempty"`
	LastName     *string    `json:"lastName,omitempty"`
	MiddleName   *string    `json:"middleName,omitempty"`
	PictureURL   *string    `json:"pictureUrl,omitempty"`
	Suspended    *bool      `json:"suspended,omitempty"`
	DepartmentID *uuid.UUID `json:"departmentId,omitempty"`
	RoleID       *int32     `json:"roleId,omitempty"`
}

// RegisterUserRequest is used to set credentials for a user
type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Department represents a department in the system
type Department struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// CreateDepartmentRequest is used to create a new department
type CreateDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateDepartmentRequest is used to update a department
type UpdateDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Role represents a role in the system
type Role struct {
	ID          int32        `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Error represents an API error
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RuMessage string `json:"ruMessage,omitempty"`
	Details   string `json:"details,omitempty"`
}
