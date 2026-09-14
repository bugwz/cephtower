package mutation

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestEveryNativeActionBuildsRegisteredCommand(t *testing.T) {
	image := base64.RawURLEncoding.EncodeToString([]byte("rbd/image"))
	pair := func(left, right string) string {
		return base64.RawURLEncoding.EncodeToString([]byte(left + "\x00" + right))
	}
	tests := []struct {
		action, key string
		parameters  map[string]any
	}{
		{"cluster.refresh", "cluster/1", nil},
		{"health.mute", "health/mute/HEALTH_WARN", nil}, {"health.unmute", "health/mute/HEALTH_WARN", nil},
		{"host.create", "host", map[string]any{"hostname": "node1", "address": "192.0.2.10", "labels": []any{"storage"}, "maintenance": true}},
		{"host.update", "host/node1", map[string]any{"labels_add": []any{"storage"}, "labels_remove": []any{"old"}}},
		{"host.delete", "host/node1", nil}, {"host.action", "host/node1/action", map[string]any{"action": "rescan"}},
		{"device.identify", "host/node1/identify-device", map[string]any{"device": "/dev/sdb", "state": "on"}},
		{"service.create", "service", map[string]any{"service_type": "mon", "service_id": "mon"}},
		{"service.update", "service/mon", map[string]any{"service_type": "mon"}}, {"service.delete", "service/mon", nil},
		{"daemon.action", "daemon/osd.1/action", map[string]any{"action": "restart"}},
		{"upgrade.check", "upgrade/check", map[string]any{"version": "20.2.2"}},
		{"upgrade.action", "upgrade/action", map[string]any{"action": "start", "version": "20.2.2"}},
		{"manager.fail", "manager/mgr.a/fail", nil}, {"monitor.action", "monitor/action", map[string]any{"action": "scrub"}},
		{"manager_module.update", "manager-module/prometheus", map[string]any{"enabled": true}},
		{"osd.action", "osd/1/action", map[string]any{"action": "deep-scrub"}},
		{"osd_flag.update", "osd-flag", map[string]any{"action": "set", "flag": "noout"}},
		{"osd.removal_check", "osd/removal-check", map[string]any{"osd_ids": []any{"1", "2"}}},
		{"osd.delete", "osd/1", map[string]any{"zap": true}},
		{"osd_deployment.preview", "osd-deployment/preview", map[string]any{"data_devices": map[string]any{"all": true}}},
		{"osd_deployment.create", "osd-deployment", map[string]any{"data_devices": map[string]any{"paths": []any{"/dev/sdb"}}}},
		{"device.zap", "host/node1/device/" + pair("node1", "/dev/sdb") + "/zap", nil},
		{"crush_rule.create", "crush-rule", map[string]any{"name": "ssd", "root": "default", "failure_domain": "host"}},
		{"crush_rule.update", "crush-rule/old", map[string]any{"name": "new"}}, {"crush_rule.delete", "crush-rule/old", nil},
		{"erasure_code_profile.create", "erasure-code-profile", map[string]any{"name": "ec", "plugin": "jerasure", "k": "2", "m": "1"}},
		{"erasure_code_profile.delete", "erasure-code-profile/ec", nil},
		{"pool.create", "pool", map[string]any{"name": "data", "pg_num": "32"}},
		{"pool.update", "pool/data", map[string]any{"field": "size", "value": "3"}}, {"pool.delete", "pool/data", nil},
		{"rbd_image.create", "rbd/image", map[string]any{"image_spec": "rbd/image", "size": "1024"}},
		{"rbd_image.update", "rbd/image/" + image, map[string]any{"size": "2048"}}, {"rbd_image.delete", "rbd/image/" + image, nil},
		{"rbd_image.action", "rbd/image/" + image + "/action", map[string]any{"action": "flatten"}},
		{"rbd_snapshot.create", "rbd/image/" + image + "/snapshot", map[string]any{"name": "snap1"}},
		{"rbd_snapshot.update", "rbd/image/" + image + "/snapshot/snap1", map[string]any{"name": "snap2"}},
		{"rbd_snapshot.delete", "rbd/image/" + image + "/snapshot/snap1", nil},
		{"rbd_snapshot.action", "rbd/image/" + image + "/snapshot/snap1/action", map[string]any{"action": "rollback"}},
		{"rbd_namespace.create", "rbd/namespace", map[string]any{"pool": "rbd", "name": "ns"}},
		{"rbd_namespace.delete", "rbd/namespace/rbd/ns", nil},
		{"rbd_trash.restore", "rbd/trash/image-id/restore", map[string]any{"pool": "rbd", "name": "restored"}},
		{"rbd_trash.delete", "rbd/trash/" + pair("rbd", "image-id"), nil},
		{"rbd_trash.purge", "rbd/trash/purge", map[string]any{"pool": "rbd"}},
		{"rbd_group.create", "rbd/group", map[string]any{"pool": "rbd", "name": "group1"}},
		{"rbd_mirroring.update", "rbd/mirroring", map[string]any{"pool": "rbd", "mode": "image"}},
		{"filesystem.create", "filesystem", map[string]any{"name": "cephfs"}}, {"filesystem.update", "filesystem/cephfs", map[string]any{"max_mds": "2"}}, {"filesystem.delete", "filesystem/cephfs", nil},
		{"subvolume_group.create", "filesystem/cephfs/subvolume-group", map[string]any{"name": "group", "pool": "cephfs.data"}},
		{"subvolume_group.update", "filesystem/cephfs/subvolume-group/group", map[string]any{"size": "1024"}}, {"subvolume_group.delete", "filesystem/cephfs/subvolume-group/group", nil},
		{"subvolume.create", "filesystem/cephfs/subvolume", map[string]any{"name": "sub", "group": "_nogroup", "pool": "cephfs.data"}},
		{"subvolume.update", "filesystem/cephfs/subvolume/sub", map[string]any{"size": "2048"}}, {"subvolume.delete", "filesystem/cephfs/subvolume/sub", nil},
		{"cephfs_snapshot.create", "filesystem/cephfs/subvolume/sub/snapshot", map[string]any{"name": "snap"}},
		{"cephfs_snapshot.delete", "filesystem/cephfs/subvolume/sub/snapshot/snap", nil},
		{"cephfs_snapshot.clone", "filesystem/cephfs/subvolume/sub/snapshot/snap/clone", map[string]any{"target": "clone"}},
		{"snapshot_schedule.create", "filesystem/cephfs/snapshot-schedule", map[string]any{"path": "/", "schedule": "1h"}},
		{"cephfs_authorization.create", "filesystem/cephfs/authorization", map[string]any{"client": "client.app", "path": "/", "access": "rw"}},
		{"cephfs_client.evict", "filesystem/cephfs/client/123", nil},
		{"cephfs_entry.quota", "filesystem/cephfs/entry/quota", map[string]any{"path": "/data", "max_bytes": "1024"}},
		{"rgw_user.create", "rgw/user", map[string]any{"uid": "user1", "display_name": "User One"}},
		{"rgw_user.update", "rgw/user/user1", map[string]any{"email": "user1@example.test"}}, {"rgw_user.delete", "rgw/user/user1", nil},
		{"rgw_account.create", "rgw/account", map[string]any{"account_id": "RGW00000000000000001", "account_name": "Account One"}},
		{"rgw_role.delete", "rgw/role/reader", map[string]any{"name": "reader"}},
		{"rgw_role.create", "rgw/role", map[string]any{"name": "reader", "path": "/", "assume_role_policy": `{"Version":"2012-10-17","Statement":[]}`}},
		{"rgw_key.create", "rgw/user/user1/key", map[string]any{"access_key": "ACCESS123", "secret_key": "SECRET123"}},
		{"rgw_key.delete", "rgw/user/user1/key", map[string]any{"access_key": "ACCESS123"}},
		{"rgw_realm.create", "rgw/realm", map[string]any{"name": "realm1"}}, {"rgw_zonegroup.create", "rgw/zonegroup", map[string]any{"name": "zg1"}}, {"rgw_zone.create", "rgw/zone", map[string]any{"name": "zone1"}}, {"rgw_period.commit", "rgw/period/commit", nil},
		{"nfs_cluster.create", "nfs/cluster", map[string]any{"name": "nfs1"}}, {"nfs_cluster.delete", "nfs/cluster/nfs1", nil},
		{"nfs_export.create", "nfs/export", map[string]any{"cluster": "nfs1", "pseudo": "/export", "path": "/data", "filesystem": "cephfs"}},
		{"nfs_export.update", "nfs/export/" + pair("nfs1", "/export"), map[string]any{"cluster": "nfs1", "pseudo": "/export", "path": "/data", "filesystem": "cephfs"}},
		{"nfs_export.delete", "nfs/export/" + pair("nfs1", "/export"), nil},
		{"smb_cluster.create", "smb/cluster", map[string]any{"name": "smb1", "auth_mode": "user"}}, {"smb_cluster.update", "smb/cluster/smb1", map[string]any{"auth_mode": "user"}}, {"smb_cluster.delete", "smb/cluster/smb1", nil},
		{"smb_share.create", "smb/share", map[string]any{"cluster": "smb1", "name": "share1", "filesystem": "cephfs", "path": "/data"}},
		{"smb_share.update", "smb/share/" + pair("smb1", "share1"), map[string]any{"cluster": "smb1", "filesystem": "cephfs"}}, {"smb_share.delete", "smb/share/" + pair("smb1", "share1"), nil},
		{"config_value.set", "configuration/value/" + pair("global", "osd_pool_default_size"), map[string]any{"value": "3"}}, {"config_value.delete", "configuration/value/" + pair("global", "osd_pool_default_size"), nil},
	}
	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			command, err := build(Request{Action: test.action, ResourceKey: test.key}, test.parameters)
			if err != nil {
				t.Fatal(err)
			}
			if command.binary == "" || len(command.args) == 0 || command.timeout <= 0 {
				t.Fatalf("incomplete command: %#v", command)
			}
			if !Supports(test.action) {
				t.Fatalf("built action %q is absent from Supports", test.action)
			}
		})
	}
}

