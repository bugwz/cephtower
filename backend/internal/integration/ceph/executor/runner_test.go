package executor

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunnerExecutesWithoutShellAndLimitsOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX process groups")
	}
	dir := t.TempDir()
	fake := filepath.Join(dir, "ceph")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf 'ok'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Paths: map[Binary]string{BinaryCeph: fake}, TempRoot: dir}
	result, err := runner.Run(context.Background(), ClusterAccess{MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "secret"}, CommandSpec{ID: "test", Binary: BinaryCeph, Timeout: 5 * time.Second, MaxOutput: 16})
	if err != nil || string(result.Stdout) != "ok" {
		t.Fatalf("Run() = %q, %v", result.Stdout, err)
	}
}

func TestRunnerConfiguresCephFSShellWithoutUnsupportedArguments(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX fixture executable")
	}
	dir := t.TempDir()
	fake := filepath.Join(dir, "cephfs-shell")
	script := `#!/bin/sh
test -n "$CEPH_CONF" || exit 21
test -f "$CEPHFS_SHELL_CONF" || exit 23
grep -q 'colors = False' "$CEPHFS_SHELL_CONF" || exit 24
case "$CEPH_ARGS" in
  *"--name=client.test"*"--keyring="*"--client_snapdir=.snap"*) ;;
  *) exit 22 ;;
esac
printf '%s\n' "$@"
`
	if err := os.WriteFile(fake, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Paths: map[Binary]string{BinaryCephFSShell: fake}, TempRoot: dir}
	spec := CommandSpec{ID: "test.cephfs", Binary: BinaryCephFSShell, Args: []string{"--fs", "cephfs", "stat", "/"}, Timeout: 5 * time.Second}
	result, err := runner.Run(context.Background(), ClusterAccess{MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "secret"}, spec)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(string(result.Stdout)); !reflect.DeepEqual(got, spec.Args) {
		t.Fatalf("argv = %v, want %v", got, spec.Args)
	}
	if !reflect.DeepEqual(result.RedactedArgs, spec.Args) {
		t.Fatalf("redacted args = %v", result.RedactedArgs)
	}
}

func TestRunnerRejectsUnsafeCephFSShellClientName(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "cephfs-shell")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	runner := Runner{Paths: map[Binary]string{BinaryCephFSShell: fake}, TempRoot: dir}
	_, err := runner.Run(context.Background(), ClusterAccess{MonitorAddresses: "mon:6789", ClientUsername: "client.test --keyring=other", ClientKey: "secret"}, CommandSpec{ID: "test.cephfs", Binary: BinaryCephFSShell, Args: []string{"stat", "/"}, Timeout: time.Second})
	if err == nil {
		t.Fatal("unsafe cephfs-shell client name accepted")
	}
}
func TestRunnerRejectsUnknownBinary(t *testing.T) {
	_, err := (&Runner{}).Run(context.Background(), ClusterAccess{}, CommandSpec{ID: "bad", Binary: "sh", Timeout: time.Second})
	if err == nil {
		t.Fatal("unknown binary accepted")
	}
}
