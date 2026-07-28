package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpOnlyDocumentsConfigAndReplace(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"-h"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	usage := stdout.String()
	for _, want := range []string{"--config", "--replace"} {
		if !strings.Contains(usage, want) {
			t.Fatalf("help output missing %q:\n%s", want, usage)
		}
	}
	for _, unwanted := range []string{
		"--host",
		"--port",
		"--user",
		"--password",
		"--known-hosts",
		"--app-config",
		"--release-dir",
		"--non-interactive",
	} {
		if strings.Contains(usage, unwanted) {
			t.Fatalf("help output still documents %q:\n%s", unwanted, usage)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("help wrote to stderr:\n%s", stderr.String())
	}
}

func TestParseReplace(t *testing.T) {
	got, err := parseReplace("bin, conf,config,data,log")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"bin", "conf", "data", "log"} {
		if !got[key] {
			t.Fatalf("expected %s to be selected", key)
		}
	}
	if _, err := parseReplace("cache"); err == nil {
		t.Fatal("expected invalid replace value to fail")
	}
}

func TestBackupRemoteRootCommand(t *testing.T) {
	command := backupRemoteRootCommand()
	for _, want := range []string{
		"/opt/cephtower.backup",
		"$(date +%Y%m%d%H%M%S)",
		"cp -a '/opt/cephtower'/. \"$backup_dir\"",
	} {
		if !strings.Contains(command, want) {
			t.Fatalf("backup command missing %q: %s", want, command)
		}
	}
}

func TestNormalizeRemoteOSArch(t *testing.T) {
	info, err := normalizeRemoteOSArch("Linux", "x86_64")
	if err != nil {
		t.Fatal(err)
	}
	if info.GOOS != "linux" || info.GOARCH != "amd64" {
		t.Fatalf("unexpected normalized info: %#v", info)
	}
	info, err = normalizeRemoteOSArch("Linux", "aarch64")
	if err != nil {
		t.Fatal(err)
	}
	if info.GOARCH != "arm64" {
		t.Fatalf("unexpected arch: %#v", info)
	}
}

func TestConfigWithServerDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("server:\n    address: 0.0.0.0\n    dir: ./app\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	key := "0123456789abcdefghijklmnopqrstuv"
	raw, err := configWithServerDir(path, "/opt/cephtower", key)
	if err != nil {
		t.Fatal(err)
	}
	value := string(raw)
	if !strings.Contains(value, "dir: /opt/cephtower") {
		t.Fatalf("server.dir was not rewritten:\n%s", value)
	}
	if !strings.Contains(value, "encryption_key: "+key) {
		t.Fatalf("database.encryption_key was not rewritten:\n%s", value)
	}
}

func TestEncryptionKeyFromConfig(t *testing.T) {
	key := "0123456789abcdefghijklmnopqrstuv"
	got, ok, err := encryptionKeyFromConfig([]byte("database:\n  encryption_key: " + key + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !ok || got != key {
		t.Fatalf("expected key %q, got %q ok=%t", key, got, ok)
	}
	if _, ok, err := encryptionKeyFromConfig([]byte("database:\n  encryption_key: \"\"\n")); err != nil || ok {
		t.Fatalf("expected empty key to be absent, ok=%t err=%v", ok, err)
	}
	if _, _, err := encryptionKeyFromConfig([]byte("database:\n  encryption_key: short\n")); err == nil {
		t.Fatal("expected invalid key to fail")
	}
}

func TestGenerateDatabaseEncryptionKey(t *testing.T) {
	key, err := generateDatabaseEncryptionKey()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDatabaseEncryptionKey(key); err != nil {
		t.Fatal(err)
	}
}

func TestSelectReleaseArtifact(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "cephtower-0.0.1-old-linux-amd64")
	newPath := filepath.Join(dir, "cephtower-0.0.1-new-linux-amd64")
	if err := os.WriteFile(oldPath, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := selectReleaseArtifact(dir, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if got != newPath {
		t.Fatalf("expected newest artifact %s, got %s", newPath, got)
	}
}

func TestDiscoverLabMachinesReadsStateJSON(t *testing.T) {
	repoRoot := t.TempDir()
	stateDir := filepath.Join(repoRoot, "tools", "aliyun-ceph-lab", ".state")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(stateDir, "ceph-dev.json")
	state := `{
  "version": 1,
  "cluster_name": "ceph-dev",
  "region_id": "ap-southeast-1",
  "nodes": [
    {
      "name": "ceph-node-1",
      "status": "Running",
      "public_ip": "203.0.113.10",
      "ssh": {
        "port": 22,
        "user": "root",
        "password": "secret"
      }
    }
  ]
}`
	if err := os.WriteFile(statePath, []byte(state), 0o600); err != nil {
		t.Fatal(err)
	}
	machines, err := discoverLabMachines(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	nodes := flattenDiscoveredNodes(machines)
	if len(nodes) != 1 {
		t.Fatalf("expected one node, got %d", len(nodes))
	}
	if nodes[0].ClusterName != "ceph-dev" || nodeHost(nodes[0]) != "203.0.113.10" {
		t.Fatalf("unexpected discovered node: %#v", nodes[0])
	}
}
