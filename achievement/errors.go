package achievement

import (
	"github.com/kozlov-ma/sesc-backend/pkg/domerr"
)

var (
	ErrInvalidName = domerr.New("некорректное имя", domerr.KindValidation)

	ErrAchievementGroupNotFound    = domerr.New("группа достижений не существует", domerr.KindNotFound)
	ErrAchievementTemplateNotFound = domerr.New("шаблон достижения не существует", domerr.KindNotFound)
	ErrInvalidAchievementKind      = domerr.New("некорректный тип достижения", domerr.KindValidation)

	ErrAchievementNotFound = domerr.New("достижение не существует", domerr.KindNotFound)
	ErrDocumentNotFound    = domerr.New("документ не существует", domerr.KindNotFound)

	ErrInvalidPointsLimit = domerr.New(
		"максимальное количество баллов должно быть положительным",
		domerr.KindValidation,
	)

	ErrWrongAchievementStatus = domerr.New(
		"текущий статус достижения не позволяет работать с ним таким образом",
		domerr.KindConflict,
	)

	ErrNegativePoints = domerr.New(
		"количество баллов не может быть отрицательным",
		domerr.KindValidation,
	)

	ErrPointsLimitExceeded = domerr.New(
		"нельзя оценить достижение выше, чем оно оценено сейчас",
		domerr.KindConflict,
	)
)
