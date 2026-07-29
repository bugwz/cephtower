package sqlite

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"cephtower/backend/internal/config"
)

func Dialector(cfg config.SQLiteConfig, workDir string) (gorm.Dialector, error) {
	path := resolveName(cfg.Name, workDir)
	if err := ensureDirectory(path); err != nil {
		return nil, err
	}
	return sqlite.Open(path), nil
}

func ResolveName(name, workDir string) string {
	return resolveName(name, workDir)
}

func resolveName(name, workDir string) string {
	if name == "" {
		name = "cephtower.db"
	}
	if workDir == "" {
		workDir = "."
	}
	return filepath.Join(workDir, "data", "db", name)
}

func ensureDirectory(path string) error {
	if path == "" {
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
