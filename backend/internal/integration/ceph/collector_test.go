package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

func TestCollectFastStoresCephVersionsHash(t *testing.T) {
	base := fixtureExecutor{t}
	version := "ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable - RelWithDebInfo)"
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.versions": []byte(`{"mon":{"` + version + `":3}}`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "fast")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Kind != "overview" {
			continue
		}
		payload, ok := row.Payload.(cephdomain.Overview)
		if !ok {
			t.Fatalf("overview payload type = %T", row.Payload)
		}
		if payload.CephVersion != "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)" {
			t.Fatalf("overview ceph version = %q", payload.CephVersion)
		}
		if row.SourceVersion != payload.CephVersion {
			t.Fatalf("source version = %q, want %q", row.SourceVersion, payload.CephVersion)
		}
		return
	}
	t.Fatal("overview record was not collected")
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
		"collect.osd_tree":      []byte(`{"nodes":[{"id":-1,"name":"default","type":"root","children":[-3]},{"id":-3,"name":"node-a","type":"host","children":[0]},{"id":0,"name":"osd.0","type":"osd","status":"up","crush_weight":1.0,"device_class":"ssd"}]}`),
		"collect.osd_dump":      []byte(`{"flags":"","osds":[{"osd":0,"up":1,"in":1}]}`),
		"collect.pool":          []byte(`[{"pool_id":1,"pool_name":"pool-a","type":1,"size":3,"min_size":2,"pg_num":8,"pg_placement_num":8,"pg_autoscale_mode":"on","application_metadata":{"rbd":{}},"crush_rule":0,"flags_names":"hashpspool,allow_ec_overwrites","options":{"compression_mode":"passive","compression_algorithm":"zstd","compression_min_blob_size":4096,"compression_max_blob_size":"1048576","compression_required_ratio":"0.875"},"quota_max_bytes":0,"quota_max_objects":0}]`),
		"collect.pool_quota":    []byte(`{"pool_name":"pool-a","pool_id":1,"quota_max_objects":1000,"current_num_objects":0,"quota_max_bytes":1099511627776,"current_num_bytes":0}`),
		"collect.rbd_mirroring": []byte(`{"mirror_mode":"pool"}`),
		"collect.rbd_pool_config": []byte(`[
			{"name":"rbd_qos_bps_limit","value":"0","source":"global","description":"IO byte limit"},
			{"key":"rbd_qos_iops_limit","val":100,"who":"pool:pool-a"}
		]`),
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
			if payload.CrushPath["root"] != "default" || payload.CrushPath["host"] != "node-a" {
				t.Fatalf("osd crush path = %#v", payload.CrushPath)
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
			if payload.PGAutoscaleMode == nil || *payload.PGAutoscaleMode != "on" {
				t.Fatalf("pool autoscale = %v, want on", payload.PGAutoscaleMode)
			}
			if len(payload.Applications) != 1 || payload.Applications[0] != "rbd" {
				t.Fatalf("pool applications = %v, want rbd", payload.Applications)
			}
			if !reflect.DeepEqual(payload.Flags, []string{"hashpspool", "allow_ec_overwrites"}) {
				t.Fatalf("pool flags = %v", payload.Flags)
			}
			if payload.CompressionMode == nil || *payload.CompressionMode != "passive" ||
				payload.CompressionAlgorithm == nil || *payload.CompressionAlgorithm != "zstd" ||
				payload.CompressionMinBlobSize == nil || *payload.CompressionMinBlobSize != 4096 ||
				payload.CompressionMaxBlobSize == nil || *payload.CompressionMaxBlobSize != 1048576 ||
				payload.CompressionRequiredRatio == nil || *payload.CompressionRequiredRatio != 0.875 {
				t.Fatalf("pool compression options were not collected: %#v", payload)
			}
			if payload.QuotaMaxBytes == nil || *payload.QuotaMaxBytes != 1099511627776 || payload.QuotaMaxObjects == nil || *payload.QuotaMaxObjects != 1000 {
				t.Fatalf("pool quotas = %v/%v, want 1099511627776/1000", payload.QuotaMaxBytes, payload.QuotaMaxObjects)
			}
			if payload.RBDMirroring == nil || *payload.RBDMirroring != "pool" {
				t.Fatalf("pool rbd mirroring = %v, want pool", payload.RBDMirroring)
			}
			if payload.RawDetail["pool_name"] != "pool-a" || payload.RawDetail["type"] != "replicated" || payload.RawDetail["application_metadata"] != "rbd" {
				t.Fatalf("pool raw detail = %#v", payload.RawDetail)
			}
			if len(payload.Configuration) != 2 || payload.Configuration[0].Name != "rbd_qos_bps_limit" || payload.Configuration[0].Source != "global" {
				t.Fatalf("pool configuration = %#v", payload.Configuration)
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
				{"path":"/dev/sdb","available":true,"human_readable_type":"ssd","rejected_reasons":[],"sys_api":{"size":"1073741824","rotational":"0","model":"fast-disk","vendor":"fixture","serial":"serial-a"}},
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
	if rows[0].ParentKind != "host" || rows[0].ParentKey != "node-a" {
		t.Fatalf("device parent = %s:%s, want host:node-a", rows[0].ParentKind, rows[0].ParentKey)
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
	if payload.DeviceType == nil || *payload.DeviceType != "ssd" {
		t.Fatalf("device type = %v, want ssd", payload.DeviceType)
	}
	if payload.Model == nil || *payload.Model != "fast-disk" {
		t.Fatalf("device model = %v, want fast-disk", payload.Model)
	}
	if payload.Vendor == nil || *payload.Vendor != "fixture" {
		t.Fatalf("device vendor = %v, want fixture", payload.Vendor)
	}
	if payload.Serial == nil || *payload.Serial != "serial-a" {
		t.Fatalf("device serial = %v, want serial-a", payload.Serial)
	}
}

func TestCollectInventoryAcceptsDecimalSysAPISize(t *testing.T) {
	base := fixtureExecutor{t}
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.device": []byte(`[
			{"name":"ceph-node-1","devices":[
				{"device_id":"disk-a","path":"/dev/vdb","sys_api":{"size":107374182400.0,"rotational":"1"}}
			]}
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
	if payload.SizeBytes == nil || *payload.SizeBytes != 107374182400 {
		t.Fatalf("device size = %v, want 107374182400", payload.SizeBytes)
	}
	if payload.Hostname != "ceph-node-1" || payload.Path != "/dev/vdb" {
		t.Fatalf("device identity = %s:%s, want ceph-node-1:/dev/vdb", payload.Hostname, payload.Path)
	}
}

func TestCollectTopologyNormalizesDaemonCephVersions(t *testing.T) {
	base := fixtureExecutor{t}
	version := "ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable - RelWithDebInfo)"
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.host":   []byte(`[{"hostname":"ceph-node-1","ceph_version":"` + version + `"}]`),
		"collect.daemon": []byte(`[{"daemon_name":"mgr.ceph-node-1.x","daemon_type":"mgr","version":"` + version + `"}]`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology")
	if err != nil {
		t.Fatal(err)
	}
	var sawDaemon bool
	for _, row := range rows {
		switch row.Kind {
		case "daemon":
			payload, ok := row.Payload.(cephdomain.Daemon)
			if !ok {
				t.Fatalf("daemon payload type = %T", row.Payload)
			}
			if payload.Version == nil || *payload.Version != "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)" {
				t.Fatalf("daemon version = %v", payload.Version)
			}
			sawDaemon = true
		}
	}
	if !sawDaemon {
		t.Fatalf("collected daemon=%v, want true", sawDaemon)
	}
}

func TestCollectTopologyKeepsHostNodeFacts(t *testing.T) {
	base := fixtureExecutor{t}
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.host": []byte(`[{
			"hostname":"ceph-node-1",
			"addr":"172.31.0.214",
			"service_instances":[{"type":"osd","count":2}],
			"facts":{
					"system":"Alibaba Cloud Linux 3",
					"platform":"Linux",
					"distro":"alinux",
					"kernel_release":"5.10.134-18.al8.x86_64",
					"kernel_build":"#1 SMP",
					"arch":"x86_64",
					"cpu_model":"Intel Xeon",
				"cpu_cores":4,
				"memory_bytes":8388608
			}
		}]`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology")
	if err != nil {
		t.Fatal(err)
	}
	var payload cephdomain.Host
	for _, row := range rows {
		if row.Kind == "host" {
			var ok bool
			payload, ok = row.Payload.(cephdomain.Host)
			if !ok {
				t.Fatalf("host payload type = %T", row.Payload)
			}
			break
		}
	}
	if payload.Hostname != "ceph-node-1" {
		t.Fatalf("host payload was not collected")
	}
	if payload.System == nil || *payload.System != "Alibaba Cloud Linux 3" {
		t.Fatalf("host system = %v", payload.System)
	}
	if payload.Platform == nil || *payload.Platform != "Linux" {
		t.Fatalf("host platform = %v", payload.Platform)
	}
	if payload.Distro == nil || *payload.Distro != "alinux" {
		t.Fatalf("host distro = %v", payload.Distro)
	}
	if payload.KernelRelease == nil || *payload.KernelRelease != "5.10.134-18.al8.x86_64" {
		t.Fatalf("host kernel release = %v", payload.KernelRelease)
	}
	if payload.KernelBuild == nil || *payload.KernelBuild != "#1 SMP" {
		t.Fatalf("host kernel build = %v", payload.KernelBuild)
	}
	if payload.Arch == nil || *payload.Arch != "x86_64" {
		t.Fatalf("host arch = %v", payload.Arch)
	}
	if payload.CPUCores == nil || *payload.CPUCores != 4 {
		t.Fatalf("host cpu cores = %v", payload.CPUCores)
	}
	if payload.MemoryBytes == nil || *payload.MemoryBytes != 8388608 {
		t.Fatalf("host memory bytes = %v", payload.MemoryBytes)
	}
	if len(payload.ServiceInstances) != 1 || payload.ServiceInstances[0].Type != "osd" || payload.ServiceInstances[0].Count != 2 {
		t.Fatalf("host service instances = %#v", payload.ServiceInstances)
	}
}

func TestCollectTopologyFillsHostFactsFromDaemonMetadata(t *testing.T) {
	base := fixtureExecutor{t}
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.host": []byte(`[{"hostname":"ceph-node-1","addr":"172.31.0.214"}]`),
		"collect.osd_metadata": []byte(`[{
				"hostname":"ceph-node-1",
				"os":"Linux",
				"distro":"centos",
				"distro_description":"CentOS Stream 9",
				"kernel_version":"5.14.0-710.el9.x86_64",
			"kernel_description":"#1 SMP PREEMPT_DYNAMIC Wed May 27 09:04:56 UTC 2026",
			"arch":"x86_64",
			"cpu":"AMD EPYC 7T83 64-Core Processor",
			"mem_total_kb":"15834904"
		}]`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology")
	if err != nil {
		t.Fatal(err)
	}
	var payload cephdomain.Host
	for _, row := range rows {
		if row.Kind == "host" {
			var ok bool
			payload, ok = row.Payload.(cephdomain.Host)
			if !ok {
				t.Fatalf("host payload type = %T", row.Payload)
			}
			break
		}
	}
	if payload.System == nil || *payload.System != "CentOS Stream 9" {
		t.Fatalf("host system = %v", payload.System)
	}
	if payload.Platform == nil || *payload.Platform != "Linux" {
		t.Fatalf("host platform = %v", payload.Platform)
	}
	if payload.Distro == nil || *payload.Distro != "centos" {
		t.Fatalf("host distro = %v", payload.Distro)
	}
	if payload.KernelRelease == nil || *payload.KernelRelease != "5.14.0-710.el9.x86_64" {
		t.Fatalf("host kernel release = %v", payload.KernelRelease)
	}
	if payload.KernelBuild == nil || *payload.KernelBuild != "#1 SMP PREEMPT_DYNAMIC Wed May 27 09:04:56 UTC 2026" {
		t.Fatalf("host kernel build = %v", payload.KernelBuild)
	}
	if payload.Arch == nil || *payload.Arch != "x86_64" {
		t.Fatalf("host arch = %v", payload.Arch)
	}
	if payload.CPUModel == nil || *payload.CPUModel != "AMD EPYC 7T83 64-Core Processor" {
		t.Fatalf("host cpu model = %v", payload.CPUModel)
	}
	if payload.MemoryBytes == nil || *payload.MemoryBytes != 16214941696 {
		t.Fatalf("host memory bytes = %v", payload.MemoryBytes)
	}
}

func TestCollectConfigurationAcceptsSectionOnlyDump(t *testing.T) {
	base := fixtureExecutor{t}
	provider := NativeProvider{Executor: malformedExecutor{base: base, override: map[string][]byte{
		"collect.config": []byte(`[
			{"section":"global","name":"public_network","value":"172.31.0.0/24","level":"advanced","can_update_at_runtime":false,"mask":""},
			{"section":"osd","name":"osd_memory_target","value":"3997510860","level":"basic","can_update_at_runtime":true,"mask":"host:ceph-node-2","location_type":"host","location_value":"ceph-node-2"},
			{"section":"mds","name":"debug_mds","value":"1","level":"advanced","mask":"fs:cephfs"}
		]`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "configuration")
	if err != nil {
		t.Fatal(err)
	}
	configs := map[string]cephdomain.ConfigValue{}
	for _, row := range rows {
		if row.Kind != "config_value" {
			continue
		}
		payload, ok := row.Payload.(cephdomain.ConfigValue)
		if !ok {
			t.Fatalf("config payload type = %T", row.Payload)
		}
		configs[row.NaturalKey] = payload
	}
	global := configs["global:public_network"]
	if global.Who != "global" || global.Name != "public_network" {
		t.Fatalf("global config = %#v", global)
	}
	hostScoped := configs["osd/host:ceph-node-2:osd_memory_target"]
	if hostScoped.Who != "osd/host:ceph-node-2" || hostScoped.LocationType == nil || *hostScoped.LocationType != "host" || hostScoped.LocationValue == nil || *hostScoped.LocationValue != "ceph-node-2" {
		t.Fatalf("host-scoped config = %#v", hostScoped)
	}
	maskScoped := configs["mds/fs:cephfs:debug_mds"]
	if maskScoped.Who != "mds/fs:cephfs" || maskScoped.Mask == nil || *maskScoped.Mask != "fs:cephfs" {
		t.Fatalf("mask-scoped config = %#v", maskScoped)
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
