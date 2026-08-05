package workspace

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseOptionsLoadsServiceConfig(t *testing.T) {
	repo := t.TempDir()
	toolDir := filepath.Join(repo, "tools", "sync-workspace")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(toolDir, "config.local.yaml")
	content := "host: 192.0.2.10\nport: 2222\nuser: root\npassword: secret\nknown_hosts: .state/hosts\nclean_on_start: true\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := parseOptions([]string{"run"}, repo, toolDir, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Target.Host != "192.0.2.10" || opts.Target.Port != 2222 || opts.Target.Password != "secret" {
		t.Fatalf("unexpected target: %#v", opts.Target)
	}
	if !opts.CleanOnStart {
		t.Fatal("expected clean_on_start")
	}
	wantKnownHosts := filepath.Join(toolDir, ".state", "hosts")
	if opts.Target.KnownHosts != wantKnownHosts {
		t.Fatalf("known_hosts = %q, want %q", opts.Target.KnownHosts, wantKnownHosts)
	}
}

func TestLoadFileConfigRejectsNestedServiceConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("service:\n  clean_on_start: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadFileConfig(path); err == nil {
		t.Fatal("expected nested service config to be rejected")
	}
}

func TestConfigForRemoteWeb(t *testing.T) {
	raw := []byte(`server:
    address: "127.0.0.1"
    port: 1234
    dir: "./app"
    bootstrap: false
database:
    encryption_key: "01234567890123456789012345678901"
`)
	payload, changed, err := configForRemoteWeb(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected config to change")
	}
	var config map[string]any
	if err := yaml.Unmarshal(payload, &config); err != nil {
		t.Fatal(err)
	}
	server := config["server"].(map[string]any)
	if server["address"] != "0.0.0.0" || server["port"] != webPort || server["dir"] != remoteAppDir {
		t.Fatalf("unexpected server config: %#v", server)
	}
	if server["bootstrap"] != false {
		t.Fatal("bootstrap value was not preserved")
	}
	database := config["database"].(map[string]any)
	if database["encryption_key"] != "01234567890123456789012345678901" {
		t.Fatal("database encryption key was not preserved")
	}

	second, changed, err := configForRemoteWeb(payload)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected an already normalized config to remain unchanged")
	}
	if !bytes.Equal(second, payload) {
		t.Fatal("unchanged config payload differs")
	}
}

func TestLocalLogPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		remote string
		local  string
		ok     bool
	}{
		{remoteStdLog, "std.log", true},
		{remoteAppLog, "cephtower.log", true},
		{remoteAppDir + "/log/cephtower.log.1", "", false},
		{remoteAppDir + "/data/cephtower.db", "", false},
		{remoteAppDir + "/log/nested/file.log", "", false},
	}
	for _, test := range tests {
		got, ok := localLogPath(test.remote)
		if got != test.local || ok != test.ok {
			t.Errorf("localLogPath(%q) = %q, %t; want %q, %t", test.remote, got, ok, test.local, test.ok)
		}
	}
}

func TestResetLocalLogDirRemovesPreviousRuns(t *testing.T) {
	t.Parallel()
	toolDir := t.TempDir()
	legacyLog := filepath.Join(toolDir, ".state", "logs", "192.0.2.10", "20260805-120000", "std.log")
	if err := os.MkdirAll(filepath.Dir(legacyLog), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyLog, []byte("old log"), 0o600); err != nil {
		t.Fatal(err)
	}

	logDir, err := resetLocalLogDir(toolDir)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(toolDir, ".state", "logs")
	if logDir != want {
		t.Fatalf("log directory = %q, want %q", logDir, want)
	}
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty log directory, got %v", entries)
	}
}

func TestReplaceLocalLogOverwritesExistingFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "logs", "std.log")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("old log contents"), 0o600); err != nil {
		t.Fatal(err)
	}
	payload := []byte("new log\n")
	if err := replaceLocalLog(path, int64(len(payload)), func(output io.Writer) error {
		_, err := output.Write(payload)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, payload) {
		t.Fatalf("local log = %q, want %q", raw, payload)
	}
}

func TestUpdateLocalLogAppendsOnlyNewRemoteRange(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "std.log")
	original := []byte("existing log\n")
	appended := []byte("new line\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := updateLocalLog(path, int64(len(original)), int64(len(original)+len(appended)), func(output io.Writer) error {
		_, err := output.Write(appended)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := append(append([]byte{}, original...), appended...)
	if !bytes.Equal(raw, want) {
		t.Fatalf("local log = %q, want %q", raw, want)
	}
}

