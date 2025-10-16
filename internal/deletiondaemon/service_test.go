package deletiondaemon

import (
	"context"
	"testing"
	"time"

	"github.com/kozlov-ma/sesc-backend/pkg/event"
)

// mockEventSink implements EventMiddleware interface
type mockEventSink struct {
	events      []*event.Record
	panicEvents []*event.Record
}

func (m *mockEventSink) ProcessEvent(r *event.Record) {
	m.events = append(m.events, r)

	// Check if this is a panic event (has "deletion_daemon_panic" in path)
	if r.Value("panic") != nil {
		m.panicEvents = append(m.panicEvents, r)
	}
}

func TestService_IsRunning(t *testing.T) {
	mockSink := &mockEventSink{}

	svc := &Service{
		achService: nil, // Not needed for this test
		storage:    nil,
		delay:      1 * time.Hour,
		enabled:    false,
		interval:   100 * time.Millisecond,
		eventSink:  mockSink,
		stopChan:   make(chan struct{}),
		isRunning:  false,
	}

	if svc.IsRunning() {
		t.Error("Daemon should not be running initially")
	}
}

func TestService_DisabledDaemon(t *testing.T) {
	mockSink := &mockEventSink{}

	svc := &Service{
		achService: nil,
		storage:    nil,
		delay:      1 * time.Hour,
		enabled:    false, // disabled
		interval:   100 * time.Millisecond,
		eventSink:  mockSink,
		stopChan:   make(chan struct{}),
		isRunning:  false,
	}

	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()

	// Start daemon (should exit immediately since disabled)
	svc.Start(ctx)

	if svc.IsRunning() {
		t.Error("Disabled daemon should not be running")
	}

	// Check that at least one event was processed (the disabled status)
	if len(mockSink.events) == 0 {
		t.Error("Expected at least one event to be processed")
	}
}
