package clusterinspect

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestCephFSDirectoriesListsChildrenAndReadsQuotas(t *testing.T) {
	service, runner, clusterID := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		switch spec.ID {
		case "cephfs.directory.list":
			return executor.CommandResult{Stdout: []byte("drwxr-xr-x           0 0 0 2026-09-23 12:00:00 ./\n" +
				"drwxr-xr-x           0 0 0 2026-09-23 12:00:00 ../\n" +
				"drwxrwx---        4096 1000 1001 2026-09-22 09:30:00 shared projects/\n" +
				"-rw-r--r--          12 1000 1001 2026-09-22 09:31:00 readme.txt\n")}, nil
		case "cephfs.directory.quota":
			if spec.Args[len(spec.Args)-1] != `"/shared projects"` {
				t.Fatalf("quota path = %q", spec.Args[len(spec.Args)-1])
			}
			return executor.CommandResult{Stdout: []byte("max_bytes: 4096\nmax_files: 20\n")}, nil
		default:
			t.Fatalf("unexpected command: %+v", spec)
			return executor.CommandResult{}, nil
		}
	}
	result, err := service.CephFSDirectories(context.Background(), clusterID, "cephfs", "/")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 2 || result.Items[0].Path != "/" || result.Items[0].Quotas != nil {
		t.Fatalf("items = %+v", result.Items)
	}
	child := result.Items[1]
	if child.Name != "shared projects" || child.Path != "/shared projects" || child.Quotas == nil || child.Quotas.MaxBytes != 4096 || child.Quotas.MaxFiles != 20 {
		t.Fatalf("child = %+v", child)
	}
	wantList := []string{"--fs", "cephfs", "ls", "-la", "/"}
	if !reflect.DeepEqual(runner.specs[0].Args, wantList) || runner.specs[0].Binary != executor.BinaryCephFSShell {
		t.Fatalf("list spec = %+v", runner.specs[0])
	}
}

func TestCephFSDirectoriesTreatsMissingQuotaAsUnlimited(t *testing.T) {
	service, runner, clusterID := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.ID == "cephfs.directory.list" {
			return executor.CommandResult{Stdout: []byte("drwxr-xr-x 0 0 0 2026-09-23 12:00:00 child/\n")}, nil
		}
		return executor.CommandResult{}, &executor.Error{Kind: "exit", ExitCode: 9, Summary: "quota is not set"}
	}
	result, err := service.CephFSDirectories(context.Background(), clusterID, "cephfs", "/data")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 2 || result.Items[0].Quotas == nil || *result.Items[0].Quotas != (CephFSQuota{}) || *result.Items[1].Quotas != (CephFSQuota{}) {
		t.Fatalf("items = %+v", result.Items)
	}
}

func TestCephFSDirectoriesRejectsUnsafeScopeAndMalformedOutput(t *testing.T) {
	service, runner, clusterID := testInspection(t)
	for _, scope := range []struct{ fs, path string }{{"", "/"}, {"--help", "/"}, {"cephfs", "relative"}, {"cephfs", "/bad,path"}} {
		if _, err := service.CephFSDirectories(context.Background(), clusterID, scope.fs, scope.path); err == nil {
			t.Fatalf("accepted scope %+v", scope)
		}
	}
	if len(runner.specs) != 0 {
		t.Fatal("unsafe scope executed")
	}
	runner.output = "not an ls line"
	if _, err := service.CephFSDirectories(context.Background(), clusterID, "cephfs", "/"); err == nil {
		t.Fatal("malformed listing accepted")
	}
	runner.run = func(executor.CommandSpec) (executor.CommandResult, error) {
		return executor.CommandResult{}, errors.New("offline")
	}
	if _, err := service.CephFSDirectories(context.Background(), clusterID, "cephfs", "/"); err == nil {
		t.Fatal("command failure became an empty listing")
	}
}

func TestParseCephFSQuotaAcceptsPartialReadback(t *testing.T) {
	quota, err := parseCephFSQuota([]byte("max_files: 9\n"))
	if err != nil || quota.MaxBytes != 0 || quota.MaxFiles != 9 {
		t.Fatalf("quota=%+v err=%v", quota, err)
	}
	if _, err := parseCephFSQuota([]byte("max_bytes: nope\n")); err == nil {
		t.Fatal("invalid quota accepted")
	}
}

func TestCephFSSnapshotsListsVisibleSnapshotDirectories(t *testing.T) {
	service, runner, clusterID := testInspection(t)
	runner.output = "drwxr-xr-x 0 0 0 2026-09-23 13:00:00 snap one/\n" +
		"drwxr-xr-x 0 0 0 2026-09-23 13:01:00 _internal/\n" +
		"-rw-r--r-- 4 0 0 2026-09-23 13:02:00 not-a-snapshot\n"
	result, err := service.CephFSSnapshots(context.Background(), clusterID, "cephfs", "/projects")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "snap one" || result.Items[0].Path != "/projects/.snap/snap one" {
		t.Fatalf("snapshots = %+v", result.Items)
	}
	want := []string{"--fs", "cephfs", "ls", "-la", "/projects/.snap"}
	if len(runner.specs) != 1 || runner.specs[0].ID != "cephfs.snapshot.list" || !reflect.DeepEqual(runner.specs[0].Args, want) {
		t.Fatalf("spec = %+v", runner.specs)
	}
}
