package mutation

import (
	"context"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type directoryRenameExecutor struct {
	outputs map[string]string
	failID  string
	specs   []executor.CommandSpec
}

func (e *directoryRenameExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.ID == e.failID {
		return executor.CommandResult{}, errors.New("fixture command failure")
	}
	return executor.CommandResult{Stdout: []byte(e.outputs[spec.ID])}, nil
}

func TestCephFSRenameVerifiesBothParents(t *testing.T) {
	const source = "drwxr-xr-x 0 0 0 2026-10-01 13:00:00 old/\n"
	const destination = "drwxr-xr-x 0 0 0 2026-10-01 13:00:00 new/\n"
	for _, tt := range []struct {
		name, pre, dest, after, failID, code string
		commands                             int
	}{
		{"success", source, destination, "", "", "", 4},
		{"source missing", "", destination, "", "", "invalid_request", 1},
		{"source is file", "-rw-r--r-- 0 0 0 2026-10-01 13:00:00 old\n", destination, "", "", "invalid_request", 1},
		{"target missing", source, "", "", "", "post_check_failed", 3},
		{"target malformed", source, "unparseable", "", "", "post_check_failed", 3},
		{"source remains", source, destination, source, "", "post_check_failed", 4},
		{"source readback malformed", source, destination, "unparseable", "", "post_check_failed", 4},
		{"readback fails", source, destination, "", "cephfs_entry.rename.path_post_check", "post_check_failed", 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{
				"cephfs_entry.rename.pre_check":              tt.pre,
				"cephfs_entry.rename.destination_post_check": tt.dest,
				"cephfs_entry.rename.path_post_check":        tt.after,
			}, failID: tt.failID}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "cephfs_entry.rename", ResourceKey: "filesystem/cephfs/entry", Parameters: map[string]any{"path": "/projects/old", "destination": "/archive/new"}})
			if tt.code == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != tt.code {
					t.Fatalf("error = %v, want %s", err, tt.code)
				}
			}
			if len(runner.specs) != tt.commands {
				t.Fatalf("executed %d commands, want %d", len(runner.specs), tt.commands)
			}
		})
	}
}
