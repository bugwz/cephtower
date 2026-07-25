package store

import (
	"context"
	"fmt"

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
	db *gorm.DB
}

func wrap(db *gorm.DB) *Database {
	if db == nil {
		return nil
	}
	return &Database{db: db}
}
func (d *Database) Transaction(fn func(*Database) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error { return fn(wrap(tx)) })
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
	return count, query.Count(&count).Error
}

const (
	EngineSQLite = "sqlite"
	EngineMySQL  = "mysql"
)

func Open(cfg config.DatabaseConfig, workDirs ...string) (*Database, error) {
	workDir := "./app"
	if len(workDirs) > 0 && workDirs[0] != "" {
		workDir = workDirs[0]
	}
	dialector, err := dialector(cfg, workDir)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", cfg.Engine, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql database handle: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping %s database: %w", cfg.Engine, err)
	}

	if err := migrate(db); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate database schema: %w", err)
	}

	return wrap(db), nil
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
