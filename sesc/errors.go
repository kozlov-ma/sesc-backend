package sesc

import (
	"github.com/kozlov-ma/sesc-backend/pkg/apperr"
	"github.com/kozlov-ma/sesc-backend/pkg/domerr"
)

var (
	ErrInvalidRole            = domerr.New("некорректная роль", domerr.KindNotFound)
	ErrUserNotFound           = domerr.New("пользователь не существует", domerr.KindNotFound)
	ErrCannotRemoveDepartment = domerr.New(
		"невозможно удалить кафедру, к которой привязаны пользователи",
		domerr.KindConflict,
	)
	ErrDepartmentNotFound    = domerr.New("кафедра не существует", domerr.KindNotFound)
	ErrInvalidUserName       = domerr.New("имя пользователя заполнено некорректно", domerr.KindValidation)
	ErrInvalidDepartmentName = domerr.New("название кафедры некорректно", domerr.KindValidation)

	ErrInvalidFileName = domerr.New("некорректное имя файла", domerr.KindValidation)
	ErrInvalidFileSize = domerr.New("некорректный размер файла", domerr.KindValidation)
	ErrFileNotFound    = domerr.New("файл не существует", domerr.KindNotFound)

	ErrInvalidEmploymentRate = domerr.New("некорректная доля ставки", domerr.KindValidation)

	ErrSuspended = apperr.New("действие вашего аккаунта приостановлено", apperr.KindForbidden)
)
