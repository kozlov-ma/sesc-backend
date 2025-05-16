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

	Subdivision    string
	JobTitle       string
	EmploymentRate float64

	PersonnelCategory PersonnelCategory
	EmploymentType    EmploymentType
	AcademicDegree    AcademicDegree

	AcademicTitle string
	Honors        string
	Category      string

	DateOfEmployment time.Time
	UnemploymentDate time.Time
	CreateDate       time.Time
	UpdateDate       time.Time
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
		"academic_degree", u.AcademicDegree,
		"employment_type", u.EmploymentType,
		"personnel_category", u.PersonnelCategory,
		"academic_title", u.AcademicTitle,
		"honors", u.Honors,
		"category", u.Category,

		"date_of_employment", u.DateOfEmployment,
		"unemployment_date", u.UnemploymentDate,
		"created_at", u.CreateDate,
		"updated_at", u.UpdateDate,
	)
}

func (u User) HasPermission(permission Permission) bool {
	return u.Role.HasPermission(permission)
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
		PersonnelCategory: int(u.PersonnelCategory),
		EmploymentType:    int(u.EmploymentType),
		AcademicDegree:    int(u.AcademicDegree),
		AcademicTitle:     u.AcademicTitle,
		Honors:            u.Honors,
		Category:          u.Category,
		DateOfEmployment:  u.DateOfEmployment,
		UnemploymentDate:  u.UnemploymentDate,
	}
}

type PersonnelCategory int

const (
	ProfessorialPedagogical PersonnelCategory = iota + 1
	Pedagogical
	EducationalSupport
	AdministrativeManagerial
)

type EmploymentType int

const (
	Main EmploymentType = iota + 1
	InternalPartTime
	ExternalPartTime
)

type AcademicDegree int

const (
	NoDegree AcademicDegree = iota
	Candidate
	Doctor
)
