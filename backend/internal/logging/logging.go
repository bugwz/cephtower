package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"cephtower/backend/internal/app"
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

	options := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
				return slog.String(slog.TimeKey, attr.Value.Time().Format(time.RFC3339))
			}
			return attr
		},
	}
	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "", "txt":
		handler = newPlainTextHandler(output, options)
	case "json":
		handler = slog.NewJSONHandler(output, options)
	default:
		return nil, fmt.Errorf("unsupported logging format %q", cfg.Format)
	}

	return slog.New(handler), nil
}

type plainTextHandler struct {
	level  slog.Leveler
	output io.Writer
}

func newPlainTextHandler(output io.Writer, options *slog.HandlerOptions) slog.Handler {
	return &plainTextHandler{level: options.Level, output: output}
}

func (h *plainTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *plainTextHandler) Handle(_ context.Context, record slog.Record) error {
	_, err := fmt.Fprintf(h.output, "%s %s %s\n", record.Time.Format(time.RFC3339), record.Level, record.Message)
	return err
}

func (h *plainTextHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *plainTextHandler) WithGroup(_ string) slog.Handler {
	return h
}

func Install(cfg config.LoggingConfig, workDirs ...string) (*slog.Logger, func() error, error) {
	app.Global.LogRetentionCleanup = nil
	output := strings.ToLower(strings.TrimSpace(cfg.Output))
	if output == "" {
		output = "both"
	}

	var writer io.Writer = os.Stdout
	var fileWriter *rotatingFileWriter
	var cleanupTask func(context.Context)
	var err error
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
		fileWriter, err = newRotatingFileWriter(path, cfg.Rotation, cfg.Retention)
		if err != nil {
			return nil, func() error { return nil }, err
		}
		if output == "file" {
			writer = fileWriter
		} else {
			writer = io.MultiWriter(os.Stdout, fileWriter)
		}
		cleanupTask = func(_ context.Context) { fileWriter.runCleanup(time.Now()) }
		app.Global.LogRetentionCleanup = cleanupTask
	}

	logger, err := NewLogger(cfg, writer)
	if err != nil {
		if fileWriter != nil {
			_ = fileWriter.Close()
		}
		return nil, func() error { return nil }, err
	}
	slog.SetDefault(logger)
	return logger, func() error {
		if fileWriter != nil {
			return fileWriter.Close()
		}
		return nil
	}, nil
}

type rotatingFileWriter struct {
	mu             sync.Mutex
	file           *os.File
	path           string
	rotationDays   int
	retentionDays  int
	nextRotationAt time.Time
	now            func() time.Time
}

func newRotatingFileWriter(path, rotation, retention string) (*rotatingFileWriter, error) {
	if strings.TrimSpace(rotation) == "" {
		rotation = "7days"
	}
	if strings.TrimSpace(retention) == "" {
		retention = "70days"
	}
	rotateEvery, err := config.ParseDuration(rotation)
	if err != nil {
		return nil, fmt.Errorf("parse logging rotation: %w", err)
	}
	retainFor, err := config.ParseDuration(retention)
	if err != nil {
		return nil, fmt.Errorf("parse logging retention: %w", err)
	}
	rotationDays := int(rotateEvery / (24 * time.Hour))
	retentionDays := int(retainFor / (24 * time.Hour))

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open log file %q: %w", path, err)
	}
	now := time.Now()
	fileStart := now
	if info, statErr := file.Stat(); statErr == nil && info.Size() > 0 {
		fileStart = info.ModTime()
	}
	if historyStart, found := latestHistoryStart(path, now.Location()); found {
		fileStart = historyStart
	}
	writer := &rotatingFileWriter{
		file:           file,
		path:           path,
		rotationDays:   rotationDays,
		retentionDays:  retentionDays,
		nextRotationAt: nextRotationAt(fileStart, rotationDays),
		now:            time.Now,
	}
	if !writer.now().Before(writer.nextRotationAt) {
		if err := writer.rotate(writer.now()); err != nil {
			_ = file.Close()
			return nil, err
		}
	}
	return writer, nil
}

func (w *rotatingFileWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.now()
	if !now.Before(w.nextRotationAt) {
		if err := w.rotate(now); err != nil {
			return 0, err
		}
	}
	return w.file.Write(data)
}

func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *rotatingFileWriter) rotate(now time.Time) error {
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close log file before rotation: %w", err)
	}
	historyPath := w.historyPath(now)
	if err := os.Rename(w.path, historyPath); err != nil {
		return fmt.Errorf("rotate log file: %w", err)
	}
	file, err := os.OpenFile(w.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("reopen log file after rotation: %w", err)
	}
	w.file = file
	w.nextRotationAt = nextRotationAt(now, w.rotationDays)
	return nil
}

func (w *rotatingFileWriter) runCleanup(now time.Time) {
	if err := w.removeExpired(now); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "clean historical log files: %v\n", err)
	}
}

func nextRotationAt(start time.Time, rotationDays int) time.Time {
	startOfDay := startOfDay(start)
	return startOfDay.AddDate(0, 0, rotationDays)
}

func startOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func latestHistoryStart(path string, location *time.Location) (time.Time, bool) {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(path), base+"-*"+ext))
	if err != nil {
		return time.Time{}, false
	}
	var latest time.Time
	for _, historyPath := range paths {
		name := strings.TrimSuffix(filepath.Base(historyPath), ext)
		stamp := strings.TrimPrefix(name, base+"-")
		if len(stamp) < len("20060102150405") {
			continue
		}
		createdAt, parseErr := time.ParseInLocation("20060102150405", stamp[:len("20060102150405")], location)
		if parseErr == nil && createdAt.After(latest) {
			latest = createdAt
		}
	}
	return latest, !latest.IsZero()
}

func (w *rotatingFileWriter) historyPath(now time.Time) string {
	ext := filepath.Ext(w.path)
	base := strings.TrimSuffix(w.path, ext)
	path := base + "-" + now.Format("20060102150405") + ext
	for index := 1; ; index++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
		path = fmt.Sprintf("%s-%d%s", base+"-"+now.Format("20060102150405"), index, ext)
	}
}

func (w *rotatingFileWriter) removeExpired(now time.Time) error {
	ext := filepath.Ext(w.path)
	base := strings.TrimSuffix(filepath.Base(w.path), ext)
	pattern := filepath.Join(filepath.Dir(w.path), base+"-*"+ext)
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("find historical log files: %w", err)
	}
	cutoff := startOfDay(now).AddDate(0, 0, -w.retentionDays)
	sort.Strings(paths)
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ext)
		stamp := strings.TrimPrefix(name, base+"-")
		if len(stamp) < len("20060102150405") {
			continue
		}
		stamp = stamp[:len("20060102150405")]
		createdAt, err := time.ParseInLocation("20060102150405", stamp, now.Location())
		if err != nil || !createdAt.Before(cutoff) {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove expired log file %q: %w", path, err)
		}
	}
	return nil
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
