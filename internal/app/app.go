package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kozlov-ma/sesc-backend/api"
	"github.com/kozlov-ma/sesc-backend/auth/authservice"
	"github.com/kozlov-ma/sesc-backend/company/companyservice"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/migrate"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/filesvc"
	"github.com/kozlov-ma/sesc-backend/internal/s3svc"
	"github.com/kozlov-ma/sesc-backend/internal/sescsvc"
	"github.com/kozlov-ma/sesc-backend/internal/slogsink"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	Router      *chi.Mux
	Server      *http.Server
	Client      *ent.Client
	API         *api.API
	Log         *slog.Logger
	FileService *filesvc.FileService
	Cleanup     func()
}

type DBOptions struct {
	SkipMigrations bool
	Client         *ent.Client
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	return NewWithDBOptions(ctx, cfg, log, DBOptions{})
}

func NewWithDBOptions(ctx context.Context, cfg *config.Config, log *slog.Logger, dbOpts DBOptions) (*App, error) {
	var client *ent.Client
	var err error
	var cleanup func()

	if dbOpts.Client != nil {
		client = dbOpts.Client
		cleanup = func() {}
	} else {
		dbType := string(cfg.Database.Type)
		if dbType == "" {
			dbType = string(config.DatabaseTypePostgres)
		}

		client, err = ent.Open(dbType, cfg.Database.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to set up database: %w", err)
		}

		cleanup = func() {
			if err := client.Close(); err != nil {
				log.ErrorContext(ctx, "couldn't close db", "error", err)
			}
		}
	}

	if !dbOpts.SkipMigrations {
		if err := client.Schema.Create(ctx, migrate.WithDropIndex(true), migrate.WithDropColumn(true)); err != nil {
			cleanup()
			return nil, fmt.Errorf("couldn't apply migrations: %w", err)
		}
	}

	ldapConfig := companyservice.LDAPConfig{
		URL:          cfg.LDAP.URL,
		BindDN:       cfg.LDAP.BindDN,
		BindPassword: cfg.LDAP.BindPassword,
		BaseDN:       cfg.LDAP.BaseDN,
		SyncInterval: cfg.LDAP.SyncInterval,
	}

	eventSink := slogsink.New(log)

	cs, err := companyservice.NewLDAPService(ctx, ldapConfig, eventSink)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to create LDAP company service: %w", err)
	}

	sescService := sescsvc.New(client, cs)

	s3Storage, err := s3svc.NewStorage(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.BucketName,
		cfg.MinIO.UseSSL,
		cfg.MinIO.BaseURL,
	)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to create S3 storage: %w", err)
	}

	fileService := filesvc.New(
		client,
		s3Storage,
		cfg.MinIO.BucketName,
	)

	authSvc := authservice.NewCompany(cs, 24*time.Hour, []byte(cfg.JWTSecret))
	apiService := api.New(sescService, authSvc, fileService, slogsink.New(log))

	router := chi.NewRouter()
	apiService.RegisterRoutes(router)

	server := &http.Server{
		Addr:              cfg.HTTP.ServerAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
	}

	return &App{
		Router:      router,
		Server:      server,
		Client:      client,
		API:         apiService,
		Log:         log,
		FileService: fileService,
		Cleanup:     cleanup,
	}, nil
}

const shutdownTimeout = 15 * time.Second

func (a *App) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := a.Server.Shutdown(shutdownCtx); err != nil {
			a.Log.Error("couldn't shut down server", "error", err)
		}
	}()

	a.Log.InfoContext(ctx, "starting server", "address", a.Server.Addr)
	if err := a.Server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (a *App) Close() {
	if a.Cleanup != nil {
		a.Cleanup()
	}
}
