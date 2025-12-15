package achsvc

import (
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
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
	ach *ent.Achievement,
	achTemplate *ent.AchievementTemplate,
) ViewAchievementAction {
	return ViewAchievementAction{
		OwnerID:      ach.OwnerID,
		DepartmentID: ach.DepartmentID,
		Status:       achievement.Status(ach.Status),
		ReviewerRole: achTemplate.ReviewerRole,
	}
}

func (a ViewAchievementAction) AllowsUser(u company.User) bool {
	if a.OwnerID == u.ID {
		return true
	}

	if u.HasRole(company.Dephead) {
		if a.Status != achievement.StatusDraft && a.DepartmentID == u.DepartmentID {
			return true
		}
	}

	if u.HasRole(
		company.ScientificDeputy,
		company.OlympiadDeputy,
		company.DevelopmentDeputy,
		company.AcademicDirector,
	) {
		if a.Status != achievement.StatusDraft &&
			a.Status != achievement.StatusDepheadRequestedChanges &&
			a.Status != achievement.StatusDepheadReview &&
			u.HasRole(a.ReviewerRole) {
			return true
		}
	}

	return false
}

type SubmitAchievementAction struct {
	OwnerID   string
	AchStatus achievement.Status
}

func NewSubmitAchievementAction(ach *ent.Achievement) SubmitAchievementAction {
	return SubmitAchievementAction{OwnerID: ach.OwnerID, AchStatus: achievement.Status(ach.Status)}
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
	ach *ent.Achievement,
	achTemplate *ent.AchievementTemplate,
) ReviewAchievementAction {
	return ReviewAchievementAction{
		Status:       achievement.Status(ach.Status),
		ReviewerRole: achTemplate.ReviewerRole,
		DepartmentID: ach.DepartmentID,
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

type ManageAchievementDocumentsAction struct {
	OwnerID string
	Status  achievement.Status
}

func NewManageAchievementDocumentsAction(ach *ent.Achievement) ManageAchievementDocumentsAction {
	return ManageAchievementDocumentsAction{OwnerID: ach.OwnerID, Status: achievement.Status(ach.Status)}
}

func (a ManageAchievementDocumentsAction) AllowsUser(u company.User) bool {
	return u.ID == a.OwnerID && a.Status == achievement.StatusDraft
}

type DeleteAchievementAction struct {
	OwnerID string
	Status  achievement.Status
}

func NewDeleteAchievementAction(ach *ent.Achievement) DeleteAchievementAction {
	return DeleteAchievementAction{OwnerID: ach.OwnerID, Status: achievement.Status(ach.Status)}
}

func (a DeleteAchievementAction) AllowsUser(u company.User) bool {
	return u.ID == a.OwnerID && a.Status == achievement.StatusDraft
}

type UpdatePointsAction struct {
	OwnerID string
	Status  achievement.Status
}

func NewUpdatePointsAction(ach *ent.Achievement) UpdatePointsAction {
	return UpdatePointsAction{OwnerID: ach.OwnerID, Status: achievement.Status(ach.Status)}
}

func (a UpdatePointsAction) AllowsUser(u company.User) bool {
	return u.ID == a.OwnerID && (a.Status == achievement.StatusDepheadRequestedChanges || a.Status == achievement. StatusInspectorRequestedChanges)
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
