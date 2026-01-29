package main

import (
	"context"
	"invoice-worker/internal/config"
	"invoice-worker/internal/database"
	"invoice-worker/internal/logger"
	"invoice-worker/internal/repository"
	"invoice-worker/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	workerPollInterval = 2 * time.Second
)

func main() {
	// Load configuration
	cfg := config.NewConfig()
	log := logger.NewLogger(cfg.Logger)

	log.Info("Initializing invoice worker")

	// Initialize database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info("Connecting to database")
	pool, err := database.NewPool(ctx, cfg.Database, log)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("Initializing Azure services")

	// Initialize Queue Service
	queueService, err := service.NewQueueService(cfg.Azure.StorageConnectionString, log)
	if err != nil {
		log.Error("Failed to initialize queue service", "error", err)
		os.Exit(1)
	}

	// Initialize Blob Service
	blobService, err := service.NewBlobService(
		cfg.Azure.StorageConnectionString,
		cfg.Azure.BlobContainerName,
		log,
	)
	if err != nil {
		log.Error("Failed to initialize blob service", "error", err)
		os.Exit(1)
	}

	// Initialize Invoice Service
	invoiceService := service.NewInvoiceService(log)

	// Initialize repository and order service
	log.Info("Initializing order service")
	queries := repository.New(pool)
	orderService := service.NewOrderService(queries, log)

	// Initialize Invoice Worker
	log.Info("Initializing invoice worker")
	invoiceWorker := service.NewInvoiceWorker(
		queueService,
		blobService,
		invoiceService,
		orderService,
		pool,
		cfg.Azure.QueueName,
		log,
	)

	// Validate queue and blob container
	validationCtx, validationCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := invoiceWorker.ValidateQueueAndContainer(validationCtx, cfg.Azure.BlobContainerName); err != nil {
		log.Warn("Queue/container validation warning", "error", err)
	}
	validationCancel()

	log.Info("Starting invoice worker",
		"queue_name", cfg.Azure.QueueName,
		"blob_container", cfg.Azure.BlobContainerName,
		"poll_interval", workerPollInterval.String(),
	)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create context for worker
	workerCtx, workerCancel := context.WithCancel(context.Background())

	// Run worker in goroutine
	workerErrors := make(chan error, 1)
	go func() {
		workerErrors <- invoiceWorker.Start(workerCtx, workerPollInterval)
	}()

	// Wait for either worker error or shutdown signal
	select {
	case err := <-workerErrors:
		if err != nil {
			log.Error("Worker error", "error", err)
			os.Exit(1)
		}
	case sig := <-sigChan:
		log.Info("Received shutdown signal", "signal", sig.String())
	}

	// Graceful shutdown
	log.Info("Shutting down worker gracefully")
	workerCancel()

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	log.Info("Worker shutdown complete")
}
