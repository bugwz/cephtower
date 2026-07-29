package mysql

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	mysqldriver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"cephtower/backend/internal/config"
)

func Dialector(cfg config.MySQLConfig) (gorm.Dialector, error) {
	dsn, err := DSN(cfg)
	if err != nil {
		return nil, err
	}
	return gormmysql.Open(dsn), nil
}

func EnsureDatabase(ctx context.Context, cfg config.MySQLConfig) error {
	dsn, err := ServerDSN(cfg)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql server: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql server: %w", err)
	}
	_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+quoteIdentifier(cfg.Database))
	if err != nil {
		return fmt.Errorf("create mysql database %q: %w", cfg.Database, err)
	}
	return nil
}

func TestInitializationTarget(ctx context.Context, cfg config.MySQLConfig) error {
	if strings.TrimSpace(cfg.Database) == "" {
		return fmt.Errorf("mysql database is required")
	}
	dsn, err := ServerDSN(cfg)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql server: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql server: %w", err)
	}
	exists, err := databaseExists(ctx, db, cfg.Database)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("mysql database %q already exists", cfg.Database)
	}
	tempName, err := temporaryDatabaseName(cfg.Database)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+quoteIdentifier(tempName)); err != nil {
		return fmt.Errorf("create temporary mysql database: %w", err)
	}
	if _, err := db.ExecContext(ctx, "DROP DATABASE "+quoteIdentifier(tempName)); err != nil {
		return fmt.Errorf("drop temporary mysql database %q: %w", tempName, err)
	}
	return nil
}

func databaseExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var existing string
	err := db.QueryRowContext(ctx, "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", strings.TrimSpace(name)).Scan(&existing)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect mysql database %q: %w", name, err)
	}
	return true, nil
}

func temporaryDatabaseName(target string) (string, error) {
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate temporary mysql database name: %w", err)
	}
	base := strings.TrimSpace(target)
	if base == "" {
		base = "cephtower"
	}
	base = strings.NewReplacer("`", "_", "-", "_", ".", "_").Replace(base)
	if len(base) > 32 {
		base = base[:32]
	}
	return base + "_dbtest_" + hex.EncodeToString(random), nil
}

func DSN(cfg config.MySQLConfig) (string, error) {
	if strings.TrimSpace(cfg.Username) == "" {
		return "", fmt.Errorf("mysql username is required")
	}
	if strings.TrimSpace(cfg.Database) == "" {
		return "", fmt.Errorf("mysql database is required")
	}

	params, err := params(cfg.Params)
	if err != nil {
		return "", err
	}

	mysqlConfig := mysqldriver.NewConfig()
	mysqlConfig.User = cfg.Username
	mysqlConfig.Passwd = cfg.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	mysqlConfig.DBName = cfg.Database
	mysqlConfig.Params = params
	mysqlConfig.TLSConfig = cfg.TLS

	return mysqlConfig.FormatDSN(), nil
}

func ServerDSN(cfg config.MySQLConfig) (string, error) {
	if strings.TrimSpace(cfg.Username) == "" {
		return "", fmt.Errorf("mysql username is required")
	}

	params, err := params(cfg.Params)
	if err != nil {
		return "", err
	}

	mysqlConfig := mysqldriver.NewConfig()
	mysqlConfig.User = cfg.Username
	mysqlConfig.Passwd = cfg.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	mysqlConfig.Params = params
	mysqlConfig.TLSConfig = cfg.TLS

	return mysqlConfig.FormatDSN(), nil
}

func quoteIdentifier(value string) string {
	return "`" + strings.ReplaceAll(strings.TrimSpace(value), "`", "``") + "`"
}

func params(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, fmt.Errorf("parse mysql params: %w", err)
	}

	params := make(map[string]string, len(values))
	for key, value := range values {
		if len(value) == 0 {
			params[key] = ""
			continue
		}
		params[key] = value[len(value)-1]
	}
	return params, nil
}
