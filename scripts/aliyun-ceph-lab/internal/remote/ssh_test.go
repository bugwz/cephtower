package remote

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestVerifyHostKeyAcceptsNewAndRejectsChangedKey(t *testing.T) {
	t.Parallel()
	client := &SSH{KnownHostsPath: filepath.Join(t.TempDir(), "known_hosts")}
	address := &net.TCPAddr{IP: net.ParseIP("203.0.113.10"), Port: 22}
	first := testPublicKey(t)
	if err := client.verifyHostKey("203.0.113.10:22", address, first); err != nil {
		t.Fatalf("verifyHostKey() rejected new key: %v", err)
	}
	if err := client.verifyHostKey("203.0.113.10:22", address, first); err != nil {
		t.Fatalf("verifyHostKey() rejected saved key: %v", err)
	}
	if err := client.verifyHostKey("203.0.113.10:22", address, testPublicKey(t)); err == nil {
		t.Fatal("verifyHostKey() accepted a changed key")
	}
}

func TestHostLogUsesSafeFilenameAndSecurePermissions(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "log")
	client := &SSH{LogDir: dir}
	file, err := client.openHostLog("ceph-node-1")
	if err != nil {
		t.Fatal(err)
	}
	writeHostLogf(file, "INFO", "diagnostic %s", "output")
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "ceph-node-1.log")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "INFO diagnostic output") {
		t.Fatalf("unexpected host log: %q", raw)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("host log permissions = %o, want 600", got)
	}
}

func TestShellQuote(t *testing.T) {
	t.Parallel()
	if got, want := shellQuote("ceph's node"), `'ceph'"'"'s node'`; got != want {
		t.Fatalf("shellQuote() = %q, want %q", got, want)
	}
}

func TestRemoteScriptCommandStreamsToSecureRemoteLog(t *testing.T) {
	t.Parallel()
	command := remoteScriptCommand(
		"/var/log/ceph-lab/init-node.sh.log",
		"/var/log/ceph-lab/status/test.status",
	)
	for _, want := range []string{
		"install -d -m 0755 /var/log/ceph-lab /var/log/ceph-lab/status",
		"chmod 0600",
		"mktemp /tmp/ceph-lab-hook.XXXXXX",
		"cat >\"$script_path\"",
		"bash -o pipefail -c",
		"tee -a",
		"/var/log/ceph-lab/init-node.sh.log",
		"/var/log/ceph-lab/status/test.status",
	} {
		if !strings.Contains(command, want) {
			t.Fatalf("remoteScriptCommand() = %q, want substring %q", command, want)
		}
	}
	if output, err := exec.Command("bash", "-n", "-c", command).CombinedOutput(); err != nil {
		t.Fatalf("remoteScriptCommand() produced invalid shell: %v: %s", err, output)
	}
}

func TestRemoteScriptCommandMaterializesScriptBeforeExecution(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "deploy-ceph.sh.log")
	statusPath := filepath.Join(dir, "deploy-ceph.status")
	command := remoteScriptCommand(logPath, statusPath)
	command = strings.Replace(command,
		"install -d -m 0755 /var/log/ceph-lab /var/log/ceph-lab/status && ", "", 1)
	command = strings.ReplaceAll(command,
		"date --iso-8601=seconds", "date +%Y-%m-%dT%H:%M:%S%z")

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdin = strings.NewReader("printf 'before\\n'\ncat >/dev/null\nprintf 'after\\n'\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("remote script failed: %v: %s", err, output)
	}
	if !strings.Contains(string(output), "before\nafter\n") {
		t.Fatalf("remote script did not continue after a command consumed stdin: %q", output)
	}
	logged, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logged), "before\nafter\n") {
		t.Fatalf("remote log did not contain complete script output: %q", logged)
	}
	status, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(status)); got != "0" {
		t.Fatalf("remote status = %q, want 0", got)
	}
}

func TestNewRunIDIsUniqueAndSafe(t *testing.T) {
	t.Parallel()
	first, err := newRunID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newRunID()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) != 32 || strings.Trim(first, "0123456789abcdef") != "" {
		t.Fatalf("unexpected run IDs: %q %q", first, second)
	}
}

func TestRecoveredScriptResultUsesRemoteExitStatus(t *testing.T) {
	t.Parallel()
	var output strings.Builder
	transportErr := errors.New("missing SSH exit status")
	if err := recoveredScriptResult(&output, "init-node.sh", 0, transportErr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "INFO remote script recovered") {
		t.Fatalf("recovery output = %q", output.String())
	}
	if err := recoveredScriptResult(io.Discard, "init-node.sh", 7, transportErr); err == nil ||
		!strings.Contains(err.Error(), "status 7") {
		t.Fatalf("non-zero remote status error = %v", err)
	}
}

func testPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return publicKey
}
