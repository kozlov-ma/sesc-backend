package domerr

import "net/http"

type ErrorKind int

var (
	KindNotFound   ErrorKind = http.StatusNotFound
	KindValidation ErrorKind = http.StatusBadRequest
	KindConflict   ErrorKind = http.StatusConflict
)

type DomainError struct {
	Message    string
	StatusHint int
}

func New(msg string, kind ErrorKind) *DomainError {
	return &DomainError{msg, int(kind)}
}

func (de *DomainError) Error() string {
	return de.Message
}
