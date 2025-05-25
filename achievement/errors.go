package achievement

import "errors"

var (
	ErrAchievementGroupNotFound    = errors.New("achievement group not found")
	ErrAchievementTemplateNotFound = errors.New("achievement template not found")
	ErrInvalidAchievementKind      = errors.New("invalid achievement kind")

	ErrAchievementNotFound = errors.New("achievement not found")
	ErrDocumentNotFound    = errors.New("document not found")

	ErrWrongAchievementStatus = errors.New("achievement status does not allow this kind of change")
	ErrPointsLimitExceeded    = errors.New("points assigned exceed the template's points limit")
)
