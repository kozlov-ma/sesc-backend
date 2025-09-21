package accountingperiod

import (
	"github.com/kozlov-ma/sesc-backend/pkg/domerr"
)

var (
	ErrAccountingPeriodNotFound      = domerr.New("учетный период не найден", domerr.KindNotFound)
	ErrInvalidAccountingPeriodName   = domerr.New("название учетного периода некорректно", domerr.KindValidation)
	ErrInvalidAccountingPeriodStatus = domerr.New("статус учетного периода некорректный", domerr.KindValidation)
	ErrInvalidStatusTransition       = domerr.New("недопустимый переход статуса", domerr.KindValidation)
	ErrPlanningPeriodAlreadyExists   = domerr.New("планируемый период уже существует", domerr.KindConflict)
	ErrActivePeriodAlreadyExists     = domerr.New("активный период уже существует", domerr.KindConflict)
	ErrPeriodCannotBeModified        = domerr.New("период нельзя изменить в текущем статусе", domerr.KindConflict)
	ErrPeriodAlreadyFinished         = domerr.New("период уже завершен", domerr.KindConflict)
	ErrPeriodAlreadyCancelled        = domerr.New("период уже отменен", domerr.KindConflict)
	ErrPeriodAlreadyNotExecuted      = domerr.New("период уже помечен как неисполненный", domerr.KindConflict)
)
