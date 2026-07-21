package logging

import (
	"bytes"
	"encoding/json"
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
