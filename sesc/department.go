package sesc

import "github.com/kozlov-ma/sesc-backend/pkg/event"

// Department represents a department within an organization, like maths, physics, etc.
// A Department can have a head, which is a user who is responsible for managing the department
// and the initial review of the achievement lists of the subordinates.
//
// Departments will be filled by the system administrator.
type Department struct {
	ID          UUID
	Name        string
	Description string
}

func (d Department) EventRecord() *event.Record {
	return event.Group(
		"id", d.ID,
		"name", d.Name,
		"description", d.Description,
	)
}

var (
	NoDepartment = Department{}
)
