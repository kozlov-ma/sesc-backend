package deletiondaemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kozlov-ma/sesc-backend/internal/services/achsvc"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type ObjectStorage interface {
	RemoveObject(ctx context.Context, objectKey string) error
}

type EventMiddleware interface {
	ProcessEvent(r *event.Record)
}

type Service struct {
	achService *achsvc.ACS
	storage    ObjectStorage
	delay      time.Duration
	enabled    bool
	interval   time.Duration
	eventSink  EventMiddleware

	mu           sync.Mutex
	stopChan     chan struct{}
	isRunning    bool
	autoRestart  bool
	restartCount int
	maxRestarts  int
}

func New(
	achService *achsvc.ACS,
	storage ObjectStorage,
	delay time.Duration,
	enabled bool,
	interval time.Duration,
	eventSink EventMiddleware,
) *Service {
	return &Service{
		achService:   achService,
		storage:      storage,
		delay:        delay,
		enabled:      enabled,
		interval:     interval,
		eventSink:    eventSink,
		stopChan:     make(chan struct{}),
		isRunning:    false,
		autoRestart:  true,
		maxRestarts:  5,
		restartCount: 0,
	}
}

func (s *Service) Start(ctx context.Context) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		wasAutoRestart := s.autoRestart
		restartCount := s.restartCount
		maxRestarts := s.maxRestarts
		s.isRunning = false
		s.mu.Unlock()

		// Handle panic with auto-restart
		if r := recover(); r != nil {
			_, panicRec := event.NewRecord(context.Background(), "deletion_daemon_fatal_panic")
			panicRec.Set("panic", r)
			panicRec.Set("panic_message", fmt.Sprintf("%v", r))
			panicRec.Set("restart_count", restartCount)
			panicRec.Add(events.Error, fmt.Errorf("deletion daemon crashed: %v", r))
			s.eventSink.ProcessEvent(panicRec)

			// Auto-restart if enabled and not exceeded max restarts
			if wasAutoRestart && restartCount < maxRestarts {
				s.mu.Lock()
				s.restartCount++
				s.mu.Unlock()

				time.Sleep(time.Second) // Brief delay before restart

				_, restartRec := event.NewRecord(context.Background(), "deletion_daemon_auto_restart")
				restartRec.Set("restart_attempt", restartCount+1)
				restartRec.Set("max_restarts", maxRestarts)
				s.eventSink.ProcessEvent(restartRec)

				go s.Start(ctx)
			} else {
				_, finalRec := event.NewRecord(context.Background(), "deletion_daemon_restart_limit_reached")
				finalRec.Set("restart_count", restartCount)
				finalRec.Set("max_restarts", maxRestarts)
				finalRec.Add(events.Error, errors.New("daemon crashed and restart limit reached"))
				s.eventSink.ProcessEvent(finalRec)
			}
		}
	}()

	ctx, daemonRec := event.NewRecord(ctx, "deletion_daemon")

	daemonRec.Set("enabled", s.enabled)
	daemonRec.Set("interval", s.interval)
	daemonRec.Set("restart_count", s.restartCount)

	if !s.enabled {
		daemonRec.Set("status", "disabled")
		s.eventSink.ProcessEvent(daemonRec)
		return
	}

	daemonRec.Set("status", "starting")
	daemonRec.Set("start_time", time.Now())

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	iterationCount := 0
	successfulIterations := 0

	for {
		select {
		case <-ticker.C:
			iterationCount++

			// Protect against panics in iteration
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicRec := event.Get(ctx).Sub("deletion_daemon_panic")
						panicRec.Set("iteration", iterationCount)
						panicRec.Set("panic", r)
						panicRec.Set("panic_message", fmt.Sprintf("%v", r))
						panicRec.Add(events.Error, fmt.Errorf("deletion daemon panicked: %v", r))
						s.eventSink.ProcessEvent(panicRec)
					}
				}()

				iterCtx, iterRec := event.NewRecord(ctx, "deletion_daemon_iteration")
				iterRec.Set("iteration_number", iterationCount)
				iterRec.Set("iteration_time", time.Now())
				iterRec.Set("daemon_config", map[string]any{
					"enabled":  s.enabled,
					"interval": s.interval,
				})

				s.processScheduledDeletions(iterCtx, iterRec)
				s.eventSink.ProcessEvent(iterRec)

				successfulIterations++
				if successfulIterations >= 3 {
					s.mu.Lock()
					s.restartCount = 0
					s.mu.Unlock()
					successfulIterations = 0
				}
			}()

		case <-s.stopChan:
			daemonRec.Set("status", "stopped_via_signal")
			daemonRec.Set("iterations_completed", iterationCount)
			s.eventSink.ProcessEvent(daemonRec)
			return
		case <-ctx.Done():
			daemonRec.Set("status", "stopped_via_context")
			daemonRec.Set("iterations_completed", iterationCount)
			s.eventSink.ProcessEvent(daemonRec)
			return
		}
	}
}

func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		close(s.stopChan)
	}
}

func (s *Service) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

func (s *Service) processScheduledDeletions(ctx context.Context, rec *event.Record) {
	processRec := rec.Sub("process_scheduled_deletions")
	processRec.Set("start_time", time.Now())
	processRec.Set("deletion_delay", s.delay)

	err := s.achService.ProcessScheduledDocumentDeletions(ctx, s.storage, s.delay)
	if err != nil {
		processRec.Add(events.Error, err)
		processRec.Set("success", false)
		return
	}

	processRec.Set("success", true)
	processRec.Set("duration_ms", time.Since(processRec.Value("start_time").(time.Time)).Milliseconds())
}
