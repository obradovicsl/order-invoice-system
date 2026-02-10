package config

import (
	"encoding/json"
	"fmt"
	"invoice-worker/internal/logger"
	"invoice-worker/internal/utils/loader"
	"log/slog"
	"time"
)

type Config struct {
	Env      string
	Port     string
	Logger   logger.LoggerConfig
	Database DBConfig
	Azure    AzureConfig
}

type DBConfig struct {
	URL             string
	QueryTracer     bool
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
}

type AzureConfig struct {
	StorageConnectionString string
	BlobContainerName       string
	QueueName               string
	ReadyQueueName          string
}

func NewConfig() *Config {

	cfg := &Config{
		Env:      loader.GetEnv("ENV", "dev"),
		Port:     loader.GetEnv("SERVICE_PORT", "8084"),
		Logger:   loadLoggerConfig(),
		Database: loadDatabaseConfig(),
		Azure:    loadAzureConfig(),
	}

	cfgJSON, _ := json.MarshalIndent(cfg, "", " ")
	//nolint:govet
	slog.Info("Loaded SQS", "config", string(cfgJSON))

	return cfg
}

func loadLoggerConfig() logger.LoggerConfig {

	logger := logger.LoggerConfig{
		DefaultLevel: loader.GetEnv("LOG_LEVEL", "info"),
	}

	return logger
}

func loadDatabaseConfig() DBConfig {
	dbHost := loader.GetEnv("DB_HOST", "localhost")
	dbPort := loader.GetEnv("DB_PORT", "5432")
	dbName := loader.GetEnv("DB_NAME", "postgres")
	dbUser := loader.GetEnv("DB_USER", "postgres")
	dbPassword := loader.GetEnv("DB_PASSWORD", "postgres")
	dbSSLMode := loader.GetEnv("DB_SSLMODE", "require")

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	dbConfig := DBConfig{
		URL:             dbUrl,
		QueryTracer:     loader.GetEnvAsBool("DB_QUERY_TRACER", false),
		MaxConns:        loader.GetEnvAsInt("DB_MAX_OPEN_CONNS", 0),
		MinConns:        loader.GetEnvAsInt("DB_MIN_OPEN_CONNS", 0),
		MaxConnLifetime: time.Duration(loader.GetEnvAsInt("DB_MAX_CONN_LIFETIME", 0)) * time.Minute,
	}

	return dbConfig
}

func loadAzureConfig() AzureConfig {
	return AzureConfig{
		StorageConnectionString: loader.GetEnvOrCrash("AZURE_STORAGE_CONNECTION_STRING"),
		BlobContainerName:       loader.GetEnv("BLOB_CONTAINER_NAME", "invoices"),
		QueueName:               loader.GetEnv("QUEUE_NAME", "invoice-queue"),
		ReadyQueueName:          loader.GetEnv("READY_QUEUE_NAME", "ready-queue"),
	}
}