func TestLocalLogUpdateOffsetUsesIncrementalRangeOnlyForAppend(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "std.log")
	if err := os.WriteFile(path, []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	previous := remoteLogFile{Path: remoteStdLog, Inode: "10", Size: 5, Modified: "1"}
	tests := []struct {
		name     string
		current  remoteLogFile
		known    bool
		expected int64
	}{
		{name: "append", current: remoteLogFile{Path: remoteStdLog, Inode: "10", Size: 9, Modified: "2"}, known: true, expected: 5},
		{name: "initial", current: remoteLogFile{Path: remoteStdLog, Inode: "10", Size: 9, Modified: "2"}, expected: 0},
		{name: "rotation", current: remoteLogFile{Path: remoteStdLog, Inode: "11", Size: 9, Modified: "2"}, known: true, expected: 0},
		{name: "truncation", current: remoteLogFile{Path: remoteStdLog, Inode: "10", Size: 3, Modified: "2"}, known: true, expected: 0},
		{name: "same size rewrite", current: remoteLogFile{Path: remoteStdLog, Inode: "10", Size: 5, Modified: "2"}, known: true, expected: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := localLogUpdateOffset(path, previous, test.current, test.known); got != test.expected {
				t.Fatalf("offset = %d, want %d", got, test.expected)
			}
		})
	}
}

func TestReplaceLocalLogKeepsExistingFileWhenReadIsIncomplete(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "std.log")
	original := []byte("original\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := replaceLocalLog(path, 10, func(output io.Writer) error {
		_, err := output.Write([]byte("short"))
		return err
	}); err == nil {
		t.Fatal("expected incomplete log replacement to fail")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, original) {
		t.Fatalf("local log changed after failed replacement: %q", raw)
	}
}

func TestRemoteServiceCommandsAreValidShell(t *testing.T) {
	t.Parallel()
	commands := map[string]string{
		"start":  startServiceCommand(),
		"stop":   stopServiceCommand(),
		"status": serviceStatusCommand(),
	}
	for name, command := range commands {
		check := exec.Command("sh", "-n", "-c", command)
		if output, err := check.CombinedOutput(); err != nil {
			t.Errorf("%s command is invalid: %v\n%s\n%s", name, err, output, command)
		}
	}
	if !strings.Contains(commands["start"], "/root/cephtower/bin/cephtower") {
		t.Fatal("start command does not execute the remote development binary")
	}
}

func TestArchiveInstallCommandAtomicallyReplacesChangedFiles(t *testing.T) {
	t.Parallel()
	stagingDir := remoteStateDir + "/upload-123"
	command, err := archiveInstallCommand([]string{"backend/cmd/main.go", "frontend/package.json"}, stagingDir)
	if err != nil {
		t.Fatal(err)
	}
	check := exec.Command("sh", "-n", "-c", command)
	if output, err := check.CombinedOutput(); err != nil {
		t.Fatalf("archive install command is invalid: %v\n%s\n%s", err, output, command)
	}
	for _, expected := range []string{
		"tar --no-same-owner -xpf -",
		"mv -f -- '/root/cephtower/.sync-workspace/upload-123/backend/cmd/main.go' '/root/cephtower/backend/cmd/main.go'",
		"mv -f -- '/root/cephtower/.sync-workspace/upload-123/frontend/package.json' '/root/cephtower/frontend/package.json'",
		"rm -rf -- '/root/cephtower/.sync-workspace/upload-123'",
	} {
		if !strings.Contains(command, expected) {
			t.Fatalf("archive install command missing %q:\n%s", expected, command)
		}
	}
}

func TestArchiveInstallCommandRejectsUnsafePaths(t *testing.T) {
	t.Parallel()
	if _, err := archiveInstallCommand([]string{"tools/unsafe.go"}, remoteStateDir+"/upload-123"); err == nil {
		t.Fatal("expected unsafe archive path to be rejected")
	}
	if _, err := archiveInstallCommand([]string{"backend/main.go"}, remoteRoot+"/upload-test"); err == nil {
		t.Fatal("expected unsafe staging directory to be rejected")
	}
	if _, err := archiveInstallCommand([]string{"backend/main.go"}, remoteStateDir+"/upload-../app"); err == nil {
		t.Fatal("expected traversing staging directory to be rejected")
	}
}
