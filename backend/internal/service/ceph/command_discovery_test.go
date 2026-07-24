package ceph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cephtower/backend/internal/store"
)

func TestCephClusterRuntimeFilesSyncAndDelete(t *testing.T) {
	workDir := t.TempDir()
	cluster := &store.CephCluster{
		ID:          7,
		MonitorHost: "10.0.0.11:6789,10.0.0.12:6789",
		Keyring:     "command-secret",
	}

	if err := SyncCephClusterRuntimeFiles(workDir, cluster); err != nil {
		t.Fatalf("sync runtime files: %v", err)
	}

	runtimeDir := filepath.Join(workDir, "ceph", "7")
	confPath := filepath.Join(runtimeDir, "ceph.conf")
	keyringPath := filepath.Join(runtimeDir, "ceph.client.admin.keyring")
	conf, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("read ceph config: %v", err)
	}
	if string(conf) != "[global]\nmon host = 10.0.0.11:6789,10.0.0.12:6789\n" {
		t.Fatalf("ceph config = %q", conf)
	}

	cluster.MonitorHost = "10.0.0.21:6789"
	cluster.Keyring = "updated-secret"
	if err := SyncCephClusterRuntimeFiles(workDir, cluster); err != nil {
		t.Fatalf("resync runtime files: %v", err)
	}
	updatedConf, _ := os.ReadFile(confPath)
	updatedKeyring, _ := os.ReadFile(keyringPath)
	if !strings.Contains(string(updatedConf), cluster.MonitorHost) {
		t.Fatalf("updated ceph config = %q", updatedConf)
	}
	if !strings.Contains(string(updatedKeyring), "key = updated-secret") {
		t.Fatalf("updated ceph keyring = %q", updatedKeyring)
	}

	if err := DeleteCephClusterRuntimeFiles(workDir, cluster.ID); err != nil {
		t.Fatalf("delete runtime files: %v", err)
	}
	if _, err := os.Stat(runtimeDir); !os.IsNotExist(err) {
		t.Fatalf("runtime directory still exists, stat error = %v", err)
	}
}
