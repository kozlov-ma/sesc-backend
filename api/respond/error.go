package respond

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/pkg/apperr"
	"github.com/kozlov-ma/sesc-backend/pkg/domerr"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type ErrCode = string

type Error struct {
	Code       ErrCode   `json:"code"`
	TraceID    uuid.UUID `json:"traceId,omitzero"`
	Message    string    `json:"message,omitzero"`
	StatusCode int       `json:"statusCode"`
}

func (e *Error) AsHTTPStatusCode() int {
	return e.StatusCode
}

func WithError(ctx context.Context, e error) *Error {
	var trID uuid.UUID
	{
		rec := event.Root(ctx)
		if rec != nil {
			ti := rec.Value(events.TraceID)
			trID, _ = ti.(uuid.UUID)
		}
	}

	pe := asParamError(e)
	de := asDomainError(e)
	ae := asAppError(e)
	switch {
	case pe != nil:
		return &Error{
			Code:       CodeInvalidParam,
			TraceID:    trID,
			Message:    fmt.Sprintf("параметр %q имеет некорректное значение, или не был предоставлен", pe.ParamName),
			StatusCode: http.StatusBadRequest,
		}
	case de != nil:
		return &Error{
			Code:       CodeDomainError,
			TraceID:    trID,
			Message:    de.Message,
			StatusCode: de.StatusHint,
		}
	case ae != nil:
		return &Error{
			Code:       CodeApplicationError,
			TraceID:    trID,
			Message:    ae.Message,
			StatusCode: ae.StatusHint,
		}
	default:
		return &Error{
			Code:       CodeUnknownError,
			TraceID:    trID,
			Message:    "произошла неизвестная ошибка",
			StatusCode: http.StatusInternalServerError,
		}
	}
}

func asParamError(e error) *param.InvalidParamError {
	var pe *param.InvalidParamError
	errors.As(e, &pe)
	return pe
}

func asDomainError(e error) *domerr.DomainError {
	var de *domerr.DomainError
	errors.As(e, &de)
	return de
}

func asAppError(e error) *apperr.AppError {
	var ae *apperr.AppError
	errors.As(e, &ae)
	return ae
}
