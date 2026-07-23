package store

import (
	"fmt"

	"gorm.io/gorm"

	"cephtower/backend/internal/config"
	mysqlstore "cephtower/backend/internal/store/mysql"
	sqlitestore "cephtower/backend/internal/store/sqlite"
)

const (
	EngineSQLite = "sqlite"
	EngineMySQL  = "mysql"
)

func Open(cfg config.DatabaseConfig, workDirs ...string) (*gorm.DB, error) {
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

	if err := db.AutoMigrate(
		&Setting{},
		&CephCluster{},
		&CephClusterHost{},
		&CephClusterOSD{},
		&CephClusterOSDFlag{},
		&CephClusterDaemon{},
		&CephClusterService{},
		&CephClusterMon{},
		&CephClusterMgr{},
		&CephClusterMDS{},
		&CephClusterMgrModule{},
		&CephClusterConfiguration{},
		&CephDataFetchRun{},
		&CephClusterSummary{},
		&CephClusterHealthCheck{},
		&CephPool{},
		&CephRBDImage{},
		&CephFilesystem{},
		&CephRGWDaemon{},
		&CephRGWUser{},
		&CephRGWBucket{},
		&CephNVMeoFGateway{},
		&CephNVMeoFSubsystem{},
		&CephISCSITarget{},
		&CephNFSExport{},
		&User{},
		&PasswordResetCode{},
		&UserSession{},
	); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate database schema: %w", err)
	}

	return db, nil
}

func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
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
