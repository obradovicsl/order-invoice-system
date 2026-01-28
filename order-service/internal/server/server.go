package server

import (
	"context"
	"log/slog"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/database"
	"order-service/internal/features/order"
	"order-service/internal/repository"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

	logger.Info("Initializing repositories")
	queries := repository.New(pool)

	logger.Info("Initializing services")
	orderService := order.NewService(queries, logger)

	router := NewRouter(*config, orderService, logger)

	return &Server{
		Port:   config.Port,
		Logger: logger,
		Router: router,
		pool:   pool,
	}
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
