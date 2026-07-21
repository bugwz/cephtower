package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Path        string
	HTTPAddr    string
	Logging     LoggingConfig
	Ceph        CephDashboardConfig
	CephCommand CephCommandConfig
	Database    DatabaseConfig
	SMTP        SMTPConfig
}

type LoggingConfig struct {
	Level  string
	Format string
}

type CephDashboardConfig struct {
	BaseURL     string
	Username    string
	Password    string
	InsecureTLS bool
}

type CephCommandConfig struct {
	Bin     string
	Cluster string
	Conf    string
	Name    string
	Keyring string
	Timeout time.Duration
}

type DatabaseConfig struct {
	Engine string
	SQLite SQLiteConfig
	MySQL  MySQLConfig
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
	HTTPAddr string `yaml:"http_addr"`
	Logging  struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Ceph struct {
		BaseURL     string `yaml:"base_url"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		InsecureTLS bool   `yaml:"insecure_tls"`
	} `yaml:"ceph_dashboard"`
	CephCommand struct {
		Bin     string `yaml:"bin"`
		Cluster string `yaml:"cluster"`
		Conf    string `yaml:"conf"`
		Name    string `yaml:"name"`
		Keyring string `yaml:"keyring"`
		Timeout string `yaml:"timeout"`
	} `yaml:"ceph_command,omitempty"`
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

	httpAddr := strings.TrimSpace(raw.HTTPAddr)
	if httpAddr == "" {
		httpAddr = ":36900"
	}

	database, err := normalizeDatabaseConfig(raw)
	if err != nil {
		return Config{}, err
	}
	cephCommand, err := normalizeCephCommandConfig(raw)
	if err != nil {
		return Config{}, err
	}
	logging, err := normalizeLoggingConfig(raw)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Path:     path,
		HTTPAddr: httpAddr,
		Logging:  logging,
		Ceph: CephDashboardConfig{
			BaseURL:     strings.TrimRight(strings.TrimSpace(raw.Ceph.BaseURL), "/"),
			Username:    raw.Ceph.Username,
			Password:    raw.Ceph.Password,
			InsecureTLS: raw.Ceph.InsecureTLS,
		},
		CephCommand: cephCommand,
		Database:    database,
		SMTP: SMTPConfig{
			Host:     strings.TrimSpace(raw.SMTP.Host),
			Port:     raw.SMTP.Port,
			Username: strings.TrimSpace(raw.SMTP.Username),
			Password: raw.SMTP.Password,
			From:     strings.TrimSpace(raw.SMTP.From),
		},
	}, nil
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

	return LoggingConfig{
		Level:  level,
		Format: format,
	}, nil
}

func normalizeCephCommandConfig(raw fileConfig) (CephCommandConfig, error) {
	bin := strings.TrimSpace(raw.CephCommand.Bin)
	if bin == "" {
		bin = "ceph"
	}

	timeout := 15 * time.Second
	if value := strings.TrimSpace(raw.CephCommand.Timeout); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return CephCommandConfig{}, fmt.Errorf("invalid ceph_command timeout %q: %w", value, err)
		}
		if parsed <= 0 {
			return CephCommandConfig{}, fmt.Errorf("invalid ceph_command timeout %q: must be positive", value)
		}
		timeout = parsed
	}

	return CephCommandConfig{
		Bin:     bin,
		Cluster: strings.TrimSpace(raw.CephCommand.Cluster),
		Conf:    strings.TrimSpace(raw.CephCommand.Conf),
		Name:    strings.TrimSpace(raw.CephCommand.Name),
		Keyring: strings.TrimSpace(raw.CephCommand.Keyring),
		Timeout: timeout,
	}, nil
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
		sqlitePath = "data/cephtower.db"
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
