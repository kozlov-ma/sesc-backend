package achsvc

import (
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/company"
)

// The following access control actions mirror the pattern used in filesvc.
// For now, all AllowsUser methods are permissive stubs that return true.

type CreateAchievementAction struct {
	OwnerID string
}

func NewCreateAchievementAction(ownerID string) CreateAchievementAction {
	return CreateAchievementAction{OwnerID: ownerID}
}

func (a CreateAchievementAction) AllowsUser(u company.User) bool {
	return u.HasRole(company.Teacher) && u.ID == a.OwnerID
}

type ViewAchievementAction struct {
	OwnerID      string
	DepartmentID string
	Status       achievement.Status
	ReviewerRole company.Role
}

func NewViewAchievementAction(
	ownerID string,
	departmentID string,
	status achievement.Status,
	reviewerRole company.Role,
) ViewAchievementAction {
	return ViewAchievementAction{
		OwnerID:      ownerID,
		DepartmentID: departmentID,
		Status:       status,
		ReviewerRole: reviewerRole,
	}
}

func (a ViewAchievementAction) AllowsUser(u company.User) bool {
	if u.HasRole(company.Teacher) {
		return u.ID == a.OwnerID
	}

	if u.HasRole(company.Dephead) {
		return a.Status != achievement.StatusDraft &&
			a.Status != achievement.StatusAccounted &&
			u.DepartmentID == a.DepartmentID
	}

	if u.HasRole(
		company.ScientificDeputy,
		company.OlympiadDeputy,
		company.DevelopmentDeputy,
		company.AcademicDirector,
	) {
		if a.Status == achievement.StatusDraft || a.Status == achievement.StatusDepheadRequestedChanges {
			return false
		}
		return u.HasRole(a.ReviewerRole)
	}

	if u.HasRole(company.ChiefEconomist) {
		return a.Status == achievement.StatusDone || a.Status == achievement.StatusAccounted
	}

	return false
}

type SubmitAchievementAction struct {
	OwnerID   string
	AchStatus achievement.Status
}

func NewSubmitAchievementAction(ownerID string, achStatus achievement.Status) SubmitAchievementAction {
	return SubmitAchievementAction{OwnerID: ownerID, AchStatus: achStatus}
}

func (a SubmitAchievementAction) AllowsUser(u company.User) bool {
	return u.HasRole(company.Teacher) && a.AchStatus == achievement.StatusDraft && u.ID == a.OwnerID
}

type ReviewAchievementAction struct {
	Status       achievement.Status
	ReviewerRole company.Role
	DepartmentID string
}

func NewReviewAchievementAction(
	status achievement.Status,
	reviewerRole company.Role,
	departmentID string,
) ReviewAchievementAction {
	return ReviewAchievementAction{
		Status:       status,
		ReviewerRole: reviewerRole,
		DepartmentID: departmentID,
	}
}

func (a ReviewAchievementAction) AllowsUser(u company.User) bool {
	if a.Status == achievement.StatusDepheadReview {
		return u.HasRole(company.Dephead) &&
			u.DepartmentID != "" &&
			u.DepartmentID == a.DepartmentID
	}

	if a.Status == achievement.StatusInspectorReview {
		return u.HasRole(a.ReviewerRole)
	}

	return false
}

type ModifyAchievementAction struct {
	OwnerID       string
	AchievementID UUID
}

func NewModifyAchievementAction(ownerID string, achievementID UUID) ModifyAchievementAction {
	return ModifyAchievementAction{OwnerID: ownerID, AchievementID: achievementID}
}

func (a ModifyAchievementAction) AllowsUser(_ company.User) bool {
	return true
}

type ListUsersWithAchievementsAction struct {
	AskingUserID string
	DepartmentID string
}

func NewListUsersWithAchievementsAction(askingUserID string, departmentID string) ListUsersWithAchievementsAction {
	return ListUsersWithAchievementsAction{AskingUserID: askingUserID, DepartmentID: departmentID}
}

func (a ListUsersWithAchievementsAction) AllowsUser(u company.User) bool {
	return u.HasRole(company.Dephead) && u.DepartmentID == a.DepartmentID
}

type GenerateUserPointsReportAction struct{}

func NewGenerateUserPointsReportAction() GenerateUserPointsReportAction {
	return GenerateUserPointsReportAction{}
}

func (a GenerateUserPointsReportAction) AllowsUser(u company.User) bool {
	return u.HasRole(company.ChiefEconomist)
}

type AccountingAction struct{}

func NewAccountingAction() AccountingAction {
	return AccountingAction{}
}

func (a AccountingAction) AllowsUser(u company.User) bool {
	return u.HasRole(company.ChiefEconomist)
}
