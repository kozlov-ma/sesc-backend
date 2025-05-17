package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/kozlov-ma/sesc-backend/internal/app"
	"github.com/kozlov-ma/sesc-backend/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.ErrorContext(ctx, "failed to load configuration", "error", err)
		return
	}

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.ErrorContext(ctx, "failed to create application", "error", err)
		return
	}
	defer application.Close()

	if err := application.Start(ctx); err != nil {
		log.ErrorContext(ctx, "application error", "error", err)
		return
	}
}
