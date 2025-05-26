package sesc

import (
	"time"

	"github.com/kozlov-ma/sesc-backend/pkg/event"
)

// User represents a SESC employee that participates in the achievement list
// filling and review processes.
//
// If the User is a teacher, they must be assigned to a Department.
//
// User's Role determines what they can do within the system.
//
// Use ExtraPermissions to grant additional permissions to the user, i.e.,
// the ability to fill out achievement lists as a department head.
type User struct {
	ID UUID

	FirstName  string
	LastName   string
	MiddleName string

	PictureURL string

	Suspended bool

	Department Department

	Role Role

	Subdivision       string
	JobTitle          string
	EmploymentRate    float64
	AcademicDegree    AcademicDegree
	PersonnelCategory PersonnelCategory
	EmploymentType    EmploymentType
	AcademicTitle     string
	Honors            string
	Category          string

	DateOfEmployment time.Time
	UnemploymentDate time.Time
}

func (u User) EventRecord() *event.Record {
	return event.Group(
		"id", u.ID,
		"first_name", u.FirstName,
		"suspended", u.Suspended,
		"department_id", u.Department.ID,
		"department", u.Department,
		"role_id", u.Role.ID,
		"role", u.Role,
		"subdivision", u.Subdivision,
		"job_title", u.JobTitle,
		"employment_rate", u.EmploymentRate,
		"personnel_category", int(u.PersonnelCategory),
		"employment_type", int(u.EmploymentType),
		"academic_degree", int(u.AcademicDegree),
		"date_of_employment", u.DateOfEmployment,
	)
}

func (u User) HasPermission(permission Permission) bool {
	return u.Role.HasPermission(permission)
}

// UserUpdateOptions represents the options for updating a user.
type UserUpdateOptions struct {
	FirstName         string
	LastName          string
	MiddleName        string
	PictureURL        string
	Suspended         bool
	DepartmentID      UUID
	NewRoleID         int32
	Subdivision       string
	JobTitle          string
	EmploymentRate    float64
	AcademicDegree    AcademicDegree
	PersonnelCategory PersonnelCategory
	EmploymentType    EmploymentType
	AcademicTitle     string
	Honors            string
	Category          string
	DateOfEmployment  time.Time
	UnemploymentDate  time.Time
}

func (u UserUpdateOptions) Validate() error {
	if u.FirstName == "" || u.LastName == "" {
		return ErrInvalidUserName
	}

	if _, ok := RoleByID(u.NewRoleID); !ok {
		return ErrInvalidRole
	}

	if u.EmploymentRate < 0 || u.EmploymentRate > 1 {
		return ErrInvalidEmploymentRate
	}

	return nil
}

func (u User) UpdateOptions() UserUpdateOptions {
	return UserUpdateOptions{
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		MiddleName:        u.MiddleName,
		PictureURL:        u.PictureURL,
		Suspended:         u.Suspended,
		DepartmentID:      u.Department.ID,
		NewRoleID:         u.Role.ID,
		Subdivision:       u.Subdivision,
		JobTitle:          u.JobTitle,
		EmploymentRate:    u.EmploymentRate,
		AcademicDegree:    u.AcademicDegree,
		PersonnelCategory: u.PersonnelCategory,
		EmploymentType:    u.EmploymentType,
		AcademicTitle:     u.AcademicTitle,
		Honors:            u.Honors,
		Category:          u.Category,
		DateOfEmployment:  u.DateOfEmployment,
		UnemploymentDate:  u.UnemploymentDate,
	}
}
