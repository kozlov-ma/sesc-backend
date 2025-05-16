//nolint:mnd // the only magic numbers here are ids
package sesc

import "github.com/kozlov-ma/sesc-backend/pkg/event"

// Role is a standartized set of Permissions granted to a User influenced
// by their role in the organization.
//
// Roles are predefined in this file.
type Role struct {
	ID          int32
	Name        string
	Permissions []Permission
}

func (r Role) EventRecord() *event.Record {
	return event.Group(
		"id", r.ID,
		"name", r.Name,
	)
}

func (r Role) HasPermission(p Permission) bool {
	return r.HasPermissionWithID(p.ID)
}

func (r Role) HasPermissionWithID(id int32) bool {
	for _, p := range r.Permissions {
		if p.ID == id {
			return true
		}
	}
	return false
}

var (
	Teacher = Role{
		ID:   1,
		Name: "Преподаватель",
		Permissions: []Permission{
			PermissionDraftAchievementList,
		},
	}
	Dephead = Role{
		ID:   2,
		Name: "Заведующий кафедрой",
		Permissions: []Permission{
			PermissionDepheadReview,
		},
	}
	ContestDeputy = Role{
		ID:   3,
		Name: "Заместитель директора по олимпиадной работе",
		Permissions: []Permission{
			PermissionContestReview,
		},
	}
	ScientificDeputy = Role{
		ID:   4,
		Name: "Заместитель директора по научной работе",
		Permissions: []Permission{
			PermissionScientificReview,
		},
	}
	DevelopmentDeputy = Role{
		ID:   5,
		Name: "Заместитель директора по развитию",
		Permissions: []Permission{
			PermissionDevelopmentReview,
		},
	}
)

var Roles = []Role{
	Teacher,
	Dephead,
	ContestDeputy,
	ScientificDeputy,
	DevelopmentDeputy,
}

func RoleByID(id int32) (Role, bool) {
	for _, r := range Roles {
		if r.ID == id {
			return r, true
		}
	}
	return Role{}, false
}
