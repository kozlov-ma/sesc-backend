package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	ldapds "github.com/kozlov-ma/sesc-backend/company/companyservice/ldaps"
	"github.com/kozlov-ma/sesc-backend/internal/app"
	"github.com/kozlov-ma/sesc-backend/internal/config"
)

// @title SESC Management API
// @version 1.0
// @description API for managing SESC departments, users and permissions
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter 'Bearer ' followed by your token
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.ErrorContext(ctx, "failed to load configuration", "error", err)
		return
	}

	ldap, err := ldapds.New(ldapds.Config{
		Address:  "localhost:389",
		BaseDN:   "dc=lyceum,dc=usu,dc=ru",
		BindUser: "cn=admin,dc=lyceum,dc=usu,dc=ru",
		BindPass: "Admin123!Pass",
	})

	if err != nil {
		log.ErrorContext(ctx, "failed to set up an ldap data source", "error", err)
		return
	}

	application, err := app.NewWithDBOptions(ctx, cfg, log, app.DBOptions{
		CompanyDS: ldap,
	})

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
