package mysql

import (
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

	return mysqlConfig.FormatDSN(), nil
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
