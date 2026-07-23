package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	if entry["time"] == "" || entry["level"] != "INFO" || entry["msg"] != "backend started" {
		t.Fatalf("unexpected json log entry: %#v", entry)
	}
}

func TestInstallAppendsToLogFile(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "txt",
		Path:   "log/cephtower.log",
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
