package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Path     string
	Server   ServerConfig
	Logging  LoggingConfig
	Database DatabaseConfig
	Runtime  RuntimeConfig
	SMTP     SMTPConfig
}

type ServerConfig struct {
	Address string
	Port    int
	Dir     string
}

type LoggingConfig struct {
	Dir       string
	File      string
	Output    string
	Level     string
	Format    string
	Rotation  string
	Retention string
}

type DatabaseConfig struct {
	Engine string
	SQLite SQLiteConfig
	MySQL  MySQLConfig
}

type RuntimeConfig struct {
	Dir string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SQLiteConfig struct {
	Path string
}

type MySQLConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Params   string
}

type fileConfig struct {
	Server struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
		Dir     string `yaml:"dir"`
	} `yaml:"server"`
	Logging struct {
		Dir       string `yaml:"dir"`
		File      string `yaml:"file"`
		Path      string `yaml:"path"` // backward compatibility with older configs
		Output    string `yaml:"output"`
		Level     string `yaml:"level"`
		Format    string `yaml:"format"`
		Rotation  string `yaml:"rotation"`
		Retention string `yaml:"retention"`
	} `yaml:"log"`
	Database struct {
		Engine string `yaml:"engine"`
		SQLite struct {
			Path string `yaml:"path"`
		} `yaml:"sqlite"`
		MySQL struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Database string `yaml:"database"`
			Params   string `yaml:"params"`
		} `yaml:"mysql"`
	} `yaml:"database"`
	Runtime struct {
		Dir string `yaml:"dir"`
	} `yaml:"runtime"`
	SMTP struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		From     string `yaml:"from"`
	} `yaml:"smtp"`
}

func Load(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		path = "config/config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	var raw fileConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	server := normalizeServerConfig(raw)
	if server.Dir == "" {
		server.Dir = "./app"
	}
	if server.Address == "" {
		server.Address = "0.0.0.0"
	}
	if server.Port == 0 {
		server.Port = 36900
	}

	database, err := normalizeDatabaseConfig(raw)
	if err != nil {
		return Config{}, err
	}
	logging, err := normalizeLoggingConfig(raw)
	if err != nil {
		return Config{}, err
	}
	runtime := normalizeRuntimeConfig(raw)

	return Config{
		Path:     path,
		Server:   server,
		Logging:  logging,
		Database: database,
		Runtime:  runtime,
		SMTP: SMTPConfig{
			Host:     strings.TrimSpace(raw.SMTP.Host),
			Port:     raw.SMTP.Port,
			Username: strings.TrimSpace(raw.SMTP.Username),
			Password: raw.SMTP.Password,
			From:     strings.TrimSpace(raw.SMTP.From),
		},
	}, nil
}

func normalizeRuntimeConfig(raw fileConfig) RuntimeConfig {
	dir := strings.TrimSpace(raw.Runtime.Dir)
	if dir == "" {
		dir = "data/runtime"
	}
	return RuntimeConfig{Dir: dir}
}

func ResolveRuntimeDir(cfg Config) string {
	dir := strings.TrimSpace(cfg.Runtime.Dir)
	if dir == "" {
		dir = "data/runtime"
	}
	if filepath.IsAbs(dir) {
		return dir
	}
	workDir := cfg.Server.Dir
	if workDir == "" {
		workDir = "."
	}
	return filepath.Join(workDir, dir)
}

func normalizeServerConfig(raw fileConfig) ServerConfig {
	return ServerConfig{
		Dir:     strings.TrimSpace(raw.Server.Dir),
		Address: strings.TrimSpace(raw.Server.Address),
		Port:    raw.Server.Port,
	}
}

func SaveDatabase(path string, database DatabaseConfig) error {
	if strings.TrimSpace(path) == "" {
		path = "config/config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}

	var raw fileConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse config file %q: %w", path, err)
	}

	raw.Database.Engine = database.Engine
	raw.Database.SQLite.Path = database.SQLite.Path
	raw.Database.MySQL.Host = database.MySQL.Host
	raw.Database.MySQL.Port = database.MySQL.Port
	raw.Database.MySQL.Username = database.MySQL.Username
	raw.Database.MySQL.Password = database.MySQL.Password
	raw.Database.MySQL.Database = database.MySQL.Database
	raw.Database.MySQL.Params = database.MySQL.Params

	normalized, err := normalizeDatabaseConfig(raw)
	if err != nil {
		return err
	}
	raw.Database.Engine = normalized.Engine
	raw.Database.SQLite.Path = normalized.SQLite.Path
	raw.Database.MySQL.Host = normalized.MySQL.Host
	raw.Database.MySQL.Port = normalized.MySQL.Port
	raw.Database.MySQL.Username = normalized.MySQL.Username
	raw.Database.MySQL.Password = normalized.MySQL.Password
	raw.Database.MySQL.Database = normalized.MySQL.Database
	raw.Database.MySQL.Params = normalized.MySQL.Params

	output, err := yaml.Marshal(&raw)
	if err != nil {
		return fmt.Errorf("marshal config file %q: %w", path, err)
	}
	if err := os.WriteFile(path, output, 0o600); err != nil {
		return fmt.Errorf("write config file %q: %w", path, err)
	}
	return nil
}

