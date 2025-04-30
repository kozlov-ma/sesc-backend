package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/kozlov-ma/sesc-backend/api"
	"github.com/kozlov-ma/sesc-backend/db/pgdb"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	db, err := pgdb.Connect(ctx, log, os.Getenv("POSTGRES_ADDRESS"))
	if err != nil {
		log.ErrorContext(ctx, "failed to set up database", "error", err)
		return
	}

	defer db.Close()

	sesc := sesc.New(log, db, nil)

	api := api.New(log, sesc)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()

	log.InfoContext(ctx, "starting server")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.ErrorContext(ctx, "failed to start server", "error", err)
	}
}
