package slogsink

import (
	"context"
	"log/slog"

	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type SlogSink struct {
	log         *slog.Logger
	middlewares []EventMiddleware
}

type EventMiddleware interface {
	ProcessEvent(r *event.Record)
}

func New(log *slog.Logger, middlewares ...EventMiddleware) *SlogSink {
	return &SlogSink{
		log:         log,
		middlewares: middlewares,
	}
}

func (s *SlogSink) ProcessEvent(rec *event.Record) {
	for _, mw := range s.middlewares {
		mw.ProcessEvent(rec)
	}

	level := slog.LevelInfo
	if e := rec.Value(events.Error); e != nil {
		level = slog.LevelError
	}

	if p := rec.Value("panic"); p != nil {
		level = slog.LevelError
	}

	s.log.Log(context.TODO(), level, "event", slog.Any(rec.EventName(), rec))

	rec.Finish()
}
