package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kozlov-ma/sesc-backend/internal/app"
	"github.com/kozlov-ma/sesc-backend/internal/config"
)

const shutdownTimeout = 5 * time.Second

func main() {
	// Setup logging
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create and initialize application
	ctx := context.Background()
	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("Failed to create application", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	// Start application server
	go func() {
		log.Info("Starting test API server", "address", cfg.HTTP.ServerAddress)
		if err := application.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", "error", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down test API server...")

	// Shutdown gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := application.Server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Test API server exited")
}
