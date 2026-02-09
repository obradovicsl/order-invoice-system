package server

import (
	"catalog-service/internal/config"
	"catalog-service/internal/database"
	"catalog-service/internal/features/catalog"
	"catalog-service/internal/repository"
	"catalog-service/internal/service"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Server struct {
	Port   string
	Logger *slog.Logger
	Router http.Handler
	pool   *pgxpool.Pool
}

func NewServer(config *config.Config, logger *slog.Logger) *Server {

	logger.Info("Initializing database")

	pool, err := database.NewPool(context.Background(), config.Database, logger)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	logger.Info("Running database migrations")
	if err := runMigrations(pool, logger); err != nil {
		logger.Error("Migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Initializing repositories")
	queries := repository.New(pool)

	logger.Info("Initializing services")
	blobService, err := service.NewBlobService(config.Azure.StorageConnectionString, config.Azure.BlobContainerName, logger)
	if err != nil {
		logger.Error("Failed to initialize blob service", "error", err)
		os.Exit(1)
	}

	catalogService := catalog.NewService(queries, logger, blobService, config.Azure.BlobContainerName)

	router := NewRouter(*config, catalogService, logger)

	return &Server{
		Port:   config.Port,
		Logger: logger,
		Router: router,
		pool:   pool,
	}
}

func runMigrations(pool *pgxpool.Pool, logger *slog.Logger) error {
	// Napravi novu *sql.DB konekciju iz connection stringa
	connString := pool.Config().ConnConfig.ConnString()

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("could not open DB for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	// Pronađi apsolutnu putanju do migracija
	migrationPath, err := filepath.Abs("db/migrations")
	if err != nil {
		return fmt.Errorf("could not resolve migration path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	logger.Info("✓ Migrations applied successfully")
	return nil
}

func (serverInstance *Server) Start() error {
	serverInstance.Logger.Info("Server starting", "port", serverInstance.Port)

	server := &http.Server{
		Addr:         ":" + serverInstance.Port,
		Handler:      serverInstance.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			serverInstance.Logger.Error("Server error", "error", err)
			return err
		}
	case sig := <-sigChan:
		serverInstance.Logger.Info("Received shutdown signal", "signal", sig.String())
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverInstance.Logger.Info("Shutting down server gracefully")
	if err := server.Shutdown(shutdownCtx); err != nil {
		serverInstance.Logger.Error("Error during server shutdown", "error", err)
	}

	// Close database connection
	serverInstance.Logger.Info("Closing database connection")
	serverInstance.pool.Close()

	serverInstance.Logger.Info("Server shutdown complete")
	return nil
}
