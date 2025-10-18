package respond

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
)

func WithUser(u *ent.User) *User {
	return &User{
		ID:                u.ID,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		MiddleName:        u.MiddleName,
		PictureURL:        u.PictureURL,
		Role:              WithRole(u.Role),
		Suspended:         u.Suspended,
		DepartmentID:      u.DepartmentID,
		JobTitle:          u.JobTitle,
		EmploymentRate:    u.EmploymentRate,
		AcademicDegree:    int(u.AcademicDegree),
		PersonnelCategory: int(u.PersonnelCategory),
		EmploymentType:    int(u.EmploymentType),
		AcademicTitle:     u.AcademicTitle,
		Honors:            u.Honors,
		Category:          u.Category,
		DateOfEmployment:  u.DateOfEmployment,
		UnemploymentDate:  u.UnemploymentDate,
	}
}

func WithUsers(uu ent.Users, total int) Users {
	users := make([]*User, len(uu))
	for i, u := range uu {
		users[i] = WithUser(u)
	}
	return Users{
		Users: users,
		Total: total,
	}
}

type User struct {
	ID           uuid.UUID  `json:"id"           example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	FirstName    string     `json:"firstName"    example:"Ivan"                                 validate:"required"`
	LastName     string     `json:"lastName"     example:"Petrov"                               validate:"required"`
	MiddleName   string     `json:"middleName"   example:"Sergeevich"`
	PictureURL   string     `json:"pictureUrl"   example:"/images/users/ivan.jpg"               validate:"required"`
	Role         Role       `json:"role"                                                        validate:"required"`
	Suspended    bool       `json:"suspended"                                                   validate:"required"`
	DepartmentID *uuid.UUID `json:"departmentId"`

	JobTitle          string  `json:"jobTitle"          example:"Профессор"                 validate:"required"`
	EmploymentRate    float64 `json:"employmentRate"    example:"1.0"                       validate:"required"`
	AcademicDegree    int     `json:"academicDegree"    example:"2"`
	PersonnelCategory int     `json:"personnelCategory" example:"1"                         validate:"required"`
	EmploymentType    int     `json:"employmentType"    example:"1"                         validate:"required"`
	AcademicTitle     string  `json:"academicTitle"     example:"Профессор"`
	Honors            string  `json:"honors"            example:"Заслуженный деятель науки"`
	Category          string  `json:"category"          example:"Высшая"`

	DateOfEmployment time.Time `json:"dateOfEmployment,omitzero" example:"2020-01-15T00:00:00Z"`
	UnemploymentDate time.Time `json:"unemploymentDate,omitzero" example:"2023-12-31T00:00:00Z"`
}

type Users struct {
	Users []*User `json:"users" validate:"required"`
	Total int     `json:"total" validate:"required"`
}
