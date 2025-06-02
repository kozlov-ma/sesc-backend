package tests

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api"
	"github.com/kozlov-ma/sesc-backend/api/respond"
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
type User = respond.User

// CreateUserRequest is used to create a new user
type CreateUserRequest = api.CreateUserRequest

// PatchUserRequest is used to update a user
type PatchUserRequest struct {
	FirstName    *string    `json:"firstName,omitempty"`
	LastName     *string    `json:"lastName,omitempty"`
	MiddleName   *string    `json:"middleName,omitempty"`
	PictureURL   *string    `json:"pictureUrl,omitempty"`
	Suspended    *bool      `json:"suspended,omitempty"`
	DepartmentID *uuid.UUID `json:"departmentId,omitempty"`
	Role         *int       `json:"roleId,omitempty"`

	Subdivision       *string  `json:"subdivision,omitempty"`
	JobTitle          *string  `json:"jobTitle,omitempty"`
	EmploymentRate    *float64 `json:"employmentRate,omitempty"`
	AcademicDegree    *int     `json:"academicDegree,omitempty"`
	PersonnelCategory *int     `json:"personnelCategory,omitempty"`
	EmploymentType    *int     `json:"employmentType,omitempty"`
	AcademicTitle     *string  `json:"academicTitle,omitempty"`
	Honors            *string  `json:"honors,omitempty"`
	Category          *string  `json:"category,omitempty"`

	DateOfEmployment *string `json:"dateOfEmployment,omitempty"`
	UnemploymentDate *string `json:"unemploymentDate,omitempty"`
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
type Role = respond.Role

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
