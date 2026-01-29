package database

import (
	"context"
	"fmt"
	"invoice-worker/internal/config"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Option func(*pgxpool.Config)

func WithMaxConns(n int32) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConns = n
	}
}

func WithMinConns(n int32) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MinConns = n
	}
}

func WithMaxConnLifetime(d time.Duration) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnLifetime = d
	}
}

func NewPool(ctx context.Context, dbConfig config.DBConfig, logger *slog.Logger, opts ...Option) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	if dbConfig.MaxConns > 0 {
		if dbConfig.MaxConns > math.MaxInt32 {
			return nil, fmt.Errorf("max connections %d exceeds int32 limit", dbConfig.MaxConns)
		}
		config.MaxConns = int32(dbConfig.MaxConns)
	}
	if dbConfig.MinConns > 0 {
		if dbConfig.MinConns > math.MaxInt32 {
			return nil, fmt.Errorf("min connections %d exceeds int32 limit", dbConfig.MinConns)
		}
		config.MinConns = int32(dbConfig.MinConns)
	}

	config.MaxConnLifetime = dbConfig.MaxConnLifetime

	for _, opt := range opts {
		opt(config)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("database connected",
		slog.Int("max_conns", int(config.MaxConns)),
		slog.Int("min_conns", int(config.MinConns)),
		slog.Bool("tracer_enabled", config.ConnConfig.Tracer != nil),
	)

	return pool, nil
}
