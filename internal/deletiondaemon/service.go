package deletiondaemon

import (
	"context"
	"time"

	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/filesvc"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type Service struct {
	fileService *filesvc.FileService
	config      *config.DeletionDaemonConfig
	stopChan    chan struct{}
}

func New(fileService *filesvc.FileService, config *config.DeletionDaemonConfig) *Service {
	return &Service{
		fileService: fileService,
		config:      config,
		stopChan:    make(chan struct{}),
	}
}

// Start starts the deletion daemon
func (s *Service) Start(ctx context.Context) {
	ctx, daemonRec := event.NewRecord(ctx, "deletion_daemon")

	daemonRec.Set("enabled", s.config.Enabled)
	daemonRec.Set("interval", s.config.Interval)

	if !s.config.Enabled {
		daemonRec.Set("status", "disabled")
		return
	}

	daemonRec.Set("status", "starting")
	daemonRec.Set("start_time", time.Now())

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	iterationCount := 0
	for {
		select {
		case <-ticker.C:
			iterationCount++
			// Create new record for each iteration to ensure it gets logged
			ctx, iterRec := event.NewRecord(ctx, "deletion_daemon_iteration")
			iterRec.Set("iteration_number", iterationCount)
			iterRec.Set("iteration_time", time.Now())
			iterRec.Set("daemon_config", map[string]interface{}{
				"enabled":  s.config.Enabled,
				"interval": s.config.Interval,
			})

			s.processScheduledDeletions(ctx, iterRec)

		case <-s.stopChan:
			daemonRec.Set("status", "stopped_via_signal")
			daemonRec.Set("iterations_completed", iterationCount)
			return
		case <-ctx.Done():
			daemonRec.Set("status", "stopped_via_context")
			daemonRec.Set("iterations_completed", iterationCount)
			return
		}
	}
}

// Stop stops the deletion daemon
func (s *Service) Stop() {
	close(s.stopChan)
}

// processScheduledDeletions processes files that are ready for deletion
func (s *Service) processScheduledDeletions(ctx context.Context, rec *event.Record) {
	processRec := rec.Sub("process_scheduled_deletions")
	processRec.Set("start_time", time.Now())

	err := s.fileService.ProcessScheduledDeletions(ctx)
	if err != nil {
		processRec.Add(events.Error, err)
		processRec.Set("success", false)
		return
	}

	processRec.Set("success", true)
	processRec.Set("duration_ms", time.Since(processRec.Value("start_time").(time.Time)).Milliseconds())
}
