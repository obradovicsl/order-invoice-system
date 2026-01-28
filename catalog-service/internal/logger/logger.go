package logger

import (
	"log/slog"
	"os"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

type LoggerConfig struct {
	DefaultLevel string
}

func NewLogger(config LoggerConfig) *slog.Logger {
	defaultLevel := parseLogLevel(config.DefaultLevel)

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel,
	}))
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
