package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadReadsConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`server:
  address: 127.0.0.1
  port: 9090
log:
  dir: custom-log
  file: application.log
  level: debug
  format: json
  rotation: 2weeks
  retention: 3days
database:
  engine: mysql
  mysql:
    host: db.example.com
    port: 3307
    username: cephtower
    password: db-secret
    database: cephtower
    params: charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai
    tls: true
runtime:
  dir: data/custom-runtime
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Address != "127.0.0.1" || cfg.Server.Port != 9090 {
		t.Fatalf("Server = %#v, want 127.0.0.1:9090", cfg.Server)
	}
	if cfg.Server.Dir != DefaultServerDir {
		t.Fatalf("Server.Dir = %q, want %s", cfg.Server.Dir, DefaultServerDir)
	}
	if cfg.Runtime.Dir != "data/custom-runtime" {
		t.Fatalf("Runtime.Dir = %q, want data/custom-runtime", cfg.Runtime.Dir)
	}
	if cfg.Path != path {
		t.Fatalf("Path = %q, want %q", cfg.Path, path)
	}
	if cfg.Logging.Level != "debug" || cfg.Logging.Format != "json" || cfg.Logging.Dir != "custom-log" || cfg.Logging.File != "application.log" || cfg.Logging.Output != "both" || cfg.Logging.Rotation != "2weeks" || cfg.Logging.Retention != "3days" {
		t.Fatalf("Logging = %#v, want debug/json", cfg.Logging)
	}
	if cfg.Database.Engine != "mysql" {
		t.Fatalf("Database.Engine = %q, want mysql", cfg.Database.Engine)
	}
	if cfg.Database.MySQL.Host != "db.example.com" || cfg.Database.MySQL.Port != 3307 {
		t.Fatalf("unexpected MySQL address: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Username != "cephtower" || cfg.Database.MySQL.Password != "db-secret" {
		t.Fatalf("unexpected MySQL credentials: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Database != "cephtower" {
		t.Fatalf("Database.MySQL.Database = %q, want cephtower", cfg.Database.MySQL.Database)
	}
	if cfg.Database.MySQL.Params != "charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai" {
		t.Fatalf("Database.MySQL.Params = %q, want configured params", cfg.Database.MySQL.Params)
	}
	if cfg.Database.MySQL.TLS != "true" {
		t.Fatalf("Database.MySQL.TLS = %q, want true", cfg.Database.MySQL.TLS)
	}
}

func TestLoadRequiresExistingConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "config.yaml")

	if _, err := Load(path); err == nil {
		t.Fatal("Load() returned nil error for a missing config file")
	}
}

func TestLoadRequiresConfigPath(t *testing.T) {
	if _, err := Load(""); err == nil {
		t.Fatal("Load() returned nil error for an empty config path")
	}
}

func TestParseDurationSupportsSingularAndPluralUnits(t *testing.T) {
	tests := map[string]time.Duration{
		"1day":   24 * time.Hour,
		"7days":  7 * 24 * time.Hour,
		"1week":  7 * 24 * time.Hour,
		"2weeks": 14 * 24 * time.Hour,
		"1month": 30 * 24 * time.Hour,
	}
	for value, want := range tests {
		got, err := ParseDuration(value)
		if err != nil || got != want {
			t.Errorf("ParseDuration(%q) = %v, %v; want %v", value, got, err, want)
		}
	}
}

func TestParseDurationRejectsUnsupportedValues(t *testing.T) {
	for _, value := range []string{"0days", "days", "1.5days", "1hour"} {
		if _, err := ParseDuration(value); err == nil {
			t.Errorf("ParseDuration(%q) returned nil error", value)
		}
	}
}

func TestSaveDatabaseRewritesDatabaseConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`server:
  address: 127.0.0.1
  port: 9090
database:
  engine: sqlite
  sqlite:
    path: data/old.db
