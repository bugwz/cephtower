package ceph

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type fixtureExecutor struct{ t *testing.T }

func (f fixtureExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	names := map[string]string{"collect.status": "status.json", "collect.df": "df.json", "collect.host": "host.json", "collect.daemon": "daemon.json", "collect.service": "service.json", "collect.mon": "mon.json", "collect.quorum": "quorum.json", "collect.mgr": "mgr.json", "collect.osd_tree": "osd-tree.json", "collect.osd_dump": "osd-dump.json", "collect.pool": "pool.json", "collect.fs": "fs.json", "collect.cephfs_subvolume": "cephfs-subvolume.json", "collect.rbd_image": "rbd-image.json", "collect.rgw_status": "rgw-realm.json", "collect.nfs_cluster": "nfs-cluster.json", "collect.smb_cluster": "smb-cluster.json", "collect.device": "device.json", "collect.config": "config.json"}
	names["collect.mds_fs"] = "fs.json"
	names["collect.upgrade"] = "upgrade.json"
	name := names[spec.ID]
	if name == "" {
		return executor.CommandResult{}, fmt.Errorf("fixture is unavailable for %s", spec.ID)
	}
	data, err := os.ReadFile(filepath.Join("testdata", "v20.2.2", name))
	return executor.CommandResult{Stdout: data}, err
}
func TestCollectParsesCeph2022Fixtures(t *testing.T) {
	provider := NativeProvider{Executor: fixtureExecutor{t}}
	access := ClusterAccess{}
	for _, test := range []struct {
		module  string
		minimum int
	}{{"fast", 1}, {"topology", 6}, {"storage", 8}, {"inventory", 1}, {"configuration", 1}} {
		rows, err := provider.Collect(context.Background(), access, test.module)
		if err != nil || len(rows) < test.minimum {
			t.Fatalf("Collect(%s) = %d rows, %v", test.module, len(rows), err)
		}
	}
}

type malformedExecutor struct {
	base     executor.Executor
	override map[string][]byte
}

func (f malformedExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if value, ok := f.override[spec.ID]; ok {
		return executor.CommandResult{Stdout: value}, nil
	}
	return f.base.Run(ctx, access, spec)
}

func TestCollectRejectsMissingNullAndOverflowedCoreFields(t *testing.T) {
	base := fixtureExecutor{t}
	tests := []struct {
		name, module, command string
		payload               string
	}{
		{"missing status fields", "fast", "collect.status", `{}`},
		{"null capacity", "fast", "collect.df", `{"stats":{"total_bytes":null,"total_used_bytes":1,"total_avail_bytes":1}}`},
		{"overflow capacity", "fast", "collect.df", `{"stats":{"total_bytes":18446744073709551616,"total_used_bytes":1,"total_avail_bytes":1}}`},
		{"missing host name", "topology", "collect.host", `[{}]`},
		{"missing pool name", "storage", "collect.pool", `[{"pool":1}]`},
		{"missing device path", "inventory", "collect.device", `[{"hostname":"node"}]`},
		{"missing config name", "configuration", "collect.config", `[{"who":"global"}]`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{test.command: []byte(test.payload)}}}
			if _, err := provider.Collect(context.Background(), ClusterAccess{}, test.module); err == nil {
				t.Fatal("malformed fixture was accepted")
			}
		})
	}
}
