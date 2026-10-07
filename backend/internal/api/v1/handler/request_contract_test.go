package handler

import (
	"encoding/base64"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIDLessServiceRequestContracts(t *testing.T) {
	for _, action := range []string{"service.create", "service.update"} {
		for _, kind := range []string{"loki", "promtail", "rbd-mirror", "cephfs-mirror"} {
			fields := map[string]any{"cluster_id": float64(1), "service_type": kind, "placement": map[string]any{"count": float64(2)}}
			if err := ValidateMutationRequest(action, fields); err != nil {
				t.Fatalf("%s %s: %v", action, kind, err)
			}
		}
	}
}

func TestIngressServiceContract(t *testing.T) {
	if err := ValidateMutationRequest("service.update", map[string]any{"cluster_id": float64(1), "service_type": "ingress", "virtual_ip": "192.0.2.10/24", "frontend_port": float64(443), "monitor_port": float64(9000)}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("service.update", map[string]any{"cluster_id": float64(1), "service_type": "ingress", "ssl": true, "ssl_cert": "certificate fixture", "ssl_key": "key fixture"}); err != nil {
		t.Fatal(err)
	}
	fields := map[string]any{"cluster_id": float64(1), "service_type": "ingress", "service_id": "rgw.a", "backend_service": "rgw.a", "virtual_ip": "192.0.2.10/24", "frontend_port": float64(8080), "monitor_port": float64(9000)}
	fields["virtual_interface_networks"] = []any{"192.0.2.0/24", "2001:db8::/64"}
	fields["ssl"], fields["ssl_cert"], fields["ssl_key"] = true, "certificate fixture", "key fixture"
	if err := ValidateMutationRequest("service.create", fields); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("service.update", fields); err == nil {
		t.Fatal("unsupported ingress field update accepted")
	}
}

func TestBucketConfigurationContractRequiresRawDocument(t *testing.T) {
	for _, kind := range []string{"policy", "cors", "lifecycle", "encryption", "tagging"} {
		if err := ValidateMutationRequest("rgw_bucket_policy.update", map[string]any{"cluster_id": float64(1), "bucket_id": "AGJ1Y2tldA", "kind": kind, "document": "raw document"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, fields := range []map[string]any{
		{"kind": "policy"}, {"document": "{}"}, {"kind": "cors", "document": map[string]any{}}, {"kind": "policy", "policy": map[string]any{}}, {"kind": "unknown", "document": "{}"},
	} {
		fields["cluster_id"], fields["bucket_id"] = float64(1), "AGJ1Y2tldA"
		if err := ValidateMutationRequest("rgw_bucket_policy.update", fields); err == nil {
			t.Fatalf("accepted obsolete or invalid shape: %v", fields)
		}
	}
}

func TestBucketConfigurationDeleteContract(t *testing.T) {
	for _, kind := range []string{"policy", "cors", "lifecycle", "encryption", "tagging", "replication"} {
		if err := ValidateMutationRequest("rgw_bucket_policy.delete", map[string]any{"cluster_id": float64(1), "bucket_id": "AGJ1Y2tldA", "kind": kind}); err != nil {
			t.Fatal(err)
		}
	}
	for _, fields := range []map[string]any{{"bucket_id": "AGJ1Y2tldA"}, {"kind": "cors"}, {"bucket_id": "AGJ1Y2tldA", "kind": "versioning"}, {"bucket_id": "AGJ1Y2tldA", "kind": "cors", "document": "x"}} {
		fields["cluster_id"] = float64(1)
		if err := ValidateMutationRequest("rgw_bucket_policy.delete", fields); err == nil {
			t.Fatalf("invalid deletion accepted: %v", fields)
		}
	}
}

func TestBucketCannedACLContract(t *testing.T) {
	for _, value := range []string{"private", "public-read", "public-read-write", "authenticated-read"} {
		if err := ValidateMutationRequest("rgw_bucket.acl", map[string]any{"cluster_id": float64(1), "bucket_id": "AGJ1Y2tldA", "acl": value}); err != nil {
			t.Fatal(err)
		}
	}
	for _, fields := range []map[string]any{
		{"acl": "private"}, {"bucket_id": "AGJ1Y2tldA"}, {"bucket_id": "AGJ1Y2tldA", "acl": "unknown"},
		{"bucket_id": "AGJ1Y2tldA", "acl": true}, {"bucket_id": "AGJ1Y2tldA", "acl": "private", "document": "<xml/>"},
	} {
		fields["cluster_id"] = float64(1)
		if err := ValidateMutationRequest("rgw_bucket.acl", fields); err == nil {
			t.Fatalf("invalid ACL request accepted: %#v", fields)
		}
	}
}

func TestMutationContractsRejectUnknownAndWrongType(t *testing.T) {
	if err := ValidateMutationRequest("host.create", map[string]any{"hostname": "node-1", "password": "secret"}); err == nil {
		t.Fatal("unknown field was accepted")
	}
	if err := ValidateMutationRequest("manager_module.update", map[string]any{"enabled": "true"}); err == nil {
		t.Fatal("wrong field type was accepted")
	}
	if err := ValidateMutationRequest("service.create", map[string]any{"service_type": "mgr", "placement": map[string]any{"shell": "bad"}}); err == nil {
		t.Fatal("unknown nested field was accepted")
	}
	if err := ValidateMutationRequest("service.create", map[string]any{"cluster_id": float64(1), "service_type": "mgr", "placement": map[string]any{"count": float64(2)}}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("device.zap", map[string]any{"cluster_id": float64(1), "host": "node-1", "device": "/dev/sdb"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("device.zap", map[string]any{"cluster_id": float64(1), "device_id": "encoded"}); err == nil {
		t.Fatal("device zap without host and device was accepted")
	}
	if err := ValidateMutationRequest("pool.create", map[string]any{
		"cluster_id":                 float64(1),
		"name":                       "data",
		"pool_type":                  "replicated",
		"pg_num":                     float64(32),
		"pg_autoscale_mode":          "on",
		"size":                       float64(3),
		"applications":               []any{"rbd"},
		"erasure_code_profile":       "default",
		"crush_rule":                 "replicated_rule",
		"flags":                      []any{"allow_ec_overwrites"},
		"compression_mode":           "force",
		"compression_algorithm":      "zstd",
		"compression_min_blob_size":  float64(4096),
		"compression_max_blob_size":  float64(1048576),
		"compression_required_ratio": float64(0.875),
		"quota_max_bytes":            float64(0),
		"quota_unit":                 "GiB",
		"quota_max_objects":          float64(0),
		"rbd_mirroring":              "pool",
		"configuration": map[string]any{
			"rbd_qos_bps_limit":        float64(0),
			"rbd_qos_write_iops_burst": float64(2000),
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("pool.create", map[string]any{
		"cluster_id": float64(1),
		"name":       "ec-data",
		"pool_type":  "erasure",
		"flags":      []any{"unsupported"},
	}); err == nil {
		t.Fatal("unsupported pool flag was accepted")
	}
	if err := ValidateMutationRequest("pool.update", map[string]any{
		"cluster_id": float64(1),
		"pool":       "data",
		"field":      "crush_rule",
		"value":      "replicated_rule",
	}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("pool.update", map[string]any{
		"cluster_id":    float64(1),
		"pool":          "data",
		"operation":     "rbd_mirroring",
		"rbd_mirroring": "disabled",
	}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("erasure_code_profile.create", map[string]any{
		"cluster_id": float64(1), "name": "ec-isa", "plugin": "isa", "k": float64(7),
		"m": float64(3), "technique": "reed_sol_van", "crush-failure-domain": "host",
		"crush-num-failure-domains": float64(3), "crush-osds-per-failure-domain": float64(2),
		"crush-root": "default", "crush-device-class": "ssd",
		"directory": "/usr/lib64/ceph/erasure-code",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestServiceHostPatternUnion(t *testing.T) {
	for _, value := range []any{"node-*", map[string]any{"pattern": "node-[0-9]+", "pattern_type": "regex"}, map[string]any{"pattern": "node-*"}} {
		if err := ValidateMutationRequest("service.create", map[string]any{"cluster_id": float64(1), "service_type": "rgw", "placement": map[string]any{"host_pattern": value}}); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []any{nil, true, []any{}, map[string]any{}, map[string]any{"pattern": 3}, map[string]any{"pattern": "a", "pattern_type": "shell"}, map[string]any{"pattern": "a", "unknown": true}} {
		if err := ValidateMutationRequest("service.update", map[string]any{"cluster_id": float64(1), "service_type": "rgw", "placement": map[string]any{"host_pattern": value}}); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
}

func TestCephFSCreateQuotaStringContract(t *testing.T) {
	for _, action := range []string{"subvolume.create", "subvolume_group.create"} {
		body := map[string]any{"cluster_id": float64(1), "fs": "cephfs", "name": "test", "pool": "data", "size": "9007199254740993"}
		if action == "subvolume.create" {
			body["group"] = "_nogroup"
		}
		if err := ValidateMutationRequest(action, body); err != nil {
			t.Fatal(err)
		}
		body["size"] = float64(1024)
		if err := ValidateMutationRequest(action, body); err == nil {
			t.Fatalf("%s accepted numeric quota", action)
		}
	}
}

func TestHostMutationContractsAcceptManagementFields(t *testing.T) {
	if err := ValidateMutationRequest("host.create", map[string]any{
		"cluster_id": float64(1), "hostname": "node-1", "address": "192.0.2.10",
		"labels": []any{"_admin", "osd"}, "maintenance": true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("host.update", map[string]any{
		"cluster_id": float64(1), "host": "node-1",
		"labels_add": []any{"osd"}, "labels_remove": []any{"mon"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("host.action", map[string]any{
		"cluster_id": float64(1), "host": "node-1", "action": "maintenance_enter", "force": true,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMutationContractsCoverAllRegisteredActions(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "MutateResource" {
				return true
			}
			if len(call.Args) < 2 {
				t.Fatalf("%s: malformed mutation registration", path)
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				t.Fatalf("%s: dynamic action needs explicit contract coverage", path)
			}
			action, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			contract, ok := MutationRequestContract(action)
			if !ok || contract.Fields == nil {
				t.Errorf("%s: action %s has no request contract", path, action)
			}
			checked++
			return true
		})
	}
	if checked == 0 {
		t.Fatal("no mutation handlers inspected")
	}
}

func TestRGWAccountCreateLimitContract(t *testing.T) {
	for _, field := range []string{"max_users", "max_roles", "max_groups", "max_buckets", "max_access_keys"} {
		for _, value := range []any{float64(-1), float64(0), float64(2147483647)} {
			if err := ValidateMutationRequest("rgw_account.create", map[string]any{"cluster_id": float64(1), "account_id": "RGW123", field: value}); err != nil {
				t.Fatalf("rejected %s = %v: %v", field, value, err)
			}
		}
		for _, value := range []any{"0", false, float64(0.5)} {
			if err := ValidateMutationRequest("rgw_account.create", map[string]any{"cluster_id": float64(1), "account_id": "RGW123", field: value}); err == nil {
				t.Fatalf("accepted invalid %s = %v", field, value)
			}
		}
	}
}

func TestRGWUserMutationIdentity(t *testing.T) {
	for _, action := range []string{"rgw_user.update", "rgw_user.delete", "rgw_key.create", "rgw_key.update", "rgw_key.delete"} {
		body := func(uid any) map[string]any {
			value := map[string]any{"cluster_id": float64(1), "uid": uid}
			if action == "rgw_key.create" || action == "rgw_key.update" || action == "rgw_key.delete" {
				value["access_key"] = "test-access"
				value["confirm_owner"] = uid
			}
			if action == "rgw_key.create" || action == "rgw_key.update" {
				value["secret_key"] = "test-secret"
			}
			return value
		}
		for _, uid := range []string{"user", "tenant$user", "tenant$namespace$user", "$namespace$user"} {
			if err := ValidateMutationRequest(action, body(uid)); err != nil {
				t.Fatalf("%s rejected %q: %v", action, uid, err)
			}
		}
		for _, uid := range []any{nil, false, "", " user", "user ", "tenant/user", "user/", "-user", "a\nb", "a\x00b"} {
			if err := ValidateMutationRequest(action, body(uid)); err == nil {
				t.Fatalf("%s accepted unsafe identity %#v", action, uid)
			}
		}
	}
}

func TestRGWSubuserContract(t *testing.T) {
	valid := func() map[string]any {
		return map[string]any{"cluster_id": float64(1), "uid": "tenant$user", "action": "modify", "subuser": "swift", "confirm_subuser": "tenant$user:swift", "subuser_permission": "read"}
	}
	if err := ValidateMutationRequest("rgw_user.subuser", valid()); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"uid", "action", "subuser", "confirm_subuser"} {
		p := valid()
		delete(p, field)
		if err := ValidateMutationRequest("rgw_user.subuser", p); err == nil {
			t.Fatalf("accepted missing %s", field)
		}
	}
	for field, value := range map[string]any{"uid": "other/user", "action": "delete", "subuser_permission": "full-control", "generate_secret": true} {
		p := valid()
		p[field] = value
		if err := ValidateMutationRequest("rgw_user.subuser", p); err == nil {
			t.Fatalf("accepted invalid %s", field)
		}
	}
}

func TestHealthMuteContract(t *testing.T) {
	if err := ValidateMutationRequest("health.mute", map[string]any{"cluster_id": float64(1), "code": "OSD_DOWN", "ttl": "1h", "sticky": true}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("health.mute", map[string]any{"cluster_id": float64(1), "code": "OSD_DOWN", "sticky": "true"}); err == nil {
		t.Fatal("accepted string sticky")
	}
}

func TestCephUserRequestContract(t *testing.T) {
	for _, action := range []string{"ceph_user.create", "ceph_user.update"} {
		if err := ValidateMutationRequest(action, map[string]any{"cluster_id": float64(1), "entity": "client.backup", "caps": map[string]any{"mon": "allow r"}}); err != nil {
			t.Fatal(err)
		}
		if err := ValidateMutationRequest(action, map[string]any{"cluster_id": float64(1), "entity": "client.backup", "caps": map[string]any{"extra": "allow *"}}); err == nil {
			t.Fatal("unsupported capability accepted")
		}
	}
	if got := resourceLookupKey("ceph_user", "ceph-user/client.backup"); got != "client.backup" {
		t.Fatalf("key = %s", got)
	}
}

func TestConfigurationResourceIdentityPreservesMask(t *testing.T) {
	for _, tt := range []struct{ path, want string }{{"global\x00foo", "global:foo"}, {"mgr\x00mgr/dashboard/ssl", "mgr:mgr/dashboard/ssl"}, {"osd/host:node-a\x00foo", "osd/host:node-a:foo"}} {
		if got := resourceLookupKey("config_value", "configuration/value/"+base64.RawURLEncoding.EncodeToString([]byte(tt.path))); got != tt.want {
			t.Fatalf("key=%s want=%s", got, tt.want)
		}
	}
}

func TestRBDGroupActionRequestContract(t *testing.T) {
	if _, ok := MutationRequestContract("rbd_group.action"); !ok {
		t.Fatal("group action contract missing")
	}
	for _, verb := range []string{"rename", "remove"} {
		if err := ValidateMutationRequest("rbd_group.action", map[string]any{"cluster_id": float64(1), "group_spec": "pool/ns/group", "action": verb, "name": "new"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateMutationRequest("rbd_group.action", map[string]any{"cluster_id": float64(1), "group_spec": "pool/ns/group", "action": "unknown"}); err == nil {
		t.Fatal("unknown action accepted")
	}
}
