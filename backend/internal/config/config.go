package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/security"
	"gopkg.in/yaml.v3"
)

const (
	DefaultPath           = "/opt/cephtower/config/config.yaml"
	DefaultServerDir      = "/opt/cephtower"
	defaultServerAddress  = "0.0.0.0"
	defaultServerPort     = 36900
	defaultLogDir         = "log"
	defaultLogFile        = "cephtower.log"
	defaultLogOutput      = "both"
	defaultLogLevel       = "info"
	defaultLogFormat      = "txt"
	defaultLogRotation    = "7days"
	defaultLogRetention   = "70days"
	defaultRuntimeDir     = "data/runtime"
	defaultDatabaseEngine = "sqlite"
	defaultSQLitePath     = "data/db/cephtower.db"
	defaultMySQLHost      = "127.0.0.1"
	defaultMySQLPort      = 3306
	defaultMySQLUsername  = "root"
	defaultMySQLDatabase  = "cephtower"
	defaultMySQLParams    = "charset=utf8mb4&parseTime=True&loc=Local"
	defaultMySQLTLS       = "false"
	defaultSMTPPort       = 587
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
	EncryptionKey string       `yaml:"encryption_key"`
	Engine        string       `yaml:"engine"`
	SQLite        SQLiteConfig `yaml:"sqlite"`
	MySQL         MySQLConfig  `yaml:"mysql"`
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
	Path string `yaml:"path"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Params   string `yaml:"params"`
	TLS      string `yaml:"tls"`
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
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
		Path      string `yaml:"path,omitempty"` // backward compatibility with older configs
		Output    string `yaml:"output"`
		Level     string `yaml:"level"`
		Format    string `yaml:"format"`
		Rotation  string `yaml:"rotation"`
		Retention string `yaml:"retention"`
	} `yaml:"log"`
	Database struct {
		EncryptionKey string `yaml:"encryption_key"`
		Engine        string `yaml:"engine"`
		SQLite        struct {
			Path string `yaml:"path"`
		} `yaml:"sqlite"`
		MySQL struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Database string `yaml:"database"`
			Params   string `yaml:"params"`
			TLS      string `yaml:"tls"`
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
	path = strings.TrimSpace(path)
	if path == "" {
		return Config{}, fmt.Errorf("config file path is required")
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
		server.Dir = DefaultServerDir
	}
	if server.Address == "" {
		server.Address = defaultServerAddress
	}
	if server.Port == 0 {
		server.Port = defaultServerPort
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
			Port:     defaultInt(raw.SMTP.Port, defaultSMTPPort),
			Username: strings.TrimSpace(raw.SMTP.Username),
			Password: raw.SMTP.Password,
			From:     strings.TrimSpace(raw.SMTP.From),
		},
	}, nil
}

func normalizeRuntimeConfig(raw fileConfig) RuntimeConfig {
	dir := strings.TrimSpace(raw.Runtime.Dir)
	if dir == "" {
		dir = defaultRuntimeDir
	}
	return RuntimeConfig{Dir: dir}
}

func ResolveRuntimeDir(cfg Config) string {
	dir := strings.TrimSpace(cfg.Runtime.Dir)
	if dir == "" {
		dir = defaultRuntimeDir
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
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("config file path is required")
	}

	normalized, err := NormalizeDatabaseConfig(database)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}

	output, err := updateDatabaseValues(data, normalized)
	if err != nil {
		return fmt.Errorf("parse config file %q: %w", path, err)
	}
	if err := os.WriteFile(path, output, 0o600); err != nil {
		return fmt.Errorf("write config file %q: %w", path, err)
	}
	return nil
}

type yamlEdit struct {
	start int
	end   int
	value []byte
}

func updateDatabaseValues(data []byte, database DatabaseConfig) ([]byte, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	if len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("configuration root must be a mapping")
	}

	root := document.Content[0]
	databaseNode := mappingValue(root, "database")
	if databaseNode == nil {
		if len(root.Content) == 0 {
			return marshalDatabaseConfig(database), nil
		}
		if root.Style&yaml.FlowStyle != 0 {
			return nil, fmt.Errorf("cannot append database fields to a flow-style configuration")
		}
		return appendDatabaseSection(data, database), nil
	}
	if databaseNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("database configuration must be a mapping")
	}

	values := []struct {
		path  []string
		value string
		tag   string
	}{
		{[]string{"encryption_key"}, database.EncryptionKey, "!!str"},
		{[]string{"engine"}, database.Engine, "!!str"},
		{[]string{"sqlite", "path"}, database.SQLite.Path, "!!str"},
		{[]string{"mysql", "host"}, database.MySQL.Host, "!!str"},
		{[]string{"mysql", "port"}, strconv.Itoa(database.MySQL.Port), "!!int"},
		{[]string{"mysql", "username"}, database.MySQL.Username, "!!str"},
		{[]string{"mysql", "password"}, database.MySQL.Password, "!!str"},
		{[]string{"mysql", "database"}, database.MySQL.Database, "!!str"},
		{[]string{"mysql", "params"}, database.MySQL.Params, "!!str"},
		{[]string{"mysql", "tls"}, database.MySQL.TLS, "!!str"},
	}

	edits := make([]yamlEdit, 0, len(values))
	missing := make(map[string]struct{})
	for _, item := range values {
		node := nestedMappingValue(databaseNode, item.path...)
		if node == nil {
			missing[strings.Join(item.path, ".")] = struct{}{}
			continue
		}
		if node.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("database field %q must be a scalar", strings.Join(item.path, "."))
		}
		if node.Style == yaml.LiteralStyle || node.Style == yaml.FoldedStyle {
			return nil, fmt.Errorf("database field %q must use a single-line value", strings.Join(item.path, "."))
		}
		start, end, err := scalarRange(data, node)
		if err != nil {
			return nil, fmt.Errorf("locate database field %q: %w", strings.Join(item.path, "."), err)
		}
		edits = append(edits, yamlEdit{start: start, end: end, value: encodeScalar(item.value, item.tag, node.Style)})
	}

	databaseKey := mappingKey(root, "database")
	databaseIndent := mappingIndent(databaseNode, databaseKey.Column-1+2)
	indentStep := databaseIndent - (databaseKey.Column - 1)
	if indentStep < 1 {
		indentStep = 2
	}
	insertions := make(map[int][]byte)
	databaseLines := make([]string, 0)
	if _, ok := missing["engine"]; ok {
		databaseLines = append(databaseLines, yamlLine(databaseIndent, "engine", database.Engine, "!!str"))
	}
	if _, ok := missing["encryption_key"]; ok {
		databaseLines = append([]string{yamlLine(databaseIndent, "encryption_key", database.EncryptionKey, "!!str")}, databaseLines...)
	}

	sqliteNode := mappingValue(databaseNode, "sqlite")
	if sqliteNode == nil {
		databaseLines = append(databaseLines,
			strings.Repeat(" ", databaseIndent)+"sqlite:",
			yamlLine(databaseIndent+indentStep, "path", database.SQLite.Path, "!!str"),
		)
	} else {
		if sqliteNode.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("database field %q must be a mapping", "sqlite")
		}
		if _, ok := missing["sqlite.path"]; ok {
			if sqliteNode.Style&yaml.FlowStyle != 0 {
				return nil, fmt.Errorf("cannot append missing fields to a flow-style sqlite configuration")
			}
			indent := mappingIndent(sqliteNode, databaseIndent+indentStep)
			addInsertion(insertions, mappingInsertionOffset(data, sqliteNode), yamlLine(indent, "path", database.SQLite.Path, "!!str")+"\n")
		}
	}

	mysqlNode := mappingValue(databaseNode, "mysql")
	if mysqlNode == nil {
		databaseLines = append(databaseLines,
			strings.Repeat(" ", databaseIndent)+"mysql:",
			yamlLine(databaseIndent+indentStep, "host", database.MySQL.Host, "!!str"),
			yamlLine(databaseIndent+indentStep, "port", strconv.Itoa(database.MySQL.Port), "!!int"),
			yamlLine(databaseIndent+indentStep, "username", database.MySQL.Username, "!!str"),
			yamlLine(databaseIndent+indentStep, "password", database.MySQL.Password, "!!str"),
			yamlLine(databaseIndent+indentStep, "database", database.MySQL.Database, "!!str"),
			yamlLine(databaseIndent+indentStep, "params", database.MySQL.Params, "!!str"),
			yamlLine(databaseIndent+indentStep, "tls", database.MySQL.TLS, "!!str"),
		)
	} else {
		if mysqlNode.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("database field %q must be a mapping", "mysql")
		}
		indent := mappingIndent(mysqlNode, databaseIndent+indentStep)
		mysqlLines := make([]string, 0)
		mysqlFields := []struct {
			path  string
			key   string
			value string
			tag   string
		}{
			{"mysql.host", "host", database.MySQL.Host, "!!str"},
			{"mysql.port", "port", strconv.Itoa(database.MySQL.Port), "!!int"},
			{"mysql.username", "username", database.MySQL.Username, "!!str"},
			{"mysql.password", "password", database.MySQL.Password, "!!str"},
			{"mysql.database", "database", database.MySQL.Database, "!!str"},
			{"mysql.params", "params", database.MySQL.Params, "!!str"},
			{"mysql.tls", "tls", database.MySQL.TLS, "!!str"},
		}
		for _, field := range mysqlFields {
			if _, ok := missing[field.path]; ok {
				mysqlLines = append(mysqlLines, yamlLine(indent, field.key, field.value, field.tag))
			}
		}
		if len(mysqlLines) > 0 {
			if mysqlNode.Style&yaml.FlowStyle != 0 {
				return nil, fmt.Errorf("cannot append missing fields to a flow-style mysql configuration")
			}
			addInsertion(insertions, mappingInsertionOffset(data, mysqlNode), strings.Join(mysqlLines, "\n")+"\n")
		}
	}

	if len(databaseLines) > 0 {
		if databaseNode.Style&yaml.FlowStyle != 0 {
			return nil, fmt.Errorf("cannot append missing fields to a flow-style database configuration")
		}
		addInsertion(insertions, mappingInsertionOffset(data, databaseNode), strings.Join(databaseLines, "\n")+"\n")
	}
	for offset, value := range insertions {
		if offset > 0 && data[offset-1] != '\n' {
			value = append([]byte("\n"), value...)
		}
		edits = append(edits, yamlEdit{start: offset, end: offset, value: value})
	}

	sort.Slice(edits, func(left, right int) bool { return edits[left].start < edits[right].start })
	for index := len(edits) - 1; index >= 0; index-- {
		edit := edits[index]
		data = bytes.Join([][]byte{data[:edit.start], edit.value, data[edit.end:]}, nil)
	}
	return data, nil
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func mappingKey(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index]
		}
	}
	return nil
}

func nestedMappingValue(mapping *yaml.Node, path ...string) *yaml.Node {
	current := mapping
	for _, key := range path {
		current = mappingValue(current, key)
		if current == nil {
			return nil
		}
	}
	return current
}

func mappingIndent(mapping *yaml.Node, fallback int) int {
	if mapping != nil && len(mapping.Content) > 0 {
		return mapping.Content[0].Column - 1
	}
	return fallback
}

func mappingInsertionOffset(data []byte, mapping *yaml.Node) int {
	lastLine := mapping.Line
	var visit func(*yaml.Node)
	visit = func(node *yaml.Node) {
		if node.Line > lastLine {
			lastLine = node.Line
		}
		for _, child := range node.Content {
			visit(child)
		}
	}
	visit(mapping)

	offset := 0
	for line := 1; line <= lastLine; line++ {
		next := bytes.IndexByte(data[offset:], '\n')
		if next < 0 {
			return len(data)
		}
		offset += next + 1
	}
	return offset
}

func addInsertion(insertions map[int][]byte, offset int, value string) {
	insertions[offset] = append(insertions[offset], value...)
}

func yamlLine(indent int, key string, value string, tag string) string {
	return strings.Repeat(" ", indent) + key + ": " + string(encodeScalar(value, tag, 0))
}

func scalarRange(data []byte, node *yaml.Node) (int, int, error) {
	lineStart := 0
	for line := 1; line < node.Line; line++ {
		next := bytes.IndexByte(data[lineStart:], '\n')
		if next < 0 {
			return 0, 0, fmt.Errorf("line %d is outside the file", node.Line)
		}
		lineStart += next + 1
	}
	lineEnd := bytes.IndexByte(data[lineStart:], '\n')
	if lineEnd < 0 {
		lineEnd = len(data) - lineStart
	}
	line := data[lineStart : lineStart+lineEnd]
	startInLine := byteOffsetForColumn(line, node.Column)
	if startInLine < 0 || startInLine >= len(line) {
		return 0, 0, fmt.Errorf("column %d is outside line %d", node.Column, node.Line)
	}

	endInLine := scalarEnd(line, startInLine)
	if endInLine <= startInLine {
		return 0, 0, fmt.Errorf("could not determine scalar boundary")
	}
	return lineStart + startInLine, lineStart + endInLine, nil
}

func byteOffsetForColumn(line []byte, column int) int {
	if column < 1 {
		return -1
	}
	offset := 0
	for current := 1; current < column && offset < len(line); current++ {
		offset++
		for offset < len(line) && line[offset]&0xc0 == 0x80 {
			offset++
		}
	}
	return offset
}

func scalarEnd(line []byte, start int) int {
	switch line[start] {
	case '"':
		for index := start + 1; index < len(line); index++ {
			if line[index] == '\\' {
				index++
				continue
			}
			if line[index] == '"' {
				return index + 1
			}
		}
	case '\'':
		for index := start + 1; index < len(line); index++ {
			if line[index] != '\'' {
				continue
			}
			if index+1 < len(line) && line[index+1] == '\'' {
				index++
				continue
			}
			return index + 1
		}
	default:
		end := len(line)
		for index := start; index < len(line); index++ {
			if line[index] == '#' && index > start && (line[index-1] == ' ' || line[index-1] == '\t') {
				end = index
				break
			}
		}
		for end > start && (line[end-1] == ' ' || line[end-1] == '\t' || line[end-1] == '\r') {
			end--
		}
		return end
	}
	return -1
}

func encodeScalar(value string, tag string, style yaml.Style) []byte {
	if style == yaml.DoubleQuotedStyle {
		return []byte(strconv.Quote(value))
	}
	if style == yaml.SingleQuotedStyle {
		return []byte("'" + strings.ReplaceAll(value, "'", "''") + "'")
	}
	if tag == "!!int" {
		return []byte(value)
	}
	node := yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
	encoded, err := yaml.Marshal(&node)
	if err != nil {
		return []byte(strconv.Quote(value))
	}
	return bytes.TrimSuffix(encoded, []byte("\n"))
}

func appendDatabaseSection(data []byte, database DatabaseConfig) []byte {
	section := marshalDatabaseConfig(database)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	if len(data) > 0 {
		data = append(data, '\n')
	}
	return append(data, section...)
}

func marshalDatabaseConfig(database DatabaseConfig) []byte {
	type databaseFile struct {
		Database DatabaseConfig `yaml:"database"`
	}
	output, _ := yaml.Marshal(databaseFile{Database: database})
	return output
}

func normalizeLoggingConfig(raw fileConfig) (LoggingConfig, error) {
	level := strings.ToLower(strings.TrimSpace(raw.Logging.Level))
	if level == "" {
		level = defaultLogLevel
	}
	switch level {
	case "debug", "info", "warn", "error":
	default:
		return LoggingConfig{}, fmt.Errorf("unsupported logging level %q", raw.Logging.Level)
	}

	format := strings.ToLower(strings.TrimSpace(raw.Logging.Format))
	if format == "" {
		format = defaultLogFormat
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
		dir = defaultLogDir
	}
	if file == "" {
		file = defaultLogFile
	}

	output := strings.ToLower(strings.TrimSpace(raw.Logging.Output))
	if output == "" {
		output = defaultLogOutput
	}
	switch output {
	case "stdout", "file", "both":
	default:
		return LoggingConfig{}, fmt.Errorf("unsupported logging output %q", raw.Logging.Output)
	}

	rotation := strings.ToLower(strings.TrimSpace(raw.Logging.Rotation))
	if rotation == "" {
		rotation = defaultLogRotation
	}
	if _, err := ParseDuration(rotation); err != nil {
		return LoggingConfig{}, fmt.Errorf("invalid logging rotation %q: %w", raw.Logging.Rotation, err)
	}

	retention := strings.ToLower(strings.TrimSpace(raw.Logging.Retention))
	if retention == "" {
		retention = defaultLogRetention
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
		EncryptionKey: raw.Database.EncryptionKey,
		Engine:        raw.Database.Engine,
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
			TLS:      raw.Database.MySQL.TLS,
		},
	})
}

func NormalizeDatabaseConfig(cfg DatabaseConfig) (DatabaseConfig, error) {
	if err := security.ValidateDatabaseEncryptionKey(cfg.EncryptionKey); err != nil {
		return DatabaseConfig{}, err
	}
	engine := strings.ToLower(strings.TrimSpace(cfg.Engine))
	if engine == "" {
		engine = defaultDatabaseEngine
	}
	if engine != "sqlite" && engine != "mysql" {
		return DatabaseConfig{}, fmt.Errorf("unsupported database engine %q", cfg.Engine)
	}

	sqlitePath := strings.TrimSpace(cfg.SQLite.Path)
	if sqlitePath == "" {
		sqlitePath = defaultSQLitePath
	}

	mysqlHost := strings.TrimSpace(cfg.MySQL.Host)
	if mysqlHost == "" {
		mysqlHost = defaultMySQLHost
	}

	mysqlPort := cfg.MySQL.Port
	if mysqlPort == 0 {
		mysqlPort = defaultMySQLPort
	}

	mysqlUsername := strings.TrimSpace(cfg.MySQL.Username)
	if mysqlUsername == "" {
		mysqlUsername = defaultMySQLUsername
	}

	mysqlDatabase := strings.TrimSpace(cfg.MySQL.Database)
	if mysqlDatabase == "" {
		mysqlDatabase = defaultMySQLDatabase
	}

	mysqlParams := strings.TrimSpace(cfg.MySQL.Params)
	if mysqlParams == "" {
		mysqlParams = defaultMySQLParams
	}

	mysqlTLS := strings.ToLower(strings.TrimSpace(cfg.MySQL.TLS))
	if mysqlTLS == "" {
		mysqlTLS = defaultMySQLTLS
	}
	switch mysqlTLS {
	case "false", "true", "skip-verify", "preferred":
	default:
		return DatabaseConfig{}, fmt.Errorf("unsupported mysql tls mode %q", cfg.MySQL.TLS)
	}

	return DatabaseConfig{
		EncryptionKey: cfg.EncryptionKey,
		Engine:        engine,
		SQLite: SQLiteConfig{
			Path: sqlitePath,
		},
		MySQL: MySQLConfig{
			Host:     mysqlHost,
			Port:     mysqlPort,
			Username: mysqlUsername,
			Password: cfg.MySQL.Password,
			Database: mysqlDatabase,
			Params:   mysqlParams,
			TLS:      mysqlTLS,
		},
	}, nil
}
