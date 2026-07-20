package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"cephtower/backend/internal/config"
)

func Dialector(cfg config.SQLiteConfig) (gorm.Dialector, error) {
	if err := ensureDirectory(cfg.Path); err != nil {
		return nil, err
	}
	return sqlite.Open(cfg.Path), nil
}

func ensureDirectory(path string) error {
	if path == "" || path == ":memory:" || strings.HasPrefix(path, "file:") {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create sqlite data directory %q: %w", dir, err)
	}
	return nil
}
