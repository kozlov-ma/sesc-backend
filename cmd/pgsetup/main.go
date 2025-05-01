package main

import (
	"database/sql"
	"embed"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	db, err := sql.Open("pgx", os.Getenv("POSTGRES_ADDRESS"))
	if err != nil {
		log.Error("couldn't open postgres", "error", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("couldn't close db", "error", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Error("no access to db", "error", err)
		return
	}

	_ = goose.SetDialect("postgresql")
	goose.SetBaseFS(embedMigrations)

	if err := goose.Up(db, "migrations"); err != nil {
		log.Error("couldn't migrate", "error", err)
		return
	}

	if err := goose.Version(db, "migrations"); err != nil {
		log.Error("couldn't get db version, or version is wrong", "error", err)
		return
	}

	slog.Info("migration done")
}
