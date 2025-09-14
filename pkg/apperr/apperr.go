package apperr

import "net/http"

type ErrorKind int

var (
	KindNotFound     ErrorKind = http.StatusNotFound
	KindValidation   ErrorKind = http.StatusBadRequest
	KindConflict     ErrorKind = http.StatusConflict
	KindForbidden    ErrorKind = http.StatusForbidden
	KindUnauthorized ErrorKind = http.StatusUnauthorized
)

type AppError struct {
	Message    string
	StatusHint int
}

func New(msg string, kind ErrorKind) *AppError {
	return &AppError{msg, int(kind)}
}

func (ae *AppError) Error() string {
	return ae.Message
}
