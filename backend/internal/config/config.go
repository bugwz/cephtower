package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPAddr string
	Ceph     CephDashboardConfig
	Database DatabaseConfig
}

type CephDashboardConfig struct {
	BaseURL     string
	Username    string
	Password    string
	InsecureTLS bool
}

type DatabaseConfig struct {
	Engine string
	SQLite SQLiteConfig
	MySQL  MySQLConfig
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
	Ceph     struct {
		BaseURL     string `yaml:"base_url"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		InsecureTLS bool   `yaml:"insecure_tls"`
	} `yaml:"ceph_dashboard"`
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

	return Config{
		HTTPAddr: httpAddr,
		Ceph: CephDashboardConfig{
			BaseURL:     strings.TrimRight(strings.TrimSpace(raw.Ceph.BaseURL), "/"),
			Username:    raw.Ceph.Username,
			Password:    raw.Ceph.Password,
			InsecureTLS: raw.Ceph.InsecureTLS,
		},
		Database: database,
	}, nil
}

func normalizeDatabaseConfig(raw fileConfig) (DatabaseConfig, error) {
	engine := strings.ToLower(strings.TrimSpace(raw.Database.Engine))
	if engine == "" {
		engine = "sqlite"
	}
	if engine != "sqlite" && engine != "mysql" {
		return DatabaseConfig{}, fmt.Errorf("unsupported database engine %q", raw.Database.Engine)
	}

	sqlitePath := strings.TrimSpace(raw.Database.SQLite.Path)
	if sqlitePath == "" {
		sqlitePath = "data/cephtower.db"
	}

	mysqlHost := strings.TrimSpace(raw.Database.MySQL.Host)
	if mysqlHost == "" {
		mysqlHost = "127.0.0.1"
	}

	mysqlPort := raw.Database.MySQL.Port
	if mysqlPort == 0 {
		mysqlPort = 3306
	}

	mysqlParams := strings.TrimSpace(raw.Database.MySQL.Params)
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
			Username: raw.Database.MySQL.Username,
			Password: raw.Database.MySQL.Password,
			Database: raw.Database.MySQL.Database,
			Params:   mysqlParams,
		},
	}, nil
}
