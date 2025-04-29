package sesc

import (
	"github.com/gofrs/uuid/v5"
)

// A role is a standartized set of Permissions granted to a User influenced
// by their role in the organization.
//
// Roles are predefined in this file.
type Role struct {
	ID          UUID
	Name        string
	Permissions []Permission
}

func (r Role) HasPermission(p Permission) bool {
	return r.HasPermissionWithID(p.ID)
}

func (r Role) HasPermissionWithID(id UUID) bool {
	for _, p := range r.Permissions {
		if p.ID == id {
			return true
		}
	}
	return false
}

var (
	Teacher Role = Role{
		ID:   uuid.Must(uuid.NewV7()),
		Name: "Преподаватель",
		Permissions: []Permission{
			PermissionDraftAchievementList,
		},
	}
	Dephead Role = Role{
		ID:   uuid.Must(uuid.NewV7()),
		Name: "Заведующий кафедрой",
		Permissions: []Permission{
			PermissionDraftAchievementList,
			PermissionDepheadReview,
		},
	}
	ContestDeputy Role = Role{
		ID:   uuid.Must(uuid.NewV7()),
		Name: "Заместитель директора по олимпиадной работе",
		Permissions: []Permission{
			PermissionDraftAchievementList,
			PermissionDepheadReview,
			PermissionScientificReview,
		},
	}
	ScientificDeputy Role = Role{
		ID:   uuid.Must(uuid.NewV7()),
		Name: "Заместитель директора по научной работе",
		Permissions: []Permission{
			PermissionDraftAchievementList,
			PermissionDepheadReview,
			PermissionScientificReview,
		},
	}
	DevelopmentDeputy Role = Role{
		ID:   uuid.Must(uuid.NewV7()),
		Name: "Заместитель директора по развитию",
		Permissions: []Permission{
			PermissionDraftAchievementList,
			PermissionDepheadReview,
			PermissionScientificReview,
		},
	}
)
