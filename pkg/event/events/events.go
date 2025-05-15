// Package events declares reusable keys for event record values.
package events

import (
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

const (
	// Error key should be used to add unexpected error values to the event.
	Error = "error"

	// PostgresTime is cumulative time spent in postgres to execure the event.
	PostgresTime = "postgres_time"

	// PostgresQueries is cumulative number of postgres queries triggered by the event.
	PostgresQueries = "postgres_queries"
)

type DepartmentRecorder sesc.Department

func (dr DepartmentRecorder) EventRecord() *event.Record {
	return event.Group(
		"id", dr.ID,
		"name", dr.Name,
		"description", dr.Description,
	)
}

type RoleRecorder sesc.Role

func (rr RoleRecorder) EventRecord() *event.Record {
	return event.Group(
		"id", rr.ID,
		"name", rr.Name,
	)
}

type UserRecorder sesc.User

func (ur UserRecorder) EventRecord() *event.Record {
	return event.Group(
		"id", ur.ID,
		"first_name", ur.FirstName,
		"suspended", ur.Suspended,
		"department_id", ur.Department.ID,
		"department", DepartmentRecorder(ur.Department).EventRecord(),
		"role_id", ur.Role.ID,
		"role", RoleRecorder(ur.Role),
	)
}
