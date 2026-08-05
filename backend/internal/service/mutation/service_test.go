package mutation

import (
	"encoding/base64"
	"reflect"
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
		{"host.create", "host", map[string]any{"hostname": "node1", "address": "192.0.2.10"}},
		{"host.update", "host/node1", map[string]any{"label": "storage", "action": "add"}},
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
		{"subvolume_group.create", "filesystem/cephfs/subvolume-group", map[string]any{"name": "group"}},
		{"subvolume_group.update", "filesystem/cephfs/subvolume-group/group", map[string]any{"size": "1024"}}, {"subvolume_group.delete", "filesystem/cephfs/subvolume-group/group", nil},
		{"subvolume.create", "filesystem/cephfs/subvolume", map[string]any{"name": "sub"}},
		{"subvolume.update", "filesystem/cephfs/subvolume/sub", map[string]any{"size": "2048"}}, {"subvolume.delete", "filesystem/cephfs/subvolume/sub", nil},
		{"cephfs_snapshot.create", "filesystem/cephfs/subvolume/sub/snapshot", map[string]any{"name": "snap"}},
		{"cephfs_snapshot.clone", "filesystem/cephfs/subvolume/sub/snapshot/snap/clone", map[string]any{"target": "clone"}},
		{"snapshot_schedule.create", "filesystem/cephfs/snapshot-schedule", map[string]any{"path": "/", "schedule": "1h"}},
		{"cephfs_authorization.create", "filesystem/cephfs/authorization", map[string]any{"client": "client.app", "path": "/", "access": "rw"}},
		{"cephfs_client.evict", "filesystem/cephfs/client/123", nil},
		{"cephfs_entry.quota", "filesystem/cephfs/entry/quota", map[string]any{"path": "/data", "max_bytes": "1024"}},
		{"rgw_user.create", "rgw/user", map[string]any{"uid": "user1", "display_name": "User One"}},
		{"rgw_user.update", "rgw/user/user1", map[string]any{"email": "user1@example.test"}}, {"rgw_user.delete", "rgw/user/user1", nil},
		{"rgw_account.create", "rgw/account", map[string]any{"account_id": "RGW00000000000000001", "account_name": "Account One"}},
		{"rgw_role.create", "rgw/role", map[string]any{"name": "reader", "path": "/"}},
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
		{"config_value.set", "configuration/value/global/osd_pool_default_size", map[string]any{"value": "3"}}, {"config_value.delete", "configuration/value/global/osd_pool_default_size", nil},
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
