package executor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
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
func TestRunnerRejectsUnknownBinary(t *testing.T) {
	_, err := (&Runner{}).Run(context.Background(), ClusterAccess{}, CommandSpec{ID: "bad", Binary: "sh", Timeout: time.Second})
	if err == nil {
		t.Fatal("unknown binary accepted")
	}
}
