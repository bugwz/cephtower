package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
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

func TestCollectStorageRecordsEmptyOSDFlags(t *testing.T) {
	base := fixtureExecutor{t}
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{"collect.osd_dump": []byte(`{"flags":"","osds":[]}`)}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "storage")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Kind != "osd_flag" || row.NaturalKey != "flags" {
			continue
		}
		payload, ok := row.Payload.(map[string]any)
		if !ok {
			t.Fatalf("osd_flag payload type = %T", row.Payload)
		}
		flags, ok := payload["flags"].([]string)
		if !ok {
			t.Fatalf("osd_flag flags type = %T", payload["flags"])
		}
		if len(flags) != 0 {
			t.Fatalf("osd_flag flags = %v, want empty", flags)
		}
		return
	}
	t.Fatal("osd_flag record was not collected")
}

func TestCollectStorageAcceptsNumericPoolCrushRuleAndMapsOSDHost(t *testing.T) {
	base := fixtureExecutor{t}
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.osd_tree": []byte(`{"nodes":[{"id":-1,"name":"default","type":"root","children":[-3]},{"id":-3,"name":"node-a","type":"host","children":[0]},{"id":0,"name":"osd.0","type":"osd","status":"up","crush_weight":1.0,"device_class":"ssd"}]}`),
		"collect.osd_dump": []byte(`{"flags":"","osds":[{"osd":0,"up":1,"in":1}]}`),
		"collect.pool":     []byte(`[{"pool":1,"pool_name":"pool-a","type":1,"size":3,"min_size":2,"pg_num":8,"pg_placement_num":8,"crush_rule":0}]`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "storage")
	if err != nil {
		t.Fatal(err)
	}
	var sawOSD, sawPool bool
	for _, row := range rows {
		switch row.Kind {
		case "osd":
			payload, ok := row.Payload.(cephdomain.OSD)
			if !ok {
				t.Fatalf("osd payload type = %T", row.Payload)
			}
			if payload.Host == nil || *payload.Host != "node-a" {
				t.Fatalf("osd host = %v, want node-a", payload.Host)
			}
			sawOSD = true
		case "pool":
			payload, ok := row.Payload.(cephdomain.Pool)
			if !ok {
				t.Fatalf("pool payload type = %T", row.Payload)
			}
			if payload.CrushRule == nil || *payload.CrushRule != "0" {
				t.Fatalf("pool crush rule = %v, want 0", payload.CrushRule)
			}
			sawPool = true
		}
	}
	if !sawOSD || !sawPool {
		t.Fatalf("collected osd=%v pool=%v, want both", sawOSD, sawPool)
	}
}

func TestCollectInventoryAcceptsNestedDevicesAndSkipsPlaceholders(t *testing.T) {
	base := fixtureExecutor{t}
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.device": []byte(`[
			{"name":"node-a","devices":[
				{"path":"/dev/sdb","available":true,"rejected_reasons":[],"sys_api":{"size":"1073741824","rotational":"0","model":"fast-disk","vendor":"fixture","serial":"serial-a"}},
				{"available":false}
			]},
			{"hostname":"node-b"}
		]`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "inventory")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("inventory rows = %d, want 1", len(rows))
	}
	payload, ok := rows[0].Payload.(cephdomain.Device)
	if !ok {
		t.Fatalf("device payload type = %T", rows[0].Payload)
	}
	if payload.Hostname != "node-a" || payload.Path != "/dev/sdb" {
		t.Fatalf("device identity = %s:%s, want node-a:/dev/sdb", payload.Hostname, payload.Path)
	}
	if payload.SizeBytes == nil || *payload.SizeBytes != 1073741824 {
		t.Fatalf("device size = %v, want 1073741824", payload.SizeBytes)
	}
	if payload.Rotational == nil || *payload.Rotational {
		t.Fatalf("device rotational = %v, want false", payload.Rotational)
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
