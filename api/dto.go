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

type CreateDepartmentRequest struct {
	Name        string `json:"name"        example:"Mathematics"`
	Description string `json:"description" example:"Math department"`
}

type CreateDepartmentResponse struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name"        example:"Mathematics"`
	Description string    `json:"description" example:"Math department"`
}

type DepartmentsResponse struct {
	Departments []Department `json:"departments"`
}

type Department struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name"        example:"Mathematics"`
	Description string    `json:"description" example:"Math department"`
}

type RolesResponse struct {
	Roles []Role `json:"roles"`
}

type Role struct {
	ID          int32        `json:"id"          example:"1"`
	Name        string       `json:"name"        example:"Преподаватель"`
	Permissions []Permission `json:"permissions"`
}

type PermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	ID          int32  `json:"id"          example:"1"`
	Name        string `json:"name"        example:"draft_achievement_list"`
	Description string `json:"description" example:"Создание и заполнение листа достижений"`
}

// User Management
type CreateTeacherRequest struct {
	FirstName    string    `json:"firstName"    example:"Ivan"`
	LastName     string    `json:"lastName"     example:"Petrov"`
	MiddleName   string    `json:"middleName"   example:"Sergeevich"`
	PictureURL   string    `json:"pictureUrl"   example:"/images/users/ivan.jpg"`
	DepartmentID uuid.UUID `json:"departmentId" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UserResponse struct {
	ID         uuid.UUID  `json:"id"                  example:"550e8400-e29b-41d4-a716-446655440000"`
	FirstName  string     `json:"firstName"           example:"Ivan"`
	LastName   string     `json:"lastName"            example:"Petrov"`
	MiddleName string     `json:"middleName"          example:"Sergeevich"`
	PictureURL string     `json:"pictureUrl"          example:"/images/users/ivan.jpg"`
	Role       Role       `json:"role"`
	Department Department `json:"department,omitzero"`
}

type CreateUserRequest struct {
	FirstName  string `json:"firstName"  example:"Anna"`
	LastName   string `json:"lastName"   example:"Smirnova"`
	MiddleName string `json:"middleName" example:"Olegovna"`
	PictureURL string `json:"pictureUrl" example:"/images/users/anna.jpg"`
	RoleID     int32  `json:"roleId"     example:"2"`
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
