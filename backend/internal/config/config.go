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
}

type CephDashboardConfig struct {
	BaseURL     string
	Username    string
	Password    string
	InsecureTLS bool
}

type fileConfig struct {
	HTTPAddr string `yaml:"http_addr"`
	Ceph     struct {
		BaseURL     string `yaml:"base_url"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		InsecureTLS bool   `yaml:"insecure_tls"`
	} `yaml:"ceph_dashboard"`
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

	return Config{
		HTTPAddr: httpAddr,
		Ceph: CephDashboardConfig{
			BaseURL:     strings.TrimRight(strings.TrimSpace(raw.Ceph.BaseURL), "/"),
			Username:    raw.Ceph.Username,
			Password:    raw.Ceph.Password,
			InsecureTLS: raw.Ceph.InsecureTLS,
		},
	}, nil
}