func normalizeLoggingConfig(raw fileConfig) (LoggingConfig, error) {
	level := strings.ToLower(strings.TrimSpace(raw.Logging.Level))
	if level == "" {
		level = "info"
	}
	switch level {
	case "debug", "info", "warn", "error":
	default:
		return LoggingConfig{}, fmt.Errorf("unsupported logging level %q", raw.Logging.Level)
	}

	format := strings.ToLower(strings.TrimSpace(raw.Logging.Format))
	if format == "" {
		format = "txt"
	}
	switch format {
	case "txt", "json":
	default:
		return LoggingConfig{}, fmt.Errorf("unsupported logging format %q", raw.Logging.Format)
	}

	dir := strings.TrimSpace(raw.Logging.Dir)
	file := strings.TrimSpace(raw.Logging.File)
	if dir == "" && file == "" && strings.TrimSpace(raw.Logging.Path) != "" {
		legacyPath := strings.TrimSpace(raw.Logging.Path)
		dir = filepath.Dir(legacyPath)
		file = filepath.Base(legacyPath)
	}
	if dir == "" {
		dir = "log"
	}
	if file == "" {
		file = "cephtower.log"
	}

	output := strings.ToLower(strings.TrimSpace(raw.Logging.Output))
	if output == "" {
		output = "both"
	}
	switch output {
	case "stdout", "file", "both":
	default:
		return LoggingConfig{}, fmt.Errorf("unsupported logging output %q", raw.Logging.Output)
	}

	rotation := strings.ToLower(strings.TrimSpace(raw.Logging.Rotation))
	if rotation == "" {
		rotation = "7days"
	}
	if _, err := ParseDuration(rotation); err != nil {
		return LoggingConfig{}, fmt.Errorf("invalid logging rotation %q: %w", raw.Logging.Rotation, err)
	}

	retention := strings.ToLower(strings.TrimSpace(raw.Logging.Retention))
	if retention == "" {
		retention = "70days"
	}
	if _, err := ParseDuration(retention); err != nil {
		return LoggingConfig{}, fmt.Errorf("invalid logging retention %q: %w", raw.Logging.Retention, err)
	}

	return LoggingConfig{
		Level:     level,
		Format:    format,
		Dir:       dir,
		File:      file,
		Output:    output,
		Rotation:  rotation,
		Retention: retention,
	}, nil
}

// ParseDuration parses whole-day durations such as 7days, 2weeks, and 1month.
func ParseDuration(value string) (time.Duration, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	var amount int64
	var unit string
	if _, err := fmt.Sscanf(value, "%d%s", &amount, &unit); err != nil || amount <= 0 {
		return 0, fmt.Errorf("must be a positive number followed by a time unit")
	}

	var multiplier time.Duration
	switch unit {
	case "day", "days":
		multiplier = 24 * time.Hour
	case "week", "weeks":
		multiplier = 7 * 24 * time.Hour
	case "month", "months":
		multiplier = 30 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("unsupported time unit %q", unit)
	}
	if amount > int64((time.Duration(1<<63-1))/multiplier) {
		return 0, fmt.Errorf("duration is too large")
	}
	return time.Duration(amount) * multiplier, nil
}

func normalizeDatabaseConfig(raw fileConfig) (DatabaseConfig, error) {
	return NormalizeDatabaseConfig(DatabaseConfig{
		Engine: raw.Database.Engine,
		SQLite: SQLiteConfig{
			Path: raw.Database.SQLite.Path,
		},
		MySQL: MySQLConfig{
			Host:     raw.Database.MySQL.Host,
			Port:     raw.Database.MySQL.Port,
			Username: raw.Database.MySQL.Username,
			Password: raw.Database.MySQL.Password,
			Database: raw.Database.MySQL.Database,
			Params:   raw.Database.MySQL.Params,
		},
	})
}

func NormalizeDatabaseConfig(cfg DatabaseConfig) (DatabaseConfig, error) {
	engine := strings.ToLower(strings.TrimSpace(cfg.Engine))
	if engine == "" {
		engine = "sqlite"
	}
	if engine != "sqlite" && engine != "mysql" {
		return DatabaseConfig{}, fmt.Errorf("unsupported database engine %q", cfg.Engine)
	}

	sqlitePath := strings.TrimSpace(cfg.SQLite.Path)
	if sqlitePath == "" {
		sqlitePath = "data/db/cephtower.db"
	}

	mysqlHost := strings.TrimSpace(cfg.MySQL.Host)
	if mysqlHost == "" {
		mysqlHost = "127.0.0.1"
	}

	mysqlPort := cfg.MySQL.Port
	if mysqlPort == 0 {
		mysqlPort = 3306
	}

	mysqlParams := strings.TrimSpace(cfg.MySQL.Params)
	if mysqlParams == "" {
		mysqlParams = "charset=utf8mb4&parseTime=True&loc=Local"
	}

	return DatabaseConfig{
		Engine: engine,
		SQLite: SQLiteConfig{
			Path: sqlitePath,
		},
		MySQL: MySQLConfig{
			Host:     mysqlHost,
			Port:     mysqlPort,
			Username: strings.TrimSpace(cfg.MySQL.Username),
			Password: cfg.MySQL.Password,
			Database: strings.TrimSpace(cfg.MySQL.Database),
			Params:   mysqlParams,
		},
	}, nil
}
