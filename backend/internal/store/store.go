package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/gorm"

	"cephtower/backend/internal/config"
	mysqlstore "cephtower/backend/internal/store/mysql"
	sqlitestore "cephtower/backend/internal/store/sqlite"
)

// ErrRecordNotFound is returned when a single-record lookup has no result.
var ErrRecordNotFound = gorm.ErrRecordNotFound

// Database is the project-owned persistence handle. GORM never crosses this
// package boundary; callers use domain operations defined in the Store files.
type Database struct {
	db      *gorm.DB
	auditMu *sync.Mutex
}

func wrap(db *gorm.DB) *Database {
	if db == nil {
		return nil
	}
	return &Database{db: db, auditMu: &sync.Mutex{}}
}
func (d *Database) Transaction(fn func(*Database) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error { return fn(&Database{db: tx, auditMu: d.auditMu}) })
}

// Insert, FindRecord and CountRecords are small Store operations used for
// polymorphic discovery models and black-box persistence tests. They expose no
// ORM query object outside this package.
func (d *Database) Insert(ctx context.Context, value any) error {
	return d.db.WithContext(ctx).Create(value).Error
}
func (d *Database) FindRecord(ctx context.Context, filters map[string]any, dest any) error {
	query := d.db.WithContext(ctx)
	for field, value := range filters {
		query = query.Where(field+" = ?", value)
	}
	return query.First(dest).Error
}
func (d *Database) CountRecords(ctx context.Context, model any, filters map[string]any) (int64, error) {
	query := d.db.WithContext(ctx).Model(model)
	for field, value := range filters {
		query = query.Where(field+" = ?", value)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

const (
	EngineSQLite = "sqlite"
	EngineMySQL  = "mysql"
)

func Open(cfg config.DatabaseConfig, workDirs ...string) (*Database, error) {
	workDir := databaseWorkDir(workDirs)
	dialector, err := dialector(cfg, workDir)
	if err != nil {
		return nil, err
	}

	db, sqlDB, err := openDatabaseHandle(cfg.Engine, dialector)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate database schema: %w", err)
	}

	return wrap(db), nil
}

// OpenExisting connects to a database that must already exist. It intentionally
// does not create SQLite files, create MySQL databases, or run migrations.
func OpenExisting(cfg config.DatabaseConfig, workDirs ...string) (*Database, error) {
	workDir := databaseWorkDir(workDirs)
	if cfg.Engine == EngineSQLite {
		if err := requireSQLiteDatabase(cfg.SQLite, workDir); err != nil {
			return nil, err
		}
	}
	dialector, err := dialector(cfg, workDir)
	if err != nil {
		return nil, err
	}
	db, _, err := openDatabaseHandle(cfg.Engine, dialector)
	if err != nil {
		return nil, err
	}
	return wrap(db), nil
}

// Initialize creates the configured database container when needed and applies
// the baseline schema. Service startup should use OpenExisting instead.
func Initialize(ctx context.Context, cfg config.DatabaseConfig, workDirs ...string) (*Database, error) {
	if cfg.Engine == EngineMySQL {
		if err := mysqlstore.EnsureDatabase(ctx, cfg.MySQL); err != nil {
			return nil, err
		}
	}
	return Open(cfg, workDirs...)
}

func openDatabaseHandle(engine string, dialector gorm.Dialector) (*gorm.DB, interface{ Close() error }, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("open %s database: %w", engine, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql database handle: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, nil, fmt.Errorf("ping %s database: %w", engine, err)
	}
	if engine == EngineSQLite {
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			_ = sqlDB.Close()
			return nil, nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
		}
	}
	return db, sqlDB, nil
}

func requireSQLiteDatabase(cfg config.SQLiteConfig, workDir string) error {
	path := sqlitestore.ResolveName(cfg.Name, workDir)
	if path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sqlite database file %q does not exist", path)
		}
		return fmt.Errorf("inspect sqlite database file %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("sqlite database path %q is a directory", path)
	}
	return nil
}

