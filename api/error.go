package api

import (
	"github.com/gofrs/uuid/v5"
)

type APIError struct {
	Code      string `json:"code"              example:"DEPARTMENT_NOT_FOUND"`
	Message   string `json:"message"           example:"Department not found"`
	RuMessage string `json:"ruMessage"         example:"Кафедра не найдена"`
	Details   string `json:"details,omitempty" example:"ID: abc123"`
}

type GrantPermissionsRequest struct {
	Permissions []int32 `json:"permissions" example:"[4,5]"`
}

type RevokePermissionsRequest struct {
	Permissions []int32 `json:"permissions" example:"[4,5]"`
}

type SetRoleRequest struct {
	RoleID int32 `json:"roleId" example:"3"`
}

type SetDepartmentRequest struct {
	DepartmentID uuid.UUID `json:"departmentId" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UpdateUserInfoRequest struct {
	FirstName  string `json:"firstName"  example:"Ivan"`
	LastName   string `json:"lastName"   example:"Petrov"`
	MiddleName string `json:"middleName" example:"Sergeevich"`
	PictureURL string `json:"pictureUrl" example:"/images/users/ivan.jpg"`
}

type SetProfilePicRequest struct {
	PictureURL string `json:"pictureUrl" example:"/images/users/new-ivan.jpg"`
}

var (
	ErrInvalidRequest = APIError{
		Code:      "INVALID_REQUEST",
		Message:   "Invalid request body",
		RuMessage: "Некорректный формат запроса",
	}
)
