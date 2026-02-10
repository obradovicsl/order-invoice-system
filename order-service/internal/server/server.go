package server

import (
	"context"
	"log/slog"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/database"
	"order-service/internal/features/order"
	"order-service/internal/repository"
	"order-service/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	Port             string
	Logger           *slog.Logger
	Router           http.Handler
	pool             *pgxpool.Pool
	completionWorker *service.OrderCompletionWorker
	workerCtx        context.Context
	workerCancel     context.CancelFunc
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

	logger.Info("Initializing Azure services")
	queueService, err := service.NewQueueService(config.Azure.StorageConnectionString, logger)
	if err != nil {
		logger.Error("Failed to initialize queue service", "error", err)
		os.Exit(1)
	}

	blobService, err := service.NewBlobService(config.Azure.StorageConnectionString, config.Azure.BlobContainerName, logger)
	if err != nil {
		logger.Error("Failed to initialize blob service", "error", err)
		os.Exit(1)
	}

	logger.Info("Initializing order service")
	orderService := order.NewService(queries, logger, queueService, config.Azure.QueueName, blobService, config.Azure.BlobContainerName)

	logger.Info("Initializing order completion worker")
	completionWorker := service.NewOrderCompletionWorker(queueService, queries, pool, config.Azure.ReadyQueueName, logger)

	// Validate ready queue
	validationCtx, validationCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := completionWorker.ValidateQueue(validationCtx); err != nil {
		logger.Warn("Ready queue validation warning", "error", err)
	}
	validationCancel()

	router := NewRouter(*config, orderService, logger)

	workerCtx, workerCancel := context.WithCancel(context.Background())

	return &Server{
		Port:             config.Port,
		Logger:           logger,
		Router:           router,
		pool:             pool,
		completionWorker: completionWorker,
		workerCtx:        workerCtx,
		workerCancel:     workerCancel,
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

	// Run completion worker in a goroutine
	const workerPollInterval = 2 * time.Second
	serverInstance.Logger.Info("Starting order completion worker",
		"poll_interval", workerPollInterval.String(),
	)
	workerErrors := make(chan error, 1)
	go func() {
		workerErrors <- serverInstance.completionWorker.Start(serverInstance.workerCtx, workerPollInterval)
	}()

	// Wait for either server error, worker error, or shutdown signal
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			serverInstance.Logger.Error("Server error", "error", err)
			return err
		}
	case err := <-workerErrors:
		if err != nil {
			serverInstance.Logger.Error("Worker error", "error", err)
			return err
		}
	case sig := <-sigChan:
		serverInstance.Logger.Info("Received shutdown signal", "signal", sig.String())
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverInstance.Logger.Info("Shutting down server and worker gracefully")
	serverInstance.workerCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		serverInstance.Logger.Error("Error during server shutdown", "error", err)
	}

	// Close database connection
	serverInstance.Logger.Info("Closing database connection")
	serverInstance.pool.Close()

	serverInstance.Logger.Info("Server and worker shutdown complete")
	return nil
}
