package sesc

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

	Role             Role
	ExtraPermissions []Permission
}
