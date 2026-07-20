package mysql

import (
	"strings"
	"testing"

	"cephtower/backend/internal/config"
)

func TestDSNRequiresDatabase(t *testing.T) {
	_, err := DSN(config.MySQLConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "cephtower",
	})
	if err == nil {
		t.Fatal("DSN() returned nil error, want missing database error")
	}
}

func TestDSNFormatsConfiguredParams(t *testing.T) {
	dsn, err := DSN(config.MySQLConfig{
		Host:     "db.example.com",
		Port:     3307,
		Username: "cephtower",
		Password: "secret",
		Database: "cephtower",
		Params:   "charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai",
	})
	if err != nil {
		t.Fatalf("DSN() returned error: %v", err)
	}

	wantParts := []string{
		"cephtower:secret@tcp(db.example.com:3307)/cephtower?",
		"charset=utf8mb4",
		"parseTime=True",
		"loc=Asia%2FShanghai",
	}
	for _, part := range wantParts {
		if !strings.Contains(dsn, part) {
			t.Fatalf("DSN %q does not contain %q", dsn, part)
		}
	}
}
