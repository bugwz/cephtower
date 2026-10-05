package ceph

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestParseRGWPoolReference(t *testing.T) {
	for _, tc := range []struct{ raw, pool, namespace string }{
		{"default.rgw.meta:users.uid", "default.rgw.meta", "users.uid"},
		{"data", "data", ""}, {`pool\:name:ns\:part`, `pool:name`, `ns:part`},
		{`pool\\name:ns\\part`, `pool\name`, `ns\part`},
		{"池:空间", "池", "空间"},
	} {
		pool, namespace, ok := parseRGWPoolReference(tc.raw)
		if !ok || pool != tc.pool || namespace != tc.namespace {
			t.Fatalf("incorrect reference %q", tc.raw)
		}
	}
	for _, raw := range []string{"", ":ns", "pool:", "pool:ns:extra", `pool\`, `pool\x`, "pool\x00", "pool\n", "\xff", strings.Repeat("x", 4097)} {
		if _, _, ok := parseRGWPoolReference(raw); ok {
			t.Fatalf("ambiguous reference %q accepted", raw)
		}
	}
}
func zonePoolsFixture() map[string]any {
	zone := map[string]any{"id": "z", "name": "zone", "system_key": map[string]any{"secret_key": "secret-fixture"}}
	for _, field := range strings.Fields("domain_root control_pool dedup_pool gc_pool lc_pool log_pool intent_log_pool usage_log_pool roles_pool reshard_pool user_keys_pool user_email_pool user_swift_pool user_uid_pool otp_pool notif_pool topics_pool account_pool group_pool bucket_logging_pool restore_pool") {
		zone[field] = ""
	}
	zone["domain_root"] = "meta:root"
	zone["user_uid_pool"] = "meta:users.uid"
	zone["placement_pools"] = []any{map[string]any{"key": "default-placement", "val": map[string]any{"index_pool": "index", "data_extra_pool": "extra", "storage_classes": map[string]any{
		"STANDARD": map[string]any{"data_pool": "data"}, "COLD": map[string]any{"data_pool": "data-cold:archive"}, "INHERITED": map[string]any{"compression_type": "zstd"},
	}}}}
	return zone
}
func TestAttachRGWZonePoolReferences(t *testing.T) {
	zone := zonePoolsFixture()
	attachRGWZonePoolReferences(zone)
	refs := zone["pool_references"].([]rgwPoolReference)
	if zone["pool_references_complete"] != true || len(refs) != 6 {
		t.Fatalf("incomplete: %#v", zone)
	}
	seen := map[string]string{}
	for _, ref := range refs {
		seen[ref.Field] = ref.Pool + ":" + ref.Namespace
	}
	if seen["domain_root"] != "meta:root" || seen["user_uid_pool"] != "meta:users.uid" || seen["placement_pools[default-placement].storage_classes[COLD].data_pool"] != "data-cold:archive" {
		t.Fatal("pool names or namespaces lost")
	}
	encoded, _ := json.Marshal(refs)
	if strings.Contains(string(encoded), "secret-fixture") {
		t.Fatal("system key entered references")
	}
	for _, change := range []func(map[string]any){
		func(z map[string]any) { delete(z, "control_pool") }, func(z map[string]any) { z["log_pool"] = true },
		func(z map[string]any) { z["user_uid_pool"] = "meta:ns:extra" }, func(z map[string]any) { delete(z, "placement_pools") },
		func(z map[string]any) { z["placement_pools"] = []any{nil} },
		func(z map[string]any) { z["placement_pools"] = []any{map[string]any{"key": "p", "val": nil}} },
		func(z map[string]any) { p := z["placement_pools"].([]any); z["placement_pools"] = append(p, p[0]) },
	} {
		z := zonePoolsFixture()
		change(z)
		attachRGWZonePoolReferences(z)
		if z["pool_references_complete"] != false || len(z["pool_reference_issues"].([]string)) == 0 {
			t.Fatal("invalid config marked complete")
		}
	}
	zone = zonePoolsFixture()
	zone["future_pool"] = "future:ns"
	attachRGWZonePoolReferences(zone)
	if len(zone["pool_references"].([]rgwPoolReference)) != 7 {
		t.Fatal("new top-level pool field omitted")
	}
}
func TestZonePoolReferencesCollected(t *testing.T) {
	raw, _ := json.Marshal(zonePoolsFixture())
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rgw_zone": []byte(`{"zones":["zone"],"default_info":"z"}`), "collect.rgw_zone_detail": raw,
	}}}
	found := false
	for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
		if row.Kind == "rgw_zone" {
			found = true
			data := row.Payload.(map[string]any)
			if data["pool_references_complete"] != true || len(data["pool_references"].([]rgwPoolReference)) != 6 {
				t.Fatal("derived references not attached")
			}
		}
	}
	if !found {
		t.Fatal("zone inventory missing")
	}
}
