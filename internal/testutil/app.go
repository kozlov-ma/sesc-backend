package testutil

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/internal/app"
	"github.com/stretchr/testify/require"
)

// TestApp wraps the App for testing purposes
type TestApp struct {
	*app.App
	URL string
}

// StartTestApp creates and starts a test application
func StartTestApp(t *testing.T) *TestApp {
	t.Helper()

	ctx := t.Context()
	var log *slog.Logger
	if os.Getenv("TEST_PRINT_EVENTS") == "1" {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		log = slog.New(slog.DiscardHandler)
	}

	cfg := CreateTestConfig()

	var dbOpts app.DBOptions

	client := enttest.Open(t, string(cfg.Database.Type), cfg.Database.Address)
	dbOpts = app.DBOptions{
		Client:         client,
		SkipMigrations: true,
	}

	application, err := app.NewWithDBOptions(ctx, cfg, log, dbOpts)
	require.NoError(t, err, "Failed to create test app")

	// Get the listener to determine which port was assigned
	listener, err := net.Listen("tcp", cfg.HTTP.ServerAddress)
	require.NoError(t, err, "Failed to create listener")

	// Update the server to use our listener
	application.Server.Addr = listener.Addr().String()

	// Start the server in a goroutine
	go func() {
		if err := application.Server.Serve(listener); err != http.ErrServerClosed {
			t.Errorf("HTTP server exited with error: %v", err)
		}
	}()

	url := "http://" + listener.Addr().String()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(t.Context(), cfg.HTTP.WriteTimeout)
		defer cancel()

		err := application.Server.Shutdown(ctx)
		if err != nil {
			t.Logf("Error shutting down server: %v", err)
		}

		application.Close()
		_ = client.Close()
	})

	return &TestApp{
		App: application,
		URL: url,
	}
}
