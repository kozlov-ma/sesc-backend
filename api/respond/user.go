package respond

import (
	"github.com/kozlov-ma/sesc-backend/company"
)

func WithUser(u company.User) *User {
	return &User{
		ID:                u.ID,
		FullName:          u.FullName,
		PictureURL:        u.PictureURL,
		Role:              WithRole(u.Role),
		DepartmentID:      u.DepartmentID,
		JobTitle:          u.Extras.JobTitle,
		EmploymentRate:    u.Extras.EmploymentRate,
		AcademicDegree:    u.Extras.AcademicDegree,
		PersonnelCategory: u.Extras.PersonnelCategory,
		EmploymentType:    u.Extras.EmploymentType,
		AcademicTitle:     u.Extras.AcademicTitle,
		Honors:            u.Extras.Honors,
		Category:          u.Extras.Category,
		DateOfEmployment:  u.Extras.DateOfEmployment,
	}
}

func WithUsers(uu []company.User, total int) Users {
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
	ID           string `json:"id"           example:"ivanivanov"             validate:"required"`
	FullName     string `json:"fullName"     example:"Ivanov Ivan Ivanovich"  validate:"required"`
	PictureURL   string `json:"pictureUrl"   example:"/images/users/ivan.jpg" validate:"required"`
	Role         Role   `json:"role"                                          validate:"required"`
	DepartmentID string `json:"departmentId"`

	JobTitle          string `json:"jobTitle"          example:"Профессор"                               validate:"required"`
	EmploymentRate    string `json:"employmentRate"    example:"1.0"                                     validate:"required"`
	AcademicDegree    string `json:"academicDegree"    example:"Доктор наук"`
	PersonnelCategory string `json:"personnelCategory" example:"Административно-управленческий персонал" validate:"required"`
	EmploymentType    string `json:"employmentType"    example:"Внешнее совместительство"                validate:"required"`
	AcademicTitle     string `json:"academicTitle"     example:"Профессор"`
	Honors            string `json:"honors"            example:"Заслуженный деятель науки"`
	Category          string `json:"category"          example:"Высшая"`

	DateOfEmployment string `json:"dateOfEmployment,omitzero" example:"24.02.2022"`
}

type Users struct {
	Users []*User `json:"users" validate:"required"`
	Total int     `json:"total" validate:"required"`
}
