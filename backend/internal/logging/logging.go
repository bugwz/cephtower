package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"cephtower/backend/internal/config"
)

func NewLogger(cfg config.LoggingConfig, output io.Writer) (*slog.Logger, error) {
	if output == nil {
		output = os.Stderr
	}

	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	options := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "", "txt":
		handler = slog.NewTextHandler(output, options)
	case "json":
		handler = slog.NewJSONHandler(output, options)
	default:
		return nil, fmt.Errorf("unsupported logging format %q", cfg.Format)
	}

	return slog.New(handler), nil
}

func Install(cfg config.LoggingConfig) (*slog.Logger, error) {
	logger, err := NewLogger(cfg, os.Stderr)
	if err != nil {
		return nil, err
	}
	slog.SetDefault(logger)
	return logger, nil
}

func parseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported logging level %q", value)
	}
}