smtp:
  host: smtp.example.com
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	err := SaveDatabase(path, DatabaseConfig{
		Engine: "mysql",
		SQLite: SQLiteConfig{
			Path: "data/new.db",
		},
		MySQL: MySQLConfig{
			Host:     "db.example.com",
			Port:     3307,
			Username: "tower",
			Password: "secret",
			Database: "cephtower",
			Params:   "charset=utf8mb4",
			TLS:      "true",
		},
	})
	if err != nil {
		t.Fatalf("SaveDatabase() returned error: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Address != "127.0.0.1" || cfg.Server.Port != 9090 || cfg.SMTP.Host != "smtp.example.com" {
		t.Fatalf("non-database fields were not preserved: %#v", cfg)
	}
	if cfg.Database.Engine != "mysql" {
		t.Fatalf("Database.Engine = %q, want mysql", cfg.Database.Engine)
	}
	if cfg.Database.SQLite.Path != "data/new.db" {
		t.Fatalf("Database.SQLite.Path = %q, want data/new.db", cfg.Database.SQLite.Path)
	}
	if cfg.Database.MySQL.Host != "db.example.com" || cfg.Database.MySQL.Port != 3307 {
		t.Fatalf("unexpected MySQL address: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Username != "tower" || cfg.Database.MySQL.Password != "secret" {
		t.Fatalf("unexpected MySQL credentials: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.TLS != "true" {
		t.Fatalf("Database.MySQL.TLS = %q, want true", cfg.Database.MySQL.TLS)
	}
}

func TestSaveDatabaseOnlyReplacesDatabaseFieldValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`# service comment
server:
  address: "127.0.0.1" # keep this line unchanged

database:
  # selected database engine
  engine: "sqlite" # keep this comment
  sqlite:
    path: "data/old.db"
  mysql:
    host: "old.example.com"
    port: 3306
    username: "old-user"
    password: "old-password"
    database: "old-database"
    params: "charset=utf8mb4"

smtp:
  host: "smtp.example.com"
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	err := SaveDatabase(path, DatabaseConfig{
		Engine: "mysql",
		SQLite: SQLiteConfig{Path: "data/new.db"},
		MySQL: MySQLConfig{
			Host:     "db.example.com",
			Port:     3307,
			Username: "tower",
			Password: "secret#value",
			Database: "cephtower",
			Params:   "charset=utf8mb4&parseTime=True",
			TLS:      "skip-verify",
		},
	})
	if err != nil {
		t.Fatalf("SaveDatabase() returned error: %v", err)
	}

	want := `# service comment
server:
  address: "127.0.0.1" # keep this line unchanged

database:
  # selected database engine
  engine: "mysql" # keep this comment
  sqlite:
    path: "data/new.db"
  mysql:
    host: "db.example.com"
    port: 3307
    username: "tower"
    password: "secret#value"
    database: "cephtower"
    params: "charset=utf8mb4&parseTime=True"
    tls: skip-verify

smtp:
  host: "smtp.example.com"
`
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	if string(got) != want {
		t.Fatalf("updated config changed unrelated content:\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}

func TestSaveDatabaseRequiresExistingConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app", "config", "config.yaml")
	database := DatabaseConfig{
		Engine: "sqlite",
		SQLite: SQLiteConfig{Path: "data/initialized.db"},
		MySQL: MySQLConfig{
			Host:     "db.example.com",
			Port:     3307,
			Username: "tower",
			Password: "secret",
			Database: "cephtower",
			Params:   "charset=utf8mb4",
		},
	}

	if err := SaveDatabase(path, database); err == nil {
		t.Fatal("SaveDatabase() returned nil error for a missing config file")
	}
}

func TestLoadDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`{}`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Dir != DefaultServerDir || cfg.Server.Address != "0.0.0.0" || cfg.Server.Port != 36900 {
		t.Fatalf("Server = %#v, want default 0.0.0.0:36900", cfg.Server)
	}
	if cfg.Database.Engine != "sqlite" {
		t.Fatalf("Database.Engine = %q, want default sqlite", cfg.Database.Engine)
	}
	if cfg.Runtime.Dir != "data/runtime" {
		t.Fatalf("Runtime.Dir = %q, want default data/runtime", cfg.Runtime.Dir)
	}
	if cfg.Database.SQLite.Path != "data/db/cephtower.db" {
		t.Fatalf("Database.SQLite.Path = %q, want default data/db/cephtower.db", cfg.Database.SQLite.Path)
	}
	if cfg.Database.MySQL.Host != "127.0.0.1" || cfg.Database.MySQL.Port != 3306 {
		t.Fatalf("unexpected default MySQL address: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Username != "root" || cfg.Database.MySQL.Database != "cephtower" {
		t.Fatalf("unexpected default MySQL identity: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Params != "charset=utf8mb4&parseTime=True&loc=Local" {
		t.Fatalf("Database.MySQL.Params = %q, want default params", cfg.Database.MySQL.Params)
	}
	if cfg.Logging.Level != "info" || cfg.Logging.Format != "txt" || cfg.Logging.Dir != "log" || cfg.Logging.File != "cephtower.log" || cfg.Logging.Output != "both" || cfg.Logging.Rotation != "7days" || cfg.Logging.Retention != "70days" {
		t.Fatalf("Logging defaults = %#v, want info/txt", cfg.Logging)
	}
	if cfg.SMTP.Port != 587 {
		t.Fatalf("SMTP.Port = %d, want 587", cfg.SMTP.Port)
	}
}

func TestReferenceConfigMatchesCodeDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "..", "config", "config.yaml"))
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server != (ServerConfig{Address: defaultServerAddress, Port: defaultServerPort, Dir: DefaultServerDir}) {
		t.Fatalf("Server = %#v, want code defaults", cfg.Server)
	}
	if cfg.Logging != (LoggingConfig{Dir: defaultLogDir, File: defaultLogFile, Output: defaultLogOutput, Level: defaultLogLevel, Format: defaultLogFormat, Rotation: defaultLogRotation, Retention: defaultLogRetention}) {
		t.Fatalf("Logging = %#v, want code defaults", cfg.Logging)
	}
	if cfg.Runtime != (RuntimeConfig{Dir: defaultRuntimeDir}) {
		t.Fatalf("Runtime = %#v, want code defaults", cfg.Runtime)
	}
	if cfg.Database != (DatabaseConfig{
		Engine: defaultDatabaseEngine,
		SQLite: SQLiteConfig{Path: defaultSQLitePath},
		MySQL:  MySQLConfig{Host: defaultMySQLHost, Port: defaultMySQLPort, Username: defaultMySQLUsername, Database: defaultMySQLDatabase, Params: defaultMySQLParams, TLS: defaultMySQLTLS},
	}) {
		t.Fatalf("Database = %#v, want code defaults", cfg.Database)
	}
	if cfg.SMTP != (SMTPConfig{Port: defaultSMTPPort}) {
		t.Fatalf("SMTP = %#v, want code defaults", cfg.SMTP)
	}
}

func TestNormalizeDatabaseConfigRejectsUnsupportedTLSMode(t *testing.T) {
	_, err := NormalizeDatabaseConfig(DatabaseConfig{MySQL: MySQLConfig{TLS: "required"}})
	if err == nil {
		t.Fatal("NormalizeDatabaseConfig() returned nil error for unsupported TLS mode")
	}
}

func TestResolveRuntimeDir(t *testing.T) {
	if got := ResolveRuntimeDir(Config{
		Server:  ServerConfig{Dir: "./app"},
		Runtime: RuntimeConfig{Dir: "data/runtime"},
	}); got != filepath.Join("app", "data", "runtime") {
		t.Fatalf("ResolveRuntimeDir() = %q, want app/data/runtime", got)
	}

	if got := ResolveRuntimeDir(Config{
		Server:  ServerConfig{Dir: "./app"},
		Runtime: RuntimeConfig{Dir: "/var/lib/cephtower/runtime"},
	}); got != "/var/lib/cephtower/runtime" {
		t.Fatalf("ResolveRuntimeDir() = %q, want absolute runtime path", got)
	}
}

func TestLoadRejectsUnsupportedDatabaseEngine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`database:
  engine: postgres
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() returned nil error, want unsupported engine error")
	}
}

func TestLoadRejectsUnsupportedLoggingLevel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`log:
  level: verbose
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() returned nil error, want unsupported logging level error")
	}
}

func TestLoadRejectsUnsupportedLoggingFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`log:
  format: xml
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() returned nil error, want unsupported logging format error")
	}
}
