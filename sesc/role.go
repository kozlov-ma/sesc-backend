package sesc

// A role is a standartized set of Permissions granted to a User influenced
// by their role in the organization.
//
// Roles are predefined in this file.
type Role struct {
	ID          int32
	Name        string
	Permissions []Permission
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
	Teacher Role = Role{
		ID:   1,
		Name: "Преподаватель",
		Permissions: []Permission{
			PermissionDraftAchievementList,
		},
	}
	Dephead Role = Role{
		ID:   2,
		Name: "Заведующий кафедрой",
		Permissions: []Permission{
			PermissionDepheadReview,
		},
	}
	ContestDeputy Role = Role{
		ID:   3,
		Name: "Заместитель директора по олимпиадной работе",
		Permissions: []Permission{
			PermissionContestReview,
		},
	}
	ScientificDeputy Role = Role{
		ID:   4,
		Name: "Заместитель директора по научной работе",
		Permissions: []Permission{
			PermissionScientificReview,
		},
	}
	DevelopmentDeputy Role = Role{
		ID:   5,
		Name: "Заместитель директора по развитию",
		Permissions: []Permission{
			PermissionDevelopmentReview,
		},
	}
)