func TestHostCreateBuildsLabelsAndMaintenance(t *testing.T) {
	command, err := build(Request{Action: "host.create", ResourceKey: "host"}, map[string]any{
		"hostname": "node1", "address": "192.0.2.10",
		"labels": []any{"_admin", "osd"}, "maintenance": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"orch", "host", "add", "node1", "--addr", "192.0.2.10", "--labels", "_admin,osd", "--maintenance"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestCephFSCreateCommandsIncludeFormOptions(t *testing.T) {
	filesystem, err := build(Request{Action: "filesystem.create", ResourceKey: "filesystem"}, map[string]any{
		"name": "cephfs", "placement": "node-1;node-2", "metadata_pool": "cephfs.meta", "data_pool": "cephfs.data",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantFilesystem := []string{"fs", "volume", "create", "cephfs", "node-1;node-2", "cephfs.meta", "cephfs.data"}
	if !reflect.DeepEqual(filesystem.args, wantFilesystem) {
		t.Fatalf("filesystem args = %#v, want %#v", filesystem.args, wantFilesystem)
	}

	subvolume, err := build(Request{Action: "subvolume.create", ResourceKey: "filesystem/cephfs/subvolume"}, map[string]any{
		"name": "home", "group": "users", "size": "10737418240", "pool": "cephfs.data",
		"uid": "1000", "gid": "1000", "mode": "0750", "namespace_isolated": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantSubvolume := []string{
		"fs", "subvolume", "create", "cephfs", "home", "10737418240", "users",
		"cephfs.data", "1000", "1000", "0750", "--namespace-isolated",
	}
	if !reflect.DeepEqual(subvolume.args, wantSubvolume) {
		t.Fatalf("subvolume args = %#v, want %#v", subvolume.args, wantSubvolume)
	}
}

func TestFilesystemCreateRequiresPoolPair(t *testing.T) {
	_, err := build(Request{Action: "filesystem.create", ResourceKey: "filesystem"}, map[string]any{
		"name": "cephfs", "metadata_pool": "cephfs.meta",
	})
	if err == nil {
		t.Fatal("filesystem create accepted a metadata pool without a data pool")
	}
}

func TestHostUpdateBuildsLabelDiff(t *testing.T) {
	command, err := build(Request{Action: "host.update", ResourceKey: "host/node1"}, map[string]any{
		"labels_add": []any{"osd", "mgr"}, "labels_remove": []any{"mon"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"orch", "host", "label", "add", "node1", "osd"},
		{"orch", "host", "label", "add", "node1", "mgr"},
		{"orch", "host", "label", "rm", "node1", "mon"},
	}
	got := [][]string{command.args}
	for _, followup := range command.followups {
		got = append(got, followup.args)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestHostMaintenanceForceBuildsSafetyFlags(t *testing.T) {
	command, err := build(Request{Action: "host.action", ResourceKey: "host/node1/action"}, map[string]any{
		"action": "maintenance_enter", "force": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"orch", "host", "maintenance", "enter", "node1", "--force", "--yes-i-really-mean-it"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestRGWKeyArgumentsAreSensitive(t *testing.T) {
	command, err := build(Request{Action: "rgw_key.create", ResourceKey: "rgw/user/user1/key"}, map[string]any{"access_key": "ACCESS123", "secret_key": "SECRET123"})
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range []int{7, 9} {
		if _, ok := command.sensitive[index]; !ok {
			t.Fatalf("argument %d is not marked sensitive", index)
		}
	}
}

func TestPoolCreateBuildsErasurePoolCommand(t *testing.T) {
	command, err := build(Request{Action: "pool.create", ResourceKey: "pool"}, map[string]any{
		"name":                 "ec-data",
		"pg_num":               "32",
		"pool_type":            "erasure",
		"erasure_code_profile": "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"osd", "pool", "create", "ec-data", "32", "32", "erasure", "default"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestPoolCreateBuildsRBDMirroringFollowup(t *testing.T) {
	command, err := build(Request{Action: "pool.create", ResourceKey: "pool"}, map[string]any{
		"name":          "rbd",
		"pool_type":     "replicated",
		"pg_num":        "32",
		"rbd_mirroring": "pool",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(command.followups) != 1 {
		t.Fatalf("followups = %d, want 1", len(command.followups))
	}
	want := []string{"mirror", "pool", "enable", "rbd", "pool"}
	if !reflect.DeepEqual(command.followups[0].args, want) {
		t.Fatalf("followup args = %#v, want %#v", command.followups[0].args, want)
	}
}

func TestPoolCreateBuildsErasureFlagsAndMirroringFollowups(t *testing.T) {
	command, err := build(Request{Action: "pool.create", ResourceKey: "pool"}, map[string]any{
		"name":          "ec-rbd",
		"pool_type":     "erasure",
		"pg_num":        "32",
		"flags":         []any{"allow_ec_overwrites"},
		"applications":  []any{"rbd"},
		"rbd_mirroring": "pool",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(command.followups) != 3 {
		t.Fatalf("followups = %d, want 3", len(command.followups))
	}
	want := [][]string{
		{"osd", "pool", "set", "ec-rbd", "allow_ec_overwrites", "true"},
		{"osd", "pool", "application", "enable", "ec-rbd", "rbd"},
		{"mirror", "pool", "enable", "ec-rbd", "pool"},
	}
	for index := range want {
		if !reflect.DeepEqual(command.followups[index].args, want[index]) {
			t.Fatalf("followup %d args = %#v, want %#v", index, command.followups[index].args, want[index])
		}
	}
}

func TestPoolUpdateBuildsErasureFlagCommand(t *testing.T) {
	command, err := build(Request{Action: "pool.update", ResourceKey: "pool/ec-rbd"}, map[string]any{
		"field": "allow_ec_overwrites",
		"value": "false",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"osd", "pool", "set", "ec-rbd", "allow_ec_overwrites", "false"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestPoolUpdateBuildsRBDMirroringDisableCommand(t *testing.T) {
	command, err := build(Request{Action: "pool.update", ResourceKey: "pool/rbd"}, map[string]any{
		"operation":     "rbd_mirroring",
		"rbd_mirroring": "disabled",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"mirror", "pool", "disable", "rbd"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
	if wantCheck := []string{"mirror", "pool", "info", "rbd", "--format", "json"}; !reflect.DeepEqual(command.check, wantCheck) {
		t.Fatalf("check = %#v, want %#v", command.check, wantCheck)
	}
}

func TestErasureCodeProfileCreateBuildsAllProfileArguments(t *testing.T) {
	command, err := build(Request{Action: "erasure_code_profile.create", ResourceKey: "erasure-code-profile"}, map[string]any{
		"name": "ec-isa", "plugin": "isa", "k": "7", "m": "3", "technique": "reed_sol_van",
		"crush-failure-domain": "host", "crush-num-failure-domains": "3",
		"crush-osds-per-failure-domain": "2", "crush-root": "default",
		"crush-device-class": "ssd", "directory": "/usr/lib64/ceph/erasure-code",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"osd", "erasure-code-profile", "set", "ec-isa", "plugin=isa", "k=7", "m=3",
		"technique=reed_sol_van", "crush-failure-domain=host", "crush-num-failure-domains=3",
		"crush-osds-per-failure-domain=2", "crush-root=default", "crush-device-class=ssd",
		"directory=/usr/lib64/ceph/erasure-code",
	}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestPoolCreateBuildsConfigurationFollowups(t *testing.T) {
	command, err := build(Request{Action: "pool.create", ResourceKey: "pool"}, map[string]any{
		"name":              "data",
		"pg_num":            "32",
		"pool_type":         "replicated",
		"pg_autoscale_mode": "on",
		"size":              "3",
		"crush_rule":        "replicated_rule",
		"compression_mode":  "passive",
		"applications":      []any{"rbd", "cephfs"},
		"quota_max_bytes":   "1099511627776",
		"quota_max_objects": "1000",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"osd", "pool", "set", "data", "pg_autoscale_mode", "on"},
		{"osd", "pool", "set", "data", "size", "3"},
		{"osd", "pool", "set", "data", "crush_rule", "replicated_rule"},
		{"osd", "pool", "set", "data", "compression_mode", "passive"},
		{"osd", "pool", "application", "enable", "data", "rbd"},
		{"osd", "pool", "application", "enable", "data", "cephfs"},
		{"osd", "pool", "set-quota", "data", "max_bytes", "1099511627776"},
		{"osd", "pool", "set-quota", "data", "max_objects", "1000"},
	}
	got := make([][]string, 0, len(command.followups))
	for _, followup := range command.followups {
		got = append(got, followup.args)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("followups = %#v, want %#v", got, want)
	}
}

func TestPoolCreateBuildsCompressionOptionFollowups(t *testing.T) {
	command, err := build(Request{Action: "pool.create", ResourceKey: "pool"}, map[string]any{
		"name":                       "data",
		"pool_type":                  "replicated",
		"compression_mode":           "force",
		"compression_algorithm":      "zstd",
		"compression_min_blob_size":  "4096",
		"compression_max_blob_size":  "1048576",
		"compression_required_ratio": "0.875",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"osd", "pool", "set", "data", "compression_mode", "force"},
		{"osd", "pool", "set", "data", "compression_algorithm", "zstd"},
		{"osd", "pool", "set", "data", "compression_min_blob_size", "4096"},
		{"osd", "pool", "set", "data", "compression_max_blob_size", "1048576"},
		{"osd", "pool", "set", "data", "compression_required_ratio", "0.875"},
	}
	got := make([][]string, 0, len(command.followups))
	for _, followup := range command.followups {
		got = append(got, followup.args)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("followups = %#v, want %#v", got, want)
	}
}

func TestPoolUpdateBuildsCompressionOptionCommand(t *testing.T) {
	command, err := build(Request{Action: "pool.update", ResourceKey: "pool/data"}, map[string]any{
		"field": "compression_required_ratio",
		"value": "0.875",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"osd", "pool", "set", "data", "compression_required_ratio", "0.875"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestPoolUpdateBuildsPGAndPGPCommands(t *testing.T) {
	command, err := build(Request{Action: "pool.update", ResourceKey: "pool/data"}, map[string]any{
		"field": "pg_num",
		"value": "64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"osd", "pool", "set", "data", "pg_num", "64"}; !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
	if len(command.followups) != 1 {
		t.Fatalf("followups = %d, want 1", len(command.followups))
	}
	if want := []string{"osd", "pool", "set", "data", "pgp_num", "64"}; !reflect.DeepEqual(command.followups[0].args, want) {
		t.Fatalf("followup args = %#v, want %#v", command.followups[0].args, want)
	}
}

func TestPoolCreateBuildsRBDConfigurationFollowups(t *testing.T) {
	command, err := build(Request{Action: "pool.create", ResourceKey: "pool"}, map[string]any{
		"name":      "data",
		"pool_type": "replicated",
		"configuration": map[string]any{
			"rbd_qos_bps_limit":  "1048576",
			"rbd_qos_iops_limit": "1000",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(command.followups) != 2 {
		t.Fatalf("followups = %d, want 2", len(command.followups))
	}
	for index, want := range [][]string{
		{"config", "pool", "set", "data", "rbd_qos_bps_limit", "1048576"},
		{"config", "pool", "set", "data", "rbd_qos_iops_limit", "1000"},
	} {
		if command.followups[index].binary != executor.BinaryRBD {
			t.Fatalf("followup %d binary = %q, want rbd", index, command.followups[index].binary)
		}
		if !reflect.DeepEqual(command.followups[index].args, want) {
			t.Fatalf("followup %d args = %#v, want %#v", index, command.followups[index].args, want)
		}
	}
}

func TestPoolUpdateBuildsRBDConfigurationCommand(t *testing.T) {
	command, err := build(Request{Action: "pool.update", ResourceKey: "pool/data"}, map[string]any{
		"operation": "rbd_configuration",
		"field":     "rbd_qos_write_iops_burst",
		"value":     "2000",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"config", "pool", "set", "data", "rbd_qos_write_iops_burst", "2000"}
	if command.binary != executor.BinaryRBD || !reflect.DeepEqual(command.args, want) {
		t.Fatalf("command = %q %#v, want rbd %#v", command.binary, command.args, want)
	}
	if wantCheck := []string{"config", "pool", "list", "data", "--format", "json"}; !reflect.DeepEqual(command.check, wantCheck) {
		t.Fatalf("check = %#v, want %#v", command.check, wantCheck)
	}
}

func TestPoolApplicationDisableIncludesConfirmation(t *testing.T) {
	command, err := build(Request{Action: "pool.update", ResourceKey: "pool/data"}, map[string]any{
		"operation":   "application",
		"action":      "disable",
		"application": "rbd",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"osd", "pool", "application", "disable", "data", "rbd", "--yes-i-really-mean-it"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestHealthMuteOptions(t *testing.T) {
	for _, tt := range []struct {
		name   string
		params map[string]any
		want   []string
	}{
		{"default", nil, []string{"health", "mute", "OSD_DOWN"}},
		{"duration", map[string]any{"ttl": "30m"}, []string{"health", "mute", "OSD_DOWN", "30m"}},
		{"sticky", map[string]any{"sticky": true}, []string{"health", "mute", "OSD_DOWN", "--sticky"}},
		{"both", map[string]any{"ttl": "1h", "sticky": true}, []string{"health", "mute", "OSD_DOWN", "1h", "--sticky"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := build(Request{Action: "health.mute", ResourceKey: "health/mute/OSD_DOWN"}, tt.params)
			if err != nil || !reflect.DeepEqual(cmd.args, tt.want) {
				t.Fatalf("command = %v, %v; want %v", cmd.args, err, tt.want)
			}
		})
	}
	for _, params := range []map[string]any{{"ttl": "-1h"}, {"ttl": "0"}, {"ttl": "--sticky"}, {"ttl": "1h;whoami"}, {"ttl": 12}, {"sticky": "true"}} {
		if _, err := build(Request{Action: "health.mute", ResourceKey: "health/mute/OSD_DOWN"}, params); err == nil {
			t.Fatalf("accepted invalid options %v", params)
		}
	}
}

func TestSnapshotScheduleUsesNamedFilesystem(t *testing.T) {
	cmd, err := build(Request{Action: "snapshot_schedule.create", ResourceKey: "filesystem/cephfs-data/snapshot-schedule"}, map[string]any{"path": "/data", "schedule": "1h"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"fs", "snap-schedule", "add", "/data", "1h", "--fs", "cephfs-data"}
	if !reflect.DeepEqual(cmd.args, want) {
		t.Fatalf("args = %v; want %v", cmd.args, want)
	}
	wantCheck := []string{"fs", "snap-schedule", "status", "/data", "--fs", "cephfs-data", "--format", "json"}
	if !reflect.DeepEqual(cmd.check, wantCheck) {
		t.Fatalf("check = %v; want %v", cmd.check, wantCheck)
	}
}

func TestSnapshotScheduleSubvolumeScope(t *testing.T) {
	parameters := map[string]any{"path": "/", "schedule": "2h", "start": "2026-09-14T00:00:00", "subvol": "project-a", "group": "team-a"}
	request := Request{Action: "snapshot_schedule.create", ResourceKey: "filesystem/data/snapshot-schedule"}
	cmd, err := build(request, parameters)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"fs", "snap-schedule", "add", "/", "2h", "--fs", "data", "--start=2026-09-14T00:00:00", "--subvol=project-a", "--group=team-a"}) {
		t.Fatalf("args=%v", cmd.args)
	}
	if !reflect.DeepEqual(cmd.check, []string{"fs", "snap-schedule", "status", "/", "--fs", "data", "--format", "json", "--subvol=project-a", "--group=team-a"}) {
		t.Fatalf("scope lost in check: %v", cmd.check)
	}
	for _, key := range []string{"start", "subvol", "group"} {
		original := parameters[key]
		parameters[key] = "value\nnext"
		if _, err := build(request, parameters); err == nil {
			t.Fatalf("accepted newline in %s", key)
		}
		parameters[key] = original
	}
}

func TestSnapshotScheduleActivationTargetsExactSchedule(t *testing.T) {
	for _, verb := range []string{"activate", "deactivate", "remove"} {
		cmd, err := build(Request{Action: "snapshot_schedule.action", ResourceKey: "filesystem/data/snapshot-schedule"}, map[string]any{"path": "/project", "schedule": "1h", "start": "2026-09-14T00:00:00", "action": verb})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"fs", "snap-schedule", verb, "/project", "--repeat=1h", "--fs=data", "--start=2026-09-14T00:00:00"}
		if !reflect.DeepEqual(cmd.args, want) {
			t.Fatalf("args=%v", cmd.args)
		}
	}
}

func TestSnapshotRetentionCommands(t *testing.T) {
	for _, verb := range []string{"add", "remove"} {
		cmd, err := build(Request{Action: "snapshot_schedule.retention", ResourceKey: "filesystem/data/snapshot-schedule"}, map[string]any{"path": "/project", "retention": "24h7d", "action": verb})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"fs", "snap-schedule", "retention", verb, "/project", "24h7d", "--fs=data"}) {
			t.Fatalf("args=%v", cmd.args)
		}
	}
}

func TestSnapshotRetentionRejectsAmbiguousInput(t *testing.T) {
	for _, value := range []string{"1h2h", "1d1d", "24h garbage", "0d", "-1d", "1H", "1d\n2w"} {
		if _, err := build(Request{Action: "snapshot_schedule.retention", ResourceKey: "filesystem/data/snapshot-schedule"}, map[string]any{"path": "/", "retention": value, "action": "add"}); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}

func TestRBDSnapshotActionsPreserveNamespace(t *testing.T) {
	image := "pool-a/team-a/image-a"
	key := "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte(image)) + "/snapshot/snap-a/action"
	for _, verb := range []string{"protect", "unprotect", "rollback", "clone"} {
		cmd, err := build(Request{Action: "rbd_snapshot.action", ResourceKey: key}, map[string]any{"action": verb, "destination": "pool-b/team-b/clone-a"})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"snap", verb, image + "@snap-a"}
		check := []string{"snap", "ls", image, "--format", "json"}
		if verb == "clone" {
			want = []string{"clone", image + "@snap-a", "pool-b/team-b/clone-a"}
			check = []string{"info", "pool-b/team-b/clone-a", "--format", "json"}
		}
		if cmd.binary != executor.BinaryRBD || !reflect.DeepEqual(cmd.args, want) || !reflect.DeepEqual(cmd.check, check) {
			t.Fatalf("command=%+v", cmd)
		}
	}
	if _, err := build(Request{Action: "rbd_snapshot.action", ResourceKey: key}, map[string]any{"action": "clone"}); err == nil {
		t.Fatal("clone without destination accepted")
	}
}

func TestRBDImageActionsCheckDestinationAndNamespace(t *testing.T) {
	key := "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("pool/team/image")) + "/action"
	for _, verb := range []string{"copy", "deep-copy", "move-to-trash"} {
		cmd, err := build(Request{Action: "rbd_image.action", ResourceKey: key}, map[string]any{"action": verb, "destination": "dest/ns/image"})
		if err != nil {
			t.Fatal(err)
		}
		check := []string{"info", "dest/ns/image", "--format", "json"}
		if verb == "move-to-trash" {
			check = []string{"trash", "ls", "--pool", "pool", "--namespace", "team", "--format", "json"}
		}
		if !reflect.DeepEqual(cmd.check, check) {
			t.Fatalf("verb=%s check=%v", verb, cmd.check)
		}
	}
}

func TestRBDCapacityUsesExplicitBytes(t *testing.T) {
	image := "pool/ns/image"
	for _, action := range []string{"rbd_image.create", "rbd_image.update"} {
		key := "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte(image))
		cmd, err := build(Request{Action: action, ResourceKey: key}, map[string]any{"image_spec": image, "size": float64(1073741824)})
		if err != nil {
			t.Fatal(err)
		}
		verb := "create"
		if action == "rbd_image.update" {
			verb = "resize"
		}
		if !reflect.DeepEqual(cmd.args, []string{verb, image, "--size", "1073741824B"}) {
			t.Fatalf("capacity unit lost: %v", cmd.args)
		}
	}
}

func TestRBDGroupMemberCommandScope(t *testing.T) {
	for _, verb := range []string{"add", "remove"} {
		cmd, err := build(Request{Action: "rbd_group.member", ResourceKey: "pool/team/group"}, map[string]any{"group_spec": "pool/team/group", "image": "pool/team/image", "action": verb})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"group", "image", verb, "pool/team/group", "pool/team/image"}) || !reflect.DeepEqual(cmd.check, []string{"group", "image", "list", "pool/team/group", "--format", "json"}) {
			t.Fatalf("wrong group scope: %+v", cmd)
		}
	}
	for _, parameters := range []map[string]any{
		{"group_spec": "pool/team/group", "action": "add"},
		{"image": "pool/team/image", "action": "remove"},
		{"group_spec": "pool/team/group", "image": "pool/team/image", "action": "purge"},
	} {
		if _, err := build(Request{Action: "rbd_group.member"}, parameters); err == nil {
			t.Fatalf("accepted %v", parameters)
		}
	}
}

func TestRBDGroupSnapshotUsesQualifiedGroup(t *testing.T) {
	request := Request{Action: "rbd_group.snapshot", ResourceKey: "pool/team/group"}
	cmd, err := build(request, map[string]any{"group_spec": "pool/team/group", "name": "snapshot", "action": "create"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"group", "snap", "create", "pool/team/group@snapshot"}) || !reflect.DeepEqual(cmd.check, []string{"group", "snap", "list", "pool/team/group", "--format", "json"}) {
		t.Fatalf("wrong snapshot scope: %+v", cmd)
	}
	for _, parameters := range []map[string]any{{"group_spec": "pool/team/group"}, {"name": "snapshot"}} {
		if _, err := build(request, parameters); err == nil {
			t.Fatalf("accepted incomplete target: %v", parameters)
		}
	}
}

func TestRBDGroupSnapshotDestructiveActions(t *testing.T) {
	for _, verb := range []string{"remove", "rollback"} {
		cmd, err := build(Request{Action: "rbd_group.snapshot", ResourceKey: "pool/ns/group"}, map[string]any{"group_spec": "pool/ns/group", "name": "snapshot", "action": verb})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"group", "snap", verb, "pool/ns/group@snapshot"}) {
			t.Fatalf("wrong target: %v", cmd.args)
		}
	}
	for _, verb := range []string{"", "purge", "--help"} {
		if _, err := build(Request{Action: "rbd_group.snapshot"}, map[string]any{"group_spec": "pool/ns/group", "name": "snapshot", "action": verb}); err == nil {
			t.Fatalf("accepted %q", verb)
		}
	}
}

func TestRBDImageMirroringCommands(t *testing.T) {
	spec := "pool/team/image"
	key := "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte(spec)) + "/action"
	for _, verb := range []string{"enable-journal", "enable-snapshot", "disable", "promote", "demote", "resync", "snapshot"} {
		cmd, err := build(Request{Action: "rbd_image.action", ResourceKey: key}, map[string]any{"action": "mirror-" + verb})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"mirror", "image", verb, spec}
		if verb == "enable-journal" {
			want = []string{"mirror", "image", "enable", spec, "journal"}
		}
		if verb == "enable-snapshot" {
			want = []string{"mirror", "image", "enable", spec, "snapshot"}
		}
		check := []string{"mirror", "image", "status", spec, "--format", "json"}
		if verb == "disable" {
			check = []string{"info", spec, "--format", "json"}
		}
		if cmd.binary != executor.BinaryRBD || !reflect.DeepEqual(cmd.args, want) || !reflect.DeepEqual(cmd.check, check) {
			t.Fatalf("verb=%s command=%+v", verb, cmd)
		}
	}
}

func TestRBDMirroringPeerCommands(t *testing.T) {
	for _, verb := range []string{"add", "remove"} {
		payload := map[string]any{"pool": "pool", "action": verb, "remote_cluster": "remote", "remote_client": "client.mirror", "direction": "rx-only", "uuid": "peer-id"}
		cmd, err := build(Request{Action: "rbd_mirroring.peer", ResourceKey: "pool"}, payload)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"mirror", "pool", "peer", "add", "pool", "--remote-cluster=remote", "--remote-client-name=client.mirror", "--direction=rx-only"}
		if verb == "remove" {
			want = []string{"mirror", "pool", "peer", "remove", "pool", "peer-id"}
		}
		if !reflect.DeepEqual(cmd.args, want) || !reflect.DeepEqual(cmd.check, []string{"mirror", "pool", "info", "pool", "--format", "json"}) {
			t.Fatalf("command=%+v", cmd)
		}
	}
	for _, payload := range []map[string]any{
		{"pool": "pool", "action": "remove"},
		{"pool": "pool", "action": "add", "remote_cluster": "remote", "remote_client": "client.mirror", "direction": "tx-only"},
		{"pool": "pool", "action": "add", "direction": "rx-tx"},
	} {
		if _, err := build(Request{Action: "rbd_mirroring.peer"}, payload); err == nil {
			t.Fatalf("invalid peer accepted: %v", payload)
		}
	}
}

func TestRBDPeerUpdateFields(t *testing.T) {
	for field, value := range map[string]string{"site-name": "remote-site", "client": "client.mirror", "mon-host": "[v2:10.0.0.1:3300,v1:10.0.0.1:6789]", "direction": "tx-only"} {
		cmd, err := build(Request{Action: "rbd_mirroring.peer"}, map[string]any{"pool": "pool", "action": "set", "uuid": "peer-id", "field": field, "value": value})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"mirror", "pool", "peer", "set", "pool", "peer-id", field, value}) {
			t.Fatalf("args=%v", cmd.args)
		}
	}
	for _, p := range []map[string]any{
		{"pool": "pool", "action": "set", "uuid": "peer-id", "field": "key-file", "value": "/etc/passwd"},
		{"pool": "pool", "action": "set", "uuid": "peer-id", "field": "direction", "value": "invalid"},
		{"pool": "pool", "action": "set", "field": "client", "value": "client.mirror"},
	} {
		if _, err := build(Request{Action: "rbd_mirroring.peer"}, p); err == nil {
			t.Fatalf("invalid update accepted: %v", p)
		}
	}
}

func TestRBDGroupRenamePreservesScope(t *testing.T) {
	for _, spec := range []string{"pool/group", "pool/ns/group"} {
		parts := strings.Split(spec, "/")
		for _, verb := range []string{"rename", "remove"} {
			cmd, err := build(Request{Action: "rbd_group.action"}, map[string]any{"group_spec": spec, "action": verb, "name": "new"})
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"group", verb, spec}
			if verb == "rename" {
				args = append(args, strings.Join(parts[:len(parts)-1], "/")+"/new")
			}
			check := []string{"group", "list", "--pool", "pool"}
			if len(parts) == 3 {
				check = append(check, "--namespace", "ns")
			}
			check = append(check, "--format", "json")
			if !reflect.DeepEqual(cmd.args, args) || !reflect.DeepEqual(cmd.check, check) {
				t.Fatalf("command=%+v", cmd)
			}
		}
	}
	for _, p := range []map[string]any{
		{"group_spec": "pool/group", "action": "rename", "name": "other/new"},
		{"group_spec": "pool//group", "action": "remove"},
		{"group_spec": "group", "action": "remove"},
		{"group_spec": "pool/group", "action": "rename"},
	} {
		if _, err := build(Request{Action: "rbd_group.action"}, p); err == nil {
			t.Fatalf("invalid group action: %v", p)
		}
	}
}

func TestRBDGroupSnapshotRename(t *testing.T) {
	p := map[string]any{"group_spec": "pool/ns/group", "name": "old", "action": "rename", "new_name": "new"}
	cmd, err := build(Request{Action: "rbd_group.snapshot"}, p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"group", "snap", "rename", "pool/ns/group@old", "new"}) || !reflect.DeepEqual(cmd.check, []string{"group", "snap", "list", "pool/ns/group", "--format", "json"}) {
		t.Fatalf("command=%+v", cmd)
	}
	for _, name := range []string{"", "other/new", "group@snap"} {
		p["new_name"] = name
		if _, err := build(Request{Action: "rbd_group.snapshot"}, p); err == nil {
			t.Fatalf("invalid destination accepted: %s", name)
		}
	}
}

func TestRBDImageRenameStaysInNamespace(t *testing.T) {
	spec := "pool/team/image"
	request := Request{Action: "rbd_image.action", ResourceKey: "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte(spec)) + "/action"}
	cmd, err := build(request, map[string]any{"action": "rename", "destination": "renamed"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"rename", spec, "pool/team/renamed"}) || !reflect.DeepEqual(cmd.check, []string{"info", "pool/team/renamed", "--format", "json"}) {
		t.Fatalf("command=%+v", cmd)
	}
	for _, destination := range []string{"", "other/image", "image@snap"} {
		if _, err := build(request, map[string]any{"action": "rename", "destination": destination}); err == nil {
			t.Fatalf("invalid name accepted: %s", destination)
		}
	}
}

func TestRBDTrashPurgeScope(t *testing.T) {
	for _, pool := range []string{"pool", "pool/team"} {
		cmd, err := build(Request{Action: "rbd_trash.purge"}, map[string]any{"pool": pool})
		if err != nil {
			t.Fatal(err)
		}
		flags := []string{"--pool", "pool"}
		if pool == "pool/team" {
			flags = append(flags, "--namespace", "team")
		}
		if !reflect.DeepEqual(cmd.args, append([]string{"trash", "purge"}, flags...)) || !reflect.DeepEqual(cmd.check, append(append([]string{"trash", "ls"}, flags...), "--format", "json")) {
			t.Fatalf("command=%+v", cmd)
		}
	}
	for _, pool := range []string{"pool/", "pool/team/image"} {
		if _, err := build(Request{Action: "rbd_trash.purge"}, map[string]any{"pool": pool}); err == nil {
			t.Fatalf("invalid scope accepted: %s", pool)
		}
	}
}

func TestRBDTrashMoveExpiration(t *testing.T) {
	req := Request{Action: "rbd_image.action", ResourceKey: "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("pool/ns/image")) + "/action"}
	cmd, err := build(req, map[string]any{"action": "move-to-trash", "expires_at": "2026-12-31T23:59:59+08:00"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"trash", "mv", "pool/ns/image", "--expires-at=2026-12-31T15:59:59Z"}) {
		t.Fatalf("args=%v", cmd.args)
	}
	for _, value := range []string{"tomorrow", "2026-12-31T23:59:59", "invalid; command"} {
		if _, err := build(req, map[string]any{"action": "move-to-trash", "expires_at": value}); err == nil {
			t.Fatalf("invalid expiry accepted: %s", value)
		}
	}
}

func TestRBDTrashPurgeCutoff(t *testing.T) {
	p := map[string]any{"pool": "pool/ns", "expired_before": "2026-12-31T23:59:59+08:00"}
	cmd, err := build(Request{Action: "rbd_trash.purge"}, p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"trash", "purge", "--pool", "pool", "--namespace", "ns", "--expired-before=2026-12-31T15:59:59Z"}) {
		t.Fatalf("args=%v", cmd.args)
	}
	p["expired_before"] = "invalid"
	if _, err := build(Request{Action: "rbd_trash.purge"}, p); err == nil {
		t.Fatal("invalid cutoff accepted")
	}
}

func TestRBDImageFeatureActions(t *testing.T) {
	key := "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("pool/ns/image")) + "/action"
	for _, verb := range []string{"enable", "disable"} {
		cmd, err := build(Request{Action: "rbd_image.action", ResourceKey: key}, map[string]any{"action": "feature-" + verb, "feature": "fast-diff"})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"feature", verb, "pool/ns/image", "fast-diff"}) {
			t.Fatalf("args=%v", cmd.args)
		}
	}
	for _, feature := range []string{"deep-flatten", "unknown", ""} {
		if _, err := build(Request{Action: "rbd_image.action", ResourceKey: key}, map[string]any{"action": "feature-enable", "feature": feature}); err == nil {
			t.Fatalf("invalid feature accepted: %s", feature)
		}
	}
}

func TestRBDResizeRequiresExplicitShrinkFlag(t *testing.T) {
	req := Request{Action: "rbd_image.update", ResourceKey: "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("pool/ns/image"))}
	for _, allow := range []bool{false, true} {
		cmd, err := build(req, map[string]any{"size": float64(1024), "allow_shrink": allow})
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"resize", "pool/ns/image", "--size", "1024B"}
		if allow {
			expected = append(expected, "--allow-shrink")
		}
		if !reflect.DeepEqual(cmd.args, expected) {
			t.Fatalf("args=%v", cmd.args)
		}
	}
}

func TestRBDCreateLayoutOptions(t *testing.T) {
	p := map[string]any{"image_spec": "pool/ns/image", "size": float64(1048576), "data_pool": "ec-data", "stripe_unit": float64(4096), "stripe_count": float64(4)}
	cmd, err := build(Request{Action: "rbd_image.create"}, p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"create", "pool/ns/image", "--size", "1048576B", "--data-pool", "ec-data", "--stripe-unit", "4096B", "--stripe-count", "4"}) {
		t.Fatalf("args=%v", cmd.args)
	}
	p["stripe_count"] = float64(0)
	if _, err := build(Request{Action: "rbd_image.create"}, p); err == nil {
		t.Fatal("zero stripe count accepted")
	}
}

func TestRBDObjectSizeBounds(t *testing.T) {
	for _, size := range []int{4096, 4194304, 33554432} {
		cmd, err := build(Request{Action: "rbd_image.create"}, map[string]any{"image_spec": "pool/image", "size": float64(1048576), "object_size": float64(size)})
		if err != nil {
			t.Fatal(err)
		}
		if cmd.args[len(cmd.args)-2] != "--object-size" {
			t.Fatalf("args=%v", cmd.args)
		}
	}
	for _, size := range []int{0, 2048, 5000, 67108864} {
		if _, err := build(Request{Action: "rbd_image.create"}, map[string]any{"image_spec": "pool/image", "size": float64(1048576), "object_size": float64(size)}); err == nil {
			t.Fatalf("invalid object size accepted: %d", size)
		}
	}
}

func TestRBDImageSnapshotPurge(t *testing.T) {
	spec := "pool/ns/image"
	cmd, err := build(Request{Action: "rbd_image.action", ResourceKey: "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte(spec)) + "/action"}, map[string]any{"action": "snapshot-purge"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"snap", "purge", spec}) || !reflect.DeepEqual(cmd.check, []string{"snap", "ls", spec, "--format", "json"}) {
		t.Fatalf("command=%+v", cmd)
	}
}

func TestRBDDeleteAndUpdatePreserveNamespace(t *testing.T) {
	key := "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("pool/ns/image"))
	cmd, err := build(Request{Action: "rbd_image.delete", ResourceKey: key}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.check, []string{"ls", "--pool", "pool", "--namespace", "ns", "--format", "json"}) {
		t.Fatalf("check=%v", cmd.check)
	}
	cmd, err = build(Request{Action: "rbd_image.update", ResourceKey: key}, map[string]any{"name": "new"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"rename", "pool/ns/image", "pool/ns/new"}) {
		t.Fatalf("args=%v", cmd.args)
	}
	for _, spec := range []string{"image", "pool//image", "pool/ns/image/extra", "pool/image@snap"} {
		if _, err := decodeImageSpec(base64.RawURLEncoding.EncodeToString([]byte(spec))); err == nil {
			t.Fatalf("invalid spec accepted: %s", spec)
		}
	}
}

func TestRBDImageConfigurationActions(t *testing.T) {
	req := Request{Action: "rbd_image.action", ResourceKey: "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("pool/ns/image")) + "/action"}
	for _, verb := range []string{"set", "remove"} {
		cmd, err := build(req, map[string]any{"action": "config-" + verb, "config_name": "rbd_qos_iops_limit", "config_value": "1000"})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"config", "image", verb, "pool/ns/image", "rbd_qos_iops_limit"}
		if verb == "set" {
			want = append(want, "1000")
		}
		if !reflect.DeepEqual(cmd.args, want) {
			t.Fatalf("args=%v", cmd.args)
		}
	}
	if _, err := build(req, map[string]any{"action": "config-set", "config_name": "rbd_qos_iops_limit"}); err == nil {
		t.Fatal("missing config value accepted")
	}
}

func TestRBDDestinationsRequireCompleteImagePaths(t *testing.T) {
	key := "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("pool/ns/image"))
	for _, destination := range []string{"image", "pool//image", "pool/ns/image/extra", "pool/image@snapshot"} {
		for _, verb := range []string{"copy", "deep-copy"} {
			if _, err := build(Request{Action: "rbd_image.action", ResourceKey: key + "/action"}, map[string]any{"action": verb, "destination": destination}); err == nil {
				t.Fatalf("%s accepted %s", verb, destination)
			}
		}
		if _, err := build(Request{Action: "rbd_snapshot.action", ResourceKey: key + "/snapshot/snap/action"}, map[string]any{"action": "clone", "destination": destination}); err == nil {
			t.Fatalf("clone accepted %s", destination)
		}
	}
}

func TestRGWUserSuspensionUsesNativeSubcommand(t *testing.T) {
	for _, suspended := range []bool{true, false} {
		for _, withEdit := range []bool{true, false} {
			p := map[string]any{"suspended": suspended}
			if withEdit {
				p["display_name"] = "Updated User"
			}
			cmd, err := build(Request{Action: "rgw_user.update", ResourceKey: "rgw/user/user-a"}, p)
			if err != nil {
				t.Fatal(err)
			}
			if withEdit {
				if len(cmd.followups) != 1 {
					t.Fatalf("missing state operation: %+v", cmd)
				}
				cmd = cmd.followups[0]
			}
			verb := "enable"
			if suspended {
				verb = "suspend"
			}
			if !reflect.DeepEqual(cmd.args, []string{"user", verb, "--uid", "user-a", "--format", "json"}) {
				t.Fatalf("args=%v", cmd.args)
			}
		}
	}
}

func TestRGWUserEmailCanBeCleared(t *testing.T) {
	cmd, err := build(Request{Action: "rgw_user.update", ResourceKey: "rgw/user/user-a"}, map[string]any{"email": ""})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"user", "modify", "--uid", "user-a", "--email=", "--format", "json"}) {
		t.Fatalf("args=%v", cmd.args)
	}
	if _, err := build(Request{Action: "rgw_user.update", ResourceKey: "rgw/user/user-a"}, map[string]any{"email": "a\nb"}); err == nil {
		t.Fatal("multiline email accepted")
	}
}

func TestRGWCreateRequiresDisplayNameAndSupportsBucketLimit(t *testing.T) {
	if _, err := build(Request{Action: "rgw_user.create"}, map[string]any{"uid": "user"}); err == nil {
		t.Fatal("missing display name accepted")
	}
	for _, limit := range []float64{-1, 0, 100} {
		cmd, err := build(Request{Action: "rgw_user.create"}, map[string]any{"uid": "user", "display_name": "User", "max_buckets": limit})
		if err != nil {
			t.Fatal(err)
		}
		if cmd.args[4] != "--max-buckets" {
			t.Fatalf("args=%v", cmd.args)
		}
	}
	if _, err := build(Request{Action: "rgw_user.create"}, map[string]any{"uid": "user", "display_name": "User", "max_buckets": float64(-2)}); err == nil {
		t.Fatal("invalid limit accepted")
	}
}

func TestRGWRoleDeleteCommand(t *testing.T) {
	cmd, err := build(Request{Action: "rgw_role.delete", ResourceKey: "rgw/role/reader"}, map[string]any{"name": "reader"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"role", "delete", "--role-name", "reader", "--format", "json"}) {
		t.Fatalf("unexpected args: %#v", cmd.args)
	}
	if !reflect.DeepEqual(cmd.check, []string{"role", "list", "--format", "json"}) {
		t.Fatalf("unexpected check: %#v", cmd.check)
	}
	if _, err := build(Request{Action: "rgw_role.delete"}, map[string]any{}); err == nil {
		t.Fatal("missing name accepted")
	}
}

func TestRGWRoleUpdateValidation(t *testing.T) {
	params := map[string]any{"name": "reader", "assume_role_policy": `{"Statement":[]}`, "max_session_duration": float64(3600)}
	cmd, err := build(Request{Action: "rgw_role.update"}, params)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.args[0] != "role-trust-policy" || len(cmd.followups) != 1 || !reflect.DeepEqual(cmd.followups[0].args, []string{"role", "update", "--role-name", "reader", "--max-session-duration", "3600", "--format", "json"}) {
		t.Fatalf("unexpected chain: %#v", cmd)
	}
	for _, duration := range []any{3599, 43201, 3600.5} {
		params["max_session_duration"] = duration
		if _, err := build(Request{Action: "rgw_role.update"}, params); err == nil {
			t.Fatalf("accepted duration %v", duration)
		}
	}
	params["max_session_duration"] = float64(3600)
	params["assume_role_policy"] = `[]`
	if _, err := build(Request{Action: "rgw_role.update"}, params); err == nil {
		t.Fatal("accepted non-object policy")
	}
}

func TestRGWRolePolicyCommands(t *testing.T) {
	for _, action := range []string{"put", "delete"} {
		params := map[string]any{"name": "reader", "policy_name": "read-buckets", "action": action, "policy_document": `{"Statement":[]}`}
		cmd, err := build(Request{Action: "rgw_role.policy"}, params)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"role-policy", action, "--role-name", "reader", "--policy-name", "read-buckets"}
		if action == "put" {
			want = append(want, "--perm-policy-doc", `{"Statement":[]}`)
		}
		want = append(want, "--format", "json")
		if !reflect.DeepEqual(cmd.args, want) {
			t.Fatalf("args: %#v", cmd.args)
		}
		if !reflect.DeepEqual(cmd.check, []string{"role", "get", "--role-name", "reader", "--format", "json"}) {
			t.Fatalf("check: %#v", cmd.check)
		}
	}
	for _, policy := range []string{"", "null", "[]", "invalid"} {
		if _, err := build(Request{Action: "rgw_role.policy"}, map[string]any{"name": "reader", "policy_name": "p", "action": "put", "policy_document": policy}); err == nil {
			t.Fatalf("accepted policy %q", policy)
		}
	}
}

func TestRGWRoleUpdateIndependentFields(t *testing.T) {
	for _, tc := range []struct {
		params map[string]any
		noun   string
	}{
		{map[string]any{"name": "reader", "max_session_duration": float64(7200)}, "role"},
		{map[string]any{"name": "reader", "assume_role_policy": `{"Statement":[]}`}, "role-trust-policy"},
	} {
		cmd, err := build(Request{Action: "rgw_role.update"}, tc.params)
		if err != nil {
			t.Fatal(err)
		}
		if cmd.args[0] != tc.noun || len(cmd.followups) != 0 {
			t.Fatalf("unexpected extra write: %#v", cmd)
		}
		if !reflect.DeepEqual(cmd.check, []string{"role", "get", "--role-name", "reader", "--format", "json"}) {
			t.Fatalf("missing readback: %#v", cmd.check)
		}
	}
	if _, err := build(Request{Action: "rgw_role.update"}, map[string]any{"name": "reader"}); err == nil {
		t.Fatal("empty update accepted")
	}
}

func TestRGWAccountRoleMutationScope(t *testing.T) {
	for _, action := range []string{"rgw_role.create", "rgw_role.update", "rgw_role.delete", "rgw_role.policy"} {
		cmd, err := build(Request{Action: action}, map[string]any{"name": "reader", "account_id": "RGW123", "assume_role_policy": `{}`, "max_session_duration": float64(3600), "action": "delete", "policy_name": "p"})
		if err != nil {
			t.Fatal(err)
		}
		commands := append([]command{cmd}, cmd.followups...)
		for _, c := range commands {
			if !strings.Contains(strings.Join(c.args, " "), "--account-id RGW123") {
				t.Fatalf("unscoped write: %#v", c.args)
			}
			if len(c.check) > 2 && !strings.Contains(strings.Join(c.check, " "), "--account-id RGW123") {
				t.Fatalf("unscoped check: %#v", c.check)
			}
		}
	}
}

func TestRGWRoleCreateSessionOptions(t *testing.T) {
	params := map[string]any{"name": "reader", "assume_role_policy": `{}`, "description": "read only role", "max_session_duration": float64(7200)}
	cmd, err := build(Request{Action: "rgw_role.create"}, params)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"role", "create", "--role-name", "reader", "--max-session-duration", "7200", "--description", "read only role", "--assume-role-policy-doc", "{}", "--format", "json"}
	if !reflect.DeepEqual(cmd.args, want) {
		t.Fatalf("args: %#v", cmd.args)
	}
	for _, duration := range []float64{0, 3599, 43201, 3600.5} {
		params["max_session_duration"] = duration
		if _, err := build(Request{Action: "rgw_role.create"}, params); err == nil {
			t.Fatalf("accepted duration %v", duration)
		}
	}
}

func TestRGWAccountDeleteCommand(t *testing.T) {
	cmd, err := build(Request{Action: "rgw_account.delete"}, map[string]any{"account_id": "RGW123"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"account", "rm", "--account-id", "RGW123", "--format", "json"}) {
		t.Fatalf("args: %#v", cmd.args)
	}
	if !reflect.DeepEqual(cmd.check, []string{"account", "list", "--format", "json"}) {
		t.Fatalf("check: %#v", cmd.check)
	}
	if _, err := build(Request{Action: "rgw_account.delete"}, map[string]any{}); err == nil {
		t.Fatal("missing account accepted")
	}
}

func TestRGWAccountUpdateLimits(t *testing.T) {
	cmd, err := build(Request{Action: "rgw_account.update"}, map[string]any{"account_id": "RGW123", "email": "a@example.org", "max_users": float64(-1), "max_buckets": float64(0)})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"account", "modify", "--account-id", "RGW123", "--email=a@example.org", "--max-users", "-1", "--max-buckets", "0", "--format", "json"}
	if !reflect.DeepEqual(cmd.args, want) {
		t.Fatalf("args: %#v", cmd.args)
	}
	for _, params := range []map[string]any{{"account_id": "RGW123"}, {"account_id": "RGW123", "email": ""}, {"account_id": "RGW123", "max_users": float64(-2)}, {"account_id": "RGW123", "max_roles": float64(1.5)}} {
		if _, err := build(Request{Action: "rgw_account.update"}, params); err == nil {
			t.Fatalf("accepted invalid update: %#v", params)
		}
	}
}

func TestRGWAccountQuota(t *testing.T) {
	for _, scope := range []string{"account", "bucket"} {
		cmd, err := build(Request{Action: "rgw_account.quota"}, map[string]any{"account_id": "RGW123", "scope": scope, "enabled": true, "max_size": float64(1024), "max_objects": float64(-1)})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"quota", "enable", "--account-id", "RGW123", "--quota-scope", scope, "--max-size", "1024B", "--max-objects", "-1", "--format", "json"}
		if !reflect.DeepEqual(cmd.args, want) {
			t.Fatalf("args: %#v", cmd.args)
		}
	}
	for _, value := range []float64{-2, 1.5, 9007199254740992} {
		if _, err := build(Request{Action: "rgw_account.quota"}, map[string]any{"account_id": "RGW123", "scope": "account", "enabled": false, "max_size": value, "max_objects": float64(0)}); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
}

func TestRGWUserQuotaDisablePreservesLimits(t *testing.T) {
	cmd, err := build(Request{Action: "rgw_user.quota"}, map[string]any{"uid": "tenant$user", "scope": "user", "enabled": false, "max_size": float64(-1), "max_objects": float64(0)})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"quota", "set", "--uid", "tenant$user", "--quota-scope", "user", "--max-size", "-1024", "--max-objects", "0", "--format", "json"}) {
		t.Fatalf("args: %#v", cmd.args)
	}
	if len(cmd.followups) != 1 || !reflect.DeepEqual(cmd.followups[0].args, []string{"quota", "disable", "--uid", "tenant$user", "--quota-scope", "user", "--format", "json"}) {
		t.Fatalf("followups: %#v", cmd.followups)
	}
	if !reflect.DeepEqual(cmd.followups[0].check, []string{"user", "info", "--uid", "tenant$user", "--format", "json"}) {
		t.Fatalf("check: %#v", cmd.followups[0].check)
	}
}

func TestRGWUserCapabilities(t *testing.T) {
	for _, verb := range []string{"add", "rm"} {
		cmd, err := build(Request{Action: "rgw_user.caps"}, map[string]any{"uid": "tenant$user", "action": verb, "type": "users", "permission": "read,write"})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"caps", verb, "--uid", "tenant$user", "--caps", "users=read,write", "--format", "json"}) {
			t.Fatalf("args: %#v", cmd.args)
		}
	}
	if _, err := build(Request{Action: "rgw_user.caps"}, map[string]any{"uid": "u", "action": "add", "type": "users=*;buckets", "permission": "read"}); err == nil {
		t.Fatal("accepted multiple capability expression")
	}
}

func TestRGWUserRateLimitChain(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		params := map[string]any{"uid": "tenant$user", "enabled": enabled, "max_read_ops": float64(0), "max_write_ops": float64(10), "max_read_bytes": float64(1024), "max_write_bytes": float64(2048)}
		cmd, err := build(Request{Action: "rgw_user.ratelimit"}, params)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"ratelimit", "set", "--uid", "tenant$user", "--ratelimit-scope", "user", "--max-read-ops", "0", "--max-write-ops", "10", "--max-read-bytes", "1024", "--max-write-bytes", "2048", "--format", "json"}) {
			t.Fatalf("args: %#v", cmd.args)
		}
		verb := "disable"
		if enabled {
			verb = "enable"
		}
		if len(cmd.followups) != 1 || cmd.followups[0].args[1] != verb || cmd.followups[0].check[1] != "get" {
			t.Fatalf("followups: %#v", cmd.followups)
		}
		params["max_read_ops"] = float64(-1)
		if _, err := build(Request{Action: "rgw_user.ratelimit"}, params); err == nil {
			t.Fatal("negative accepted")
		}
	}
}

func TestBucketRateLimitTenantChain(t *testing.T) {
	id := base64.RawURLEncoding.EncodeToString([]byte("team\x00photos"))
	cmd, err := build(Request{Action: "rgw_bucket.ratelimit"}, map[string]any{"bucket_id": id, "enabled": true, "max_read_ops": float64(0), "max_write_ops": float64(0), "max_read_bytes": float64(0), "max_write_bytes": float64(0)})
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{cmd.args, cmd.followups[0].args, cmd.followups[0].check} {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "--bucket photos") || !strings.Contains(joined, "--tenant team") || !strings.Contains(joined, "--ratelimit-scope bucket") {
			t.Fatalf("lost scope: %#v", args)
		}
	}
}

func TestBucketQuotaScope(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		cmd, err := build(Request{Action: "rgw_bucket.quota"}, map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00photos")), "enabled": enabled, "max_size": float64(-1), "max_objects": float64(0)})
		if err != nil {
			t.Fatal(err)
		}
		args := strings.Join(cmd.args, " ")
		if !strings.Contains(args, "--tenant team") || !strings.Contains(args, "--max-size -1 --max-objects 0") {
			t.Fatalf("args: %#v", cmd.args)
		}
		if !enabled {
			if len(cmd.followups) != 1 || cmd.followups[0].args[1] != "disable" {
				t.Fatal("missing disable")
			}
			cmd = cmd.followups[0]
		}
		if !reflect.DeepEqual(cmd.check, []string{"bucket", "stats", "--bucket", "photos", "--tenant", "team", "--format", "json"}) {
			t.Fatalf("check: %#v", cmd.check)
		}
	}
}
