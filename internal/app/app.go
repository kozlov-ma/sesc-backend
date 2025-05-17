package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kozlov-ma/sesc-backend/api"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/migrate"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/slogsink"
	"github.com/kozlov-ma/sesc-backend/sesc"
	// database driver
	_ "github.com/lib/pq"
	// database driver
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	Router  *chi.Mux
	Server  *http.Server
	Client  *ent.Client
	API     *api.API
	Log     *slog.Logger
	Cleanup func()
}

// DBOptions contains options for database initialization
type DBOptions struct {
	// If true, skip running migrations
	SkipMigrations bool
	// Custom client to use instead of creating a new one
	Client *ent.Client
}

// New creates a new application instance from the given config
func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	return NewWithDBOptions(ctx, cfg, log, DBOptions{})
}

// NewWithDBOptions creates a new application instance with custom database options
func NewWithDBOptions(ctx context.Context, cfg *config.Config, log *slog.Logger, dbOpts DBOptions) (*App, error) {
	var client *ent.Client
	var err error
	var cleanup func()

	// Use provided client or create a new one
	if dbOpts.Client != nil {
		client = dbOpts.Client
		cleanup = func() {
			// Don't close the client if it was provided externally
		}
	} else {
		dbType := string(cfg.Database.Type)
		if dbType == "" {
			dbType = string(config.DatabaseTypePostgres)
		}

		client, err = ent.Open(dbType, cfg.Database.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to set up database: %w", err)
		}

		// Setup auto cleanup function
		cleanup = func() {
			if err := client.Close(); err != nil {
				log.ErrorContext(ctx, "couldn't close db", "error", err)
			}
		}
	}

	// Run migrations if not skipped
	if !dbOpts.SkipMigrations {
		if err := client.Schema.Create(ctx, migrate.WithDropIndex(true), migrate.WithDropColumn(true)); err != nil {
			cleanup()
			return nil, fmt.Errorf("couldn't apply migrations: %w", err)
		}
	}

	adminCredentials, err := cfg.ToIAMAdminCredentials()
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to convert admin credentials: %w", err)
	}

	iamService := iam.New(client, 7*24*time.Hour, adminCredentials, []byte(cfg.JWTSecret))
	sescService := sesc.New(client)
	apiService := api.New(sescService, iamService, slogsink.New(log))

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
		Router:  router,
		Server:  server,
		Client:  client,
		API:     apiService,
		Log:     log,
		Cleanup: cleanup,
	}, nil
}

const shutdownTimeout = 15 * time.Second

// Start starts the HTTP server
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

// Close cleans up resources used by the app
func (a *App) Close() {
	if a.Cleanup != nil {
		a.Cleanup()
	}
}
