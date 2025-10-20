package auth

import "github.com/kozlov-ma/sesc-backend/pkg/apperr"

var ErrInvalidToken = apperr.New("некорректный токен", apperr.KindUnauthorized)
