package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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

func Install(cfg config.LoggingConfig, workDirs ...string) (*slog.Logger, func() error, error) {
	output := strings.ToLower(strings.TrimSpace(cfg.Output))
	if output == "" {
		output = "both"
	}

	var writer io.Writer = os.Stdout
	var logFile *os.File
	if output == "file" || output == "both" {
		workDir := "./app"
		if len(workDirs) > 0 && workDirs[0] != "" {
			workDir = workDirs[0]
		}
		path := cfg.Path
		if path == "" {
			path = "log/cephtower.log"
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, func() error { return nil }, fmt.Errorf("create log directory: %w", err)
		}
		logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return nil, func() error { return nil }, fmt.Errorf("open log file %q: %w", path, err)
		}
		if output == "file" {
			writer = logFile
		} else {
			writer = io.MultiWriter(os.Stdout, logFile)
		}
	}

	logger, err := NewLogger(cfg, writer)
	if err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return nil, func() error { return nil }, err
	}
	slog.SetDefault(logger)
	return logger, func() error {
		if logFile != nil {
			return logFile.Close()
		}
		return nil
	}, nil
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
