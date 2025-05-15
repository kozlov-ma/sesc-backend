package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kozlov-ma/sesc-backend/api"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/migrate"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/internal/slogsink"
	"github.com/kozlov-ma/sesc-backend/sesc"
	_ "github.com/lib/pq"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client, err := ent.Open("postgres", os.Getenv("POSTGRES_ADDRESS"))
	if err != nil {
		log.ErrorContext(ctx, "failed to set up database", "error", err)
		return
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Error("couldn't close db", "error", err)
		}
	}()

	if err := client.Schema.Create(ctx, migrate.WithDropIndex(true), migrate.WithDropColumn(true)); err != nil {
		log.Error("couldn't apply migrations", "error", err)
		return
	}

	iam := iam.New(client, 7*24*time.Hour, []iam.Credentials{
		{
			Username: "admin",
			Password: "admin",
		},
	}, []byte("dinahu"))

	sesc := sesc.New(client)
	api := api.New(sesc, iam, slogsink.New(log))

	router := chi.NewRouter()

	api.RegisterRoutes(router)

	const readHeaderTimeout = 300 * time.Millisecond // Put in config.

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("couldn't shut down server", "error", err)
		}
	}()

	log.InfoContext(ctx, "starting server")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.ErrorContext(ctx, "failed to start server", "error", err)
	}
}
