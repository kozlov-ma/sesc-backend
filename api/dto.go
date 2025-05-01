package api

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
)

type UUID = uuid.UUID

// swagger:model
type APIError struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	RuMessage string `json:"ru_message"`
}

var InternalServerError = APIError{
	Status:    http.StatusInternalServerError,
	Message:   "Internal server error",
	RuMessage: "Внутренняя ошибка сервера",
}

// swagger:model
type CreateDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// swagger:model
type CreateDepartmentError struct {
	APIError
}

// swagger:model
type CreateDepartmentResponse struct {
	ID          UUID                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Error       CreateDepartmentError `json:"error,omitzero"`
}

// swagger:model
type DepartmentsError struct {
	APIError
}

// swagger:model
type DepartmentsResponse struct {
	Error       DepartmentsError `json:"error,omitzero"`
	Departments []Department     `json:"departments"`
}

// swagger:model
type Department struct {
	ID          UUID   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// swagger:model
type Permission struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// swagger:model
type PermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// swagger:model
type Role struct {
	ID          int32        `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

// swagger:model
type RolesResponse struct {
	Error RolesError `json:"error,omitzero"`
	Roles []Role     `json:"roles"`
}

// swagger:model
type RolesError struct {
	APIError
}
