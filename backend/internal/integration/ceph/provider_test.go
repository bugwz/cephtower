package ceph

import (
	"context"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type probeRecordingExecutor struct{ specs []executor.CommandSpec }

func (e *probeRecordingExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	switch spec.ID {
	case "cluster.fsid":
		return executor.CommandResult{Stdout: []byte("fixture-fsid\n")}, nil
	case "cluster.versions":
		return executor.CommandResult{Stdout: []byte(`{"mon":{"ceph version 20.2.2 (sha) tentacle (stable)":1}}`)}, nil
	case "cluster.status":
		return executor.CommandResult{Stdout: []byte(`{}`)}, nil
	default:
		return executor.CommandResult{}, nil
	}
}

func TestProbeUsesSupportedCephFSShellHelpFlag(t *testing.T) {
	runner := &probeRecordingExecutor{}
	provider := NativeProvider{Executor: runner}
	if _, err := provider.Probe(context.Background(), ClusterAccess{MonitorAddresses: "mon:6789"}); err != nil {
		t.Fatal(err)
	}
	for _, spec := range runner.specs {
		if spec.ID == "capability.cephfs_data_access" {
			if spec.Binary != executor.BinaryCephFSShell || !reflect.DeepEqual(spec.Args, []string{"--help"}) {
				t.Fatalf("cephfs probe = %+v", spec)
			}
			return
		}
	}
	t.Fatal("cephfs capability was not probed")
}

func TestCephVersionFromVersions(t *testing.T) {
	payload := []byte(`{
		"mon": {"ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable - RelWithDebInfo)": 3},
		"mgr": {"ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable - RelWithDebInfo)": 2}
	}`)

	if got, want := cephVersionFromVersions(payload), "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)"; got != want {
		t.Fatalf("cephVersionFromVersions() = %q, want %q", got, want)
	}
}

func TestCephVersionFromVersionsRejectsInvalidJSON(t *testing.T) {
	if got := cephVersionFromVersions([]byte(`{"mon":`)); got != "" {
		t.Fatalf("cephVersionFromVersions() = %q, want empty string", got)
	}
}
