package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/kozlov-ma/sesc-backend/api"
	"github.com/kozlov-ma/sesc-backend/db/entdb"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/migrate"
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

	db := entdb.New(log, client)

	sesc := sesc.New(log, db)
	api := api.New(log, sesc)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
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
