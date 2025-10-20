package company

import (
	"github.com/kozlov-ma/sesc-backend/pkg/apperr"
	"github.com/kozlov-ma/sesc-backend/pkg/domerr"
)

var (
	ErrInvalidRole        = domerr.New("некорректная роль", domerr.KindNotFound)
	ErrUserNotFound       = domerr.New("пользователь не существует", domerr.KindNotFound)
	ErrDepartmentNotFound = domerr.New("кафедра не существует", domerr.KindNotFound)

	ErrForbidden         = apperr.New("вам не разрешено выполнять это действие", apperr.KindForbidden)
	ErrNeedAuthorization = apperr.New(
		"для выполнения этого действия необходимо выполнить вход в систему",
		apperr.KindUnauthorized,
	)
)
