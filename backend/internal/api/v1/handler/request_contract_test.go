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
