package filesvc

import (
	"github.com/kozlov-ma/sesc-backend/company"
)

type CreateFileAction struct {
	OwnerID string
}

func NewCreateFileAction(ownerID string) CreateFileAction {
	return CreateFileAction{OwnerID: ownerID}
}

func (a CreateFileAction) AllowsUser(u company.User) bool {
	if u.ID != a.OwnerID {
		return false
	}

	if u.HasRole(
		company.Teacher,
		company.Dephead,
		company.AcademicDirector,
		company.OlympiadDeputy,
		company.DevelopmentDeputy,
		company.ScientificDeputy) {
		return true
	}

	return false
}

type CreateCommonFileAction struct {
}

func NewCommonCreateFileAction() CreateCommonFileAction {
	return CreateCommonFileAction{}
}

func (a CreateCommonFileAction) AllowsUser(u company.User) bool {
	return u.HasRole(
		company.Admin,
		company.ChiefEconomist,
		company.DevelopmentDeputy,
		company.OlympiadDeputy,
		company.ScientificDeputy,
		company.Dephead,
		company.AcademicDirector)
}

type ViewFileAction struct {
	OwnerID *string
}

func NewViewFileAction(ownerID *string) ViewFileAction {
	return ViewFileAction{OwnerID: ownerID}
}

func (a ViewFileAction) AllowsUser(u company.User) bool {
	if u.HasRole(company.Admin) {
		return true
	}

	// Common files (no owner)
	if a.OwnerID == nil {
		return true
	}

	// User's own files
	if *a.OwnerID == u.ID {
		return true
	}

	return false
}

type DeleteFileAction struct {
	OwnerID *string
}

func NewDeleteFileAction(ownerID *string) DeleteFileAction {
	return DeleteFileAction{OwnerID: ownerID}
}

func (a DeleteFileAction) AllowsUser(u company.User) bool {
	if u.HasRole(company.Admin) {
		return true
	}

	// User can delete their own files
	if a.OwnerID != nil && *a.OwnerID == u.ID {
		return true
	}

	return false
}

type SearchFileAction struct {
	OwnerID *string
}

func NewSearchFileAction(ownerID *string) SearchFileAction {
	return SearchFileAction{OwnerID: ownerID}
}

func (a SearchFileAction) AllowsUser(u company.User) bool {
	if u.HasRole(company.Admin) {
		return true
	}

	// User's own files
	if a.OwnerID != nil && *a.OwnerID == u.ID {
		return true
	}

	// Common files (no owner)
	if a.OwnerID == nil {
		return true
	}

	return false
}
