package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`http_addr: ":9090"
ceph_dashboard:
  base_url: https://ceph.example.com/
  username: admin
  password: change-me
  insecure_tls: true
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":9090")
	}
	if cfg.Ceph.BaseURL != "https://ceph.example.com" {
		t.Fatalf("Ceph.BaseURL = %q, want trimmed URL", cfg.Ceph.BaseURL)
	}
	if cfg.Ceph.Username != "admin" || cfg.Ceph.Password != "change-me" {
		t.Fatalf("unexpected Ceph credentials: %#v", cfg.Ceph)
	}
	if !cfg.Ceph.InsecureTLS {
		t.Fatal("Ceph.InsecureTLS = false, want true")
	}
}

func TestLoadDefaultsHTTPAddr(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`ceph_dashboard:
  base_url: https://ceph.example.com
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTPAddr != ":36900" {
		t.Fatalf("HTTPAddr = %q, want default :36900", cfg.HTTPAddr)
	}
}
