package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
)

func TestNewLoggerWritesSingleLineJSON(t *testing.T) {
	var output bytes.Buffer
	logger, err := NewLogger(config.LoggingConfig{Level: "info", Format: "json"}, &output)
	if err != nil {
		t.Fatalf("NewLogger() returned error: %v", err)
	}

	logger.Info("backend started")

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("log line count = %d, want 1; output: %q", len(lines), output.String())
	}

	var entry map[string]string
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("json log line is invalid: %v; line: %q", err, lines[0])
	}
	if entry["time"] == "" || strings.Contains(entry["time"], ".") || entry["level"] != "INFO" || entry["msg"] != "backend started" {
		t.Fatalf("unexpected json log entry: %#v", entry)
	}
}

func TestNewLoggerWritesPlainTextWithoutFieldNames(t *testing.T) {
	var output bytes.Buffer
	logger, err := NewLogger(config.LoggingConfig{Level: "info", Format: "txt"}, &output)
	if err != nil {
		t.Fatalf("NewLogger() returned error: %v", err)
	}

	logger.Info("backend started", "engine", "sqlite")

	line := strings.TrimSpace(output.String())
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		t.Fatalf("plain text log should contain time, level, and message: %q", line)
	}
	if _, err := time.Parse(time.RFC3339Nano, parts[0]); err != nil {
		t.Fatalf("plain text log time is invalid: %v; line: %q", err, line)
	}
	if parts[1] != "INFO" || parts[2] != "backend started" {
		t.Fatalf("unexpected plain text log: %q", line)
	}
	if strings.Contains(parts[0], ".") {
		t.Fatalf("plain text log time contains fractional seconds: %q", line)
	}
	if strings.Contains(line, "time=") || strings.Contains(line, "level=") || strings.Contains(line, "msg=") {
		t.Fatalf("plain text log contains field names: %q", line)
	}
}

func TestInstallAppendsToLogFile(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "txt",
		Dir:    "log",
		File:   "cephtower.log",
		Output: "file",
	}

	logger, closeLog, err := Install(cfg, workDir)
	if err != nil {
		t.Fatalf("Install() returned error: %v", err)
	}
	logger.Info("first entry")
	if err := closeLog(); err != nil {
		t.Fatalf("closeLog() returned error: %v", err)
	}

	logger, closeLog, err = Install(cfg, workDir)
	if err != nil {
		t.Fatalf("second Install() returned error: %v", err)
	}
	logger.Info("second entry")
	if err := closeLog(); err != nil {
		t.Fatalf("second closeLog() returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workDir, "log", "cephtower.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "first entry") || !strings.Contains(string(data), "second entry") {
		t.Fatalf("log file was not appended: %q", data)
	}
}

func TestRotatingFileWriterCreatesTimestampedHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log", "cephtower.log")
	writer, err := newRotatingFileWriter(path, "1day", "7days")
	if err != nil {
		t.Fatalf("newRotatingFileWriter() returned error: %v", err)
	}
	if _, err := writer.Write([]byte("first\n")); err != nil {
		t.Fatalf("first Write() returned error: %v", err)
	}

	writer.nextRotationAt = writer.now().Add(-time.Second)
	if _, err := writer.Write([]byte("second\n")); err != nil {
		t.Fatalf("second Write() returned error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	if string(current) != "second\n" {
		t.Fatalf("current log = %q, want second entry", current)
	}
	history, err := filepath.Glob(filepath.Join(filepath.Dir(path), "cephtower-*.log"))
	if err != nil || len(history) != 1 {
		t.Fatalf("history files = %v, want one timestamped file", history)
	}
	if len(filepath.Base(history[0])) != len("cephtower-20260305122456.log") {
		t.Fatalf("history file has unexpected name: %q", filepath.Base(history[0]))
	}
}

func TestNextRotationAtUsesMidnight(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	start := time.Date(2026, 3, 5, 5, 0, 0, 0, location)
	want := time.Date(2026, 3, 12, 0, 0, 0, 0, location)
	if got := nextRotationAt(start, 7); !got.Equal(want) {
		t.Fatalf("nextRotationAt() = %s, want %s", got, want)
	}
}

func TestRotatingFileWriterRemovesExpiredHistory(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "cephtower-20200101000000.log")
	if err := os.WriteFile(oldPath, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write old history: %v", err)
	}

	writer, err := newRotatingFileWriter(filepath.Join(dir, "cephtower.log"), "7days", "28days")
	if err != nil {
		t.Fatalf("newRotatingFileWriter() returned error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
	writer.runCleanup(time.Now())
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expired history still exists, stat error: %v", err)
	}
}

func TestNewLoggerHonorsConfiguredLevel(t *testing.T) {
	var output bytes.Buffer
	logger, err := NewLogger(config.LoggingConfig{Level: "warn", Format: "txt"}, &output)
	if err != nil {
		t.Fatalf("NewLogger() returned error: %v", err)
	}

	logger.Info("hidden")
	logger.Warn("visible")

	logOutput := output.String()
	if strings.Contains(logOutput, "hidden") {
		t.Fatalf("info log was not filtered: %q", logOutput)
	}
	if !strings.Contains(logOutput, "visible") {
		t.Fatalf("warn log missing from output: %q", logOutput)
	}
}

func TestNewLoggerRejectsUnsupportedFormat(t *testing.T) {
	if _, err := NewLogger(config.LoggingConfig{Level: "info", Format: "xml"}, nil); err == nil {
		t.Fatal("NewLogger() returned nil error, want unsupported format error")
	}
}
