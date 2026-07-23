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

func Dialector(cfg config.SQLiteConfig, workDir string) (gorm.Dialector, error) {
	path := resolvePath(cfg.Path, workDir)
	if err := ensureDirectory(path); err != nil {
		return nil, err
	}
	return sqlite.Open(path), nil
}

func resolvePath(path, workDir string) string {
	if path == "" || path == ":memory:" || strings.HasPrefix(path, "file:") || filepath.IsAbs(path) {
		return path
	}
	if workDir == "" {
		workDir = "."
	}
	return filepath.Join(workDir, path)
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
