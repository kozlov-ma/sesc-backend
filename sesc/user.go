package sesc

import (
	"time"
)

// UserUpdateOptions represents the options for updating a user.
type UserUpdateOptions struct {
	FirstName         string
	LastName          string
	MiddleName        string
	PictureURL        string
	Suspended         bool
	DepartmentID      *UUID
	NewRole           Role
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

	if err := ValidateRole(u.NewRole); err != nil {
		return err
	}

	if u.EmploymentRate < 0 || u.EmploymentRate > 1 {
		return ErrInvalidEmploymentRate
	}

	return nil
}
