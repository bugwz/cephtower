package main

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestRootHelpOnlyDocumentsCommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"-h"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	usage := stdout.String()
	for _, want := range []string{"deploy", "watch", "Run deploy-to-machine <command> --help"} {
		if !strings.Contains(usage, want) {
			t.Fatalf("help output missing %q:\n%s", want, usage)
		}
	}
	for _, unwanted := range []string{"--config", "--replace"} {
		if strings.Contains(usage, unwanted) {
			t.Fatalf("root help should not document command option %q:\n%s", unwanted, usage)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("help wrote to stderr:\n%s", stderr.String())
	}
}

func TestCommandHelpDocumentsCommandOptions(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"deploy", "--help"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	deployUsage := stdout.String()
	for _, want := range []string{"Usage: deploy-to-machine deploy", "--config", "--replace"} {
		if !strings.Contains(deployUsage, want) {
			t.Fatalf("deploy help output missing %q:\n%s", want, deployUsage)
		}
	}
	stdout.Reset()
	stderr.Reset()
	if err := run([]string{"watch", "--help"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	watchUsage := stdout.String()
	for _, want := range []string{"Usage: deploy-to-machine watch", "--config"} {
		if !strings.Contains(watchUsage, want) {
			t.Fatalf("watch help output missing %q:\n%s", want, watchUsage)
		}
	}
	if strings.Contains(watchUsage, "--replace") {
		t.Fatalf("watch help should not document deploy-only option --replace:\n%s", watchUsage)
	}
	if stderr.Len() != 0 {
		t.Fatalf("help wrote to stderr:\n%s", stderr.String())
	}
}

func TestHelpDoesNotDocumentRemovedOptions(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"deploy", "--help"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	usage := stdout.String()
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

func TestParseOptionsRecognizesCommands(t *testing.T) {
	repoRoot := t.TempDir()
	toolDir := filepath.Join(repoRoot, "tools", "deploy-to-machine")
	var output bytes.Buffer
	opts, err := parseOptions([]string{"deploy", "--replace", "bin"}, repoRoot, toolDir, &output)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != commandDeploy || !opts.Replace["bin"] {
		t.Fatalf("unexpected deploy options: %#v", opts)
	}
	opts, err = parseOptions([]string{"watch"}, repoRoot, toolDir, &output)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != commandWatch {
		t.Fatalf("expected watch command, got %#v", opts.Command)
	}
	if opts.RemoteLogPath != "/opt/cephtower/log/cephtower.log" {
		t.Fatalf("unexpected log path: %s", opts.RemoteLogPath)
	}
	if _, err := parseOptions([]string{"logs"}, repoRoot, toolDir, &output); err == nil {
		t.Fatal("expected unknown command to fail")
	}
}

func TestParseOptionsRequiresCommand(t *testing.T) {
	repoRoot := t.TempDir()
	toolDir := filepath.Join(repoRoot, "tools", "deploy-to-machine")
	var output bytes.Buffer
	if _, err := parseOptions(nil, repoRoot, toolDir, &output); err == nil || !strings.Contains(err.Error(), "missing command") {
		t.Fatalf("expected missing command to fail, got %v", err)
	}
	if _, err := parseOptions([]string{"--replace", "bin"}, repoRoot, toolDir, &output); err == nil || !strings.Contains(err.Error(), "missing command") {
		t.Fatalf("expected flag without command to fail, got %v", err)
	}
	if _, err := parseOptions([]string{"watch", "--replace", "bin"}, repoRoot, toolDir, &output); err == nil {
		t.Fatal("expected watch to reject deploy-only --replace")
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

func TestSaveAndLoadDeployState(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".state", "last-deploy.json")
	state := deployState{
		Version:       1,
		Target:        targetConfig{Host: "203.0.113.10", Port: 22, User: "root", Password: "secret", KnownHosts: "/tmp/known_hosts"},
		RemoteRoot:    remoteRoot,
		RemoteLogPath: remoteLogDir + "/cephtower.log",
		PID:           "1234",
	}
	if err := saveDeployState(path, state); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["remote_log_path"] != "/opt/cephtower/log/cephtower.log" {
		t.Fatalf("unexpected state payload:\n%s", string(raw))
	}
	got, err := loadDeployState(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Target.Host != state.Target.Host || got.RemoteLogPath != state.RemoteLogPath {
		t.Fatalf("unexpected loaded state: %#v", got)
	}
}

func TestLearnHostKeyRefreshesChangedKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	host := "203.0.113.10:22"
	address := &net.TCPAddr{IP: net.ParseIP("203.0.113.10"), Port: 22}
	first := testSSHPublicKey(t)
	second := testSSHPublicKey(t)

	callback, err := knownhosts.New(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := learnHostKey(path, callback)(host, address, first); err != nil {
		t.Fatalf("learnHostKey() rejected new key: %v", err)
	}
	callback, err = knownhosts.New(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := learnHostKey(path, callback)(host, address, second); err != nil {
		t.Fatalf("learnHostKey() rejected changed key: %v", err)
	}
	callback, err = knownhosts.New(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := callback(host, address, second); err != nil {
		t.Fatalf("refreshed known_hosts does not accept new key: %v", err)
	}
	if err := callback(host, address, first); err == nil {
		t.Fatal("refreshed known_hosts still accepts old key")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	normalizedHost := knownhosts.Normalize(host)
	if got := strings.Count(string(raw), normalizedHost); got != 1 {
		t.Fatalf("expected one known_hosts entry for %s, got %d:\n%s", normalizedHost, got, string(raw))
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

func testSSHPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return publicKey
}

func TestStartScriptWritesAllOutputToCephtowerLog(t *testing.T) {
	script := startServiceCommand()
	if !strings.Contains(script, ">> '/opt/cephtower/log/cephtower.log' 2>&1") {
		t.Fatalf("start command does not append stdout/stderr to cephtower.log: %s", script)
	}
	if strings.Contains(script, "cephtower.stdout.log") {
		t.Fatalf("start command still writes split stdout log: %s", script)
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
	raw, err := configWithServerDir(path, "/opt/cephtower", remoteConfigValues{DatabaseEncryptionKey: key})
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

func TestConfigWithServerDirPreservesRemoteBootstrap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("server:\n    dir: ./app\n    bootstrap: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	key := "0123456789abcdefghijklmnopqrstuv"
	bootstrap := false
	raw, err := configWithServerDir(path, "/opt/cephtower", remoteConfigValues{
		DatabaseEncryptionKey: key,
		ServerBootstrap:       &bootstrap,
	})
	if err != nil {
		t.Fatal(err)
	}
	value := string(raw)
	if !strings.Contains(value, "bootstrap: false") {
		t.Fatalf("server.bootstrap was not preserved from remote config:\n%s", value)
	}
	if strings.Contains(value, "bootstrap: true") {
		t.Fatalf("server.bootstrap kept the local template value:\n%s", value)
	}
}

func TestConfigValuesForReplacingDataEnablesBootstrap(t *testing.T) {
	for _, tc := range []struct {
		name    string
		replace replaceSet
		source  string
	}{
		{name: "data", replace: replaceSet{"data": true}, source: "--replace data"},
		{name: "all", replace: replaceSet{"all": true}, source: "--replace all"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bootstrap := false
			values := configValuesForReplace(remoteConfigValues{
				DatabaseEncryptionKey: "0123456789abcdefghijklmnopqrstuv",
				ServerBootstrap:       &bootstrap,
				ServerBootstrapSource: "remote",
			}, tc.replace)
			if values.ServerBootstrap == nil || !*values.ServerBootstrap {
				t.Fatalf("expected replacing data to enable server.bootstrap, got %#v", values.ServerBootstrap)
			}
			if values.ServerBootstrapSource != tc.source {
				t.Fatalf("unexpected bootstrap source: %q", values.ServerBootstrapSource)
			}
		})
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

func TestRemoteConfigValuesFromConfigReadsBootstrap(t *testing.T) {
	key := "0123456789abcdefghijklmnopqrstuv"
	values, err := remoteConfigValuesFromConfig([]byte("server:\n  bootstrap: false\ndatabase:\n  encryption_key: " + key + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if values.DatabaseEncryptionKey != key {
		t.Fatalf("unexpected key: %q", values.DatabaseEncryptionKey)
	}
	if values.ServerBootstrap == nil || *values.ServerBootstrap {
		t.Fatalf("expected server.bootstrap=false, got %#v", values.ServerBootstrap)
	}
	if _, err := remoteConfigValuesFromConfig([]byte("server:\n  bootstrap:\n    nested: true\n")); err == nil {
		t.Fatal("expected non-scalar bootstrap to fail")
	}
	if _, err := remoteConfigValuesFromConfig([]byte("server:\n  bootstrap: maybe\n")); err == nil {
		t.Fatal("expected invalid bootstrap to fail")
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