// TestConnection verifies database connectivity without running migrations.
func TestConnection(ctx context.Context, cfg config.DatabaseConfig, workDirs ...string) error {
	workDir := databaseWorkDir(workDirs)
	if cfg.Engine == EngineSQLite {
		if err := requireSQLiteDatabase(cfg.SQLite, workDir); err != nil {
			return err
		}
	}
	dialector, err := dialector(cfg, workDir)
	if err != nil {
		return err
	}

	db, err := gorm.Open(dialector, &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		return fmt.Errorf("open %s database: %w", cfg.Engine, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql database handle: %w", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping %s database: %w", cfg.Engine, err)
	}
	return nil
}

func TestInitializationTarget(ctx context.Context, cfg config.DatabaseConfig, workDirs ...string) error {
	workDir := databaseWorkDir(workDirs)
	switch cfg.Engine {
	case EngineSQLite:
		return testSQLiteInitializationTarget(cfg.SQLite, workDir)
	case EngineMySQL:
		return mysqlstore.TestInitializationTarget(ctx, cfg.MySQL)
	default:
		return fmt.Errorf("unsupported database engine %q", cfg.Engine)
	}
}

func testSQLiteInitializationTarget(cfg config.SQLiteConfig, workDir string) error {
	path := sqlitestore.ResolveName(cfg.Name, workDir)
	if path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("sqlite database path %q is a directory", path)
		}
		return fmt.Errorf("sqlite database file %q already exists", path)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect sqlite database file %q: %w", path, err)
	}
	return testSQLiteDirectoryWritable(filepath.Dir(path))
}

func testSQLiteDirectoryWritable(dir string) error {
	if dir == "" {
		dir = "."
	}
	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("sqlite data path %q is not a directory", dir)
		}
		return testWriteFile(dir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect sqlite data directory %q: %w", dir, err)
	}
	ancestor, err := nearestExistingDirectory(dir)
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp(ancestor, ".cephtower-dbtest-*")
	if err != nil {
		return fmt.Errorf("sqlite data directory %q cannot be created: %w", dir, err)
	}
	if err := os.Remove(tempDir); err != nil {
		return fmt.Errorf("remove sqlite data directory probe %q: %w", tempDir, err)
	}
	return nil
}

func nearestExistingDirectory(path string) (string, error) {
	current := filepath.Clean(path)
	for {
		info, err := os.Stat(current)
		if err == nil {
			if !info.IsDir() {
				return "", fmt.Errorf("sqlite data path %q is not a directory", current)
			}
			return current, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect sqlite data path %q: %w", current, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("sqlite data directory %q cannot be created", path)
		}
		current = parent
	}
}

func testWriteFile(dir string) error {
	file, err := os.CreateTemp(dir, ".cephtower-dbtest-*")
	if err != nil {
		return fmt.Errorf("sqlite data directory %q is not writable: %w", dir, err)
	}
	name := file.Name()
	closeErr := file.Close()
	removeErr := os.Remove(name)
	if closeErr != nil {
		return fmt.Errorf("close sqlite data directory probe %q: %w", name, closeErr)
	}
	if removeErr != nil {
		return fmt.Errorf("remove sqlite data directory probe %q: %w", name, removeErr)
	}
	return nil
}

func databaseWorkDir(workDirs []string) string {
	if len(workDirs) > 0 && workDirs[0] != "" {
		return workDirs[0]
	}
	return "./app"
}

func Close(db *Database) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("get sql database handle: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}

func dialector(cfg config.DatabaseConfig, workDir string) (gorm.Dialector, error) {
	switch cfg.Engine {
	case EngineSQLite:
		return sqlitestore.Dialector(cfg.SQLite, workDir)
	case EngineMySQL:
		return mysqlstore.Dialector(cfg.MySQL)
	default:
		return nil, fmt.Errorf("unsupported database engine %q", cfg.Engine)
	}
}
