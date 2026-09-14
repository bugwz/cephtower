package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"
)

func TestNamespaceSnapshotsAndTrashKeepDistinctIdentity(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_namespace":    []byte(`[{"name":"team"}]`),
		"collect.rbd_image_detail": []byte(`[{"image":"image","size":1024,"format":2},{"image":"image","snapshot":"snap","size":512,"format":2}]`),
		"collect.rbd_snapshot":     []byte(`[{"name":"snap","id":1,"size":1024}]`),
		"collect.rbd_trash":        []byte(`[{"name":"image","id":"abc123"}]`),
	}}}
	rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
	keys := map[string]bool{}
	counts := map[string]int{}
	for _, row := range rows {
		if row.Kind != "rbd_snapshot" && row.Kind != "rbd_trash" {
			continue
		}
		key := row.Kind + ":" + row.NaturalKey
		if keys[key] {
			t.Fatalf("namespace identities collide: %s", key)
		}
		keys[key] = true
		counts[row.Kind]++
		payload := row.Payload.(map[string]any)
		namespace := payload["namespace"].(string)
		if namespace != "" && namespace != "team" {
			t.Fatalf("unexpected namespace %s", namespace)
		}
		if row.Kind == "rbd_trash" {
			if payload["pool"] != "pool" || payload["image_id"] != "abc123" {
				t.Fatalf("trash identity lost: %+v", payload)
			}
		} else {
			expected := "pool/image"
			if namespace != "" {
				expected = "pool/team/image"
			}
			if payload["image_spec"] != base64.RawURLEncoding.EncodeToString([]byte(expected)) || payload["image_path"] != expected || payload["pool_name"] != "pool" {
				t.Fatalf("snapshot identity lost: %+v", payload)
			}
		}
	}
	if counts["rbd_snapshot"] != 2 || counts["rbd_trash"] != 2 {
		t.Fatalf("missing namespace resources: %v", counts)
	}
}

func TestNamespaceGroupsIncludeMembersAndSnapshots(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_namespace":       []byte(`[{"name":"team"}]`),
		"collect.rbd_group":           []byte(`["group"]`),
		"collect.rbd_group_info":      []byte(`{"group_name":"group","group_id":"group-id"}`),
		"collect.rbd_group_images":    []byte(`[{"pool":"pool","image":"image","namespace":"team","state":0}]`),
		"collect.rbd_group_snapshots": []byte(`[{"id":"snap-id","snapshot":"snapshot","state":"complete"}]`),
	}}}
	rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
	seen := map[string]bool{}
	for _, row := range rows {
		if row.Kind != "rbd_group" {
			continue
		}
		seen[row.NaturalKey] = true
		payload := row.Payload.(map[string]any)
		if payload["group_id"] != "group-id" || payload["group_spec"] != row.NaturalKey || payload["images"] == nil || payload["snapshots"] == nil {
			t.Fatalf("incomplete group: %+v", payload)
		}
	}
	if len(seen) != 2 || !seen["pool/group"] || !seen["pool/team/group"] {
		t.Fatalf("groups=%v", seen)
	}
}

func TestMirroringProducesIndependentPoolRowsWithStatus(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_mirroring":        []byte(`{"mode":"image","peers":[]}`),
		"collect.rbd_mirroring_status": []byte(`{"summary":{"health":"OK","states":{"replaying":1}},"daemons":[{"hostname":"mirror-host"}],"images":[{"name":"image"}]}`),
	}}}
	rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool-a"}, {PoolName: "pool-b"}}, fsDumpWire{}, time.Now())
	seen := map[string]bool{}
	for _, row := range rows {
		if row.Kind != "rbd_mirroring" {
			continue
		}
		payload := row.Payload.(map[string]any)
		if payload["pool"] != row.NaturalKey || payload["mode"] != "image" || payload["summary"] == nil || payload["daemons"] == nil || payload["images"] == nil {
			t.Fatalf("incomplete mirroring row: %+v", row)
		}
		seen[row.NaturalKey] = true
	}
	if len(seen) != 2 || !seen["pool-a"] || !seen["pool-b"] {
		t.Fatalf("pool identities lost: %v", seen)
	}
}

func TestMalformedMirroringInfoDoesNotBecomeEmptyPoolState(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `{"mode":12}`, `{"mode":"unexpected"}`} {
		trace := &collectionTrace{unavailable: map[string]struct{}{}}
		ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rbd_mirroring": []byte(raw)}}}
		if _, ok := provider.collectPoolMirroring(ctx, ClusterAccess{}, "pool"); ok {
			t.Fatalf("malformed state accepted: %s", raw)
		}
		if _, ok := trace.unavailable["rbd_mirroring"]; !ok {
			t.Fatal("malformed state would erase cached pool state")
		}
	}
}

func TestRBDImageInfoEnrichment(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_image_config": []byte(`[{"name":"rbd_qos_iops_limit","value":"1000","source":"image"}]`),
		"collect.rbd_image_status": []byte(`{"watchers":[{"address":"10.0.0.1:0/1","client":9007199254740993,"cookie":18446744073709551615}],"migration":{"state":"executed"},"persistent_cache":{"clean":true}}`),
		"collect.rbd_image_info":   []byte(`{"name":"image","features":["layering","exclusive-lock"],"parent":{"pool":"parent-pool","image":"parent","snapshot":"base"},"stripe_unit":4096}`),
	}}}
	payload := cephdomain.RBDImage{ImagePath: "pool/ns/image"}
	provider.enrichRBDImage(context.Background(), ClusterAccess{}, payload.ImagePath, &payload)
	watchers := objectList(payload.RuntimeStatus["watchers"])
	if len(watchers) != 1 || watchers[0]["client"] != "9007199254740993" || watchers[0]["cookie"] != "18446744073709551615" {
		t.Fatalf("watcher identifiers lost precision: %v", watchers)
	}
	if len(payload.Configuration) != 1 || payload.Configuration[0].Source != "image" || payload.Configuration[0].Value != "1000" {
		t.Fatalf("missing effective configuration: %+v", payload)
	}
	if payload.RuntimeStatus["watchers"] == nil || payload.RuntimeStatus["migration"] == nil || payload.RuntimeStatus["persistent_cache"] == nil {
		t.Fatalf("missing runtime status: %+v", payload)
	}
	if len(payload.Features) != 2 || payload.Features[0] != "layering" || payload.Parent["snapshot"] != "base" || payload.Details["stripe_unit"] == nil {
		t.Fatalf("missing details: %+v", payload)
	}
}

func TestRBDNullListsMarkUnavailable(t *testing.T) {
	for id, kind := range map[string]string{"collect.rbd_image_detail": "rbd_image", "collect.rbd_trash": "rbd_trash", "collect.rbd_group": "rbd_group"} {
		trace := &collectionTrace{unavailable: map[string]struct{}{}}
		ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{id: []byte(`null`)}}}
		p.collectStorageOptional(ctx, ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
		if _, ok := trace.unavailable[kind]; !ok {
			t.Fatalf("%s null list did not mark %s unavailable", id, kind)
		}
	}
}

func TestRGWUserStatsPreserveAccountScope(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rgw_user":        []byte(`["tenant$user"]`),
		"collect.rgw_user_detail": []byte(`{"user_id":"user","account_id":"RGW123"}`),
		"collect.rgw_user_stats":  []byte(`{"stats":{"size":1024,"num_objects":2},"last_stats_sync":"2026-09-14T00:00:00Z"}`),
	}}}
	rows := p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now())
	for _, row := range rows {
		if row.Kind == "rgw_user" {
			payload := row.Payload.(map[string]any)
			if payload["uid"] != "tenant$user" || payload["stats_scope"] != "account" || payload["storage_stats"] == nil {
				t.Fatalf("invalid stats: %v", payload)
			}
			return
		}
	}
	t.Fatal("user missing")
}

func TestRGWNullUserDetailsDoNotCreatePlaceholder(t *testing.T) {
	trace := &collectionTrace{unavailable: map[string]struct{}{}}
	ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rgw_user": []byte(`["user"]`), "collect.rgw_user_detail": []byte(`null`)}}}
	rows := p.collectRGWOptional(ctx, ClusterAccess{}, time.Now())
	for _, row := range rows {
		if row.Kind == "rgw_user" {
			t.Fatal("placeholder user emitted")
		}
	}
	if _, ok := trace.unavailable["rgw_user"]; !ok {
		t.Fatal("missing unavailable marker")
	}
}

func TestRGWAccountNativeFields(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rgw_account": []byte(`{"accounts":["RGW123"]}`), "collect.rgw_account_detail": []byte(`{"id":"RGW123","name":"Account One","max_users":100,"quota":{"enabled":true}}`), "collect.rgw_account_stats": []byte(`{"stats":{"size":12345,"num_objects":7},"last_synced":"2026-09-14T00:00:00Z"}`)}}}
	rows := p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now())
	for _, row := range rows {
		if row.Kind == "rgw_account" {
			data := row.Payload.(map[string]any)
			if data["account_id"] != "RGW123" || data["account_name"] != "Account One" || data["quota"] == nil || data["storage_stats"] == nil {
				t.Fatalf("account fields missing: %v", data)
			}
			return
		}
	}
	t.Fatal("account missing")
}

func TestRGWRoleListUsesNativeObjects(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rgw_role": []byte(`[{"RoleName":"role-a","Arn":"arn:role-a","Path":"/","AssumeRolePolicyDocument":"{}"}]`)}}}
	for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
		if row.Kind == "rgw_role" {
			if row.NaturalKey != "role-a" || row.Payload.(map[string]any)["Arn"] != "arn:role-a" {
				t.Fatalf("role=%+v", row)
			}
			return
		}
	}
	t.Fatal("native role omitted")
}

func TestAccountRoleIdentity(t *testing.T) {
	now := time.Now()
	roles := []any{map[string]any{"RoleName": "reader", "AccountId": "RGW123"}}
	rows := rgwRoleObservations(context.Background(), roles, "RGW123", now)
	if len(rows) != 1 || rows[0].NaturalKey != "RGW123/reader" {
		t.Fatalf("account identity: %#v", rows)
	}
	if rows := rgwRoleObservations(context.Background(), roles, "RGW456", now); len(rows) != 0 {
		t.Fatalf("accepted wrong account: %#v", rows)
	}
}

func TestRGWUserRateLimitShape(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rgw_user":           []byte(`["tenant$user"]`),
		"collect.rgw_user_detail":    []byte(`{"user_id":"tenant$user"}`),
		"collect.rgw_user_ratelimit": []byte(`{"user_ratelimit":{"enabled":true,"max_read_ops":100,"max_write_ops":20,"max_read_bytes":1024,"max_write_bytes":512}}`),
	}}}
	for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
		if row.Kind == "rgw_user" {
			limits, ok := row.Payload.(map[string]any)["rate_limit"].(map[string]any)
			if !ok || limits["enabled"] != true || fmt.Sprint(limits["max_read_ops"]) != "100" {
				t.Fatalf("limits: %#v", limits)
			}
			return
		}
	}
	t.Fatal("user missing")
}

func TestRGWBucketRateLimitShape(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rgw_bucket":           []byte(`["team/photos"]`),
		"collect.rgw_bucket_detail":    []byte(`{"bucket":"photos","tenant":"team"}`),
		"collect.rgw_bucket_ratelimit": []byte(`{"bucket_ratelimit":{"enabled":true,"max_read_ops":25}}`),
	}}}
	for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
		if row.Kind == "rgw_bucket" {
			if row.NaturalKey != opaquePair("team", "photos") || row.Name != "photos" {
				t.Fatalf("bucket identity: %#v", row)
			}
			limits, ok := row.Payload.(map[string]any)["rate_limit"].(map[string]any)
			if !ok || limits["enabled"] != true || fmt.Sprint(limits["max_read_ops"]) != "25" {
				t.Fatalf("limits: %#v", limits)
			}
			return
		}
	}
	t.Fatal("bucket missing")
}

func TestGlobalRGWRateLimits(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rgw_global_ratelimit": []byte(`{"user_ratelimit":{"enabled":true,"max_read_ops":123},"bucket_ratelimit":{"enabled":false},"anonymous_ratelimit":{"enabled":true}}`)}}}
	rows, err := p.collectStorage(context.Background(), ClusterAccess{})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Kind == "rgw_status" {
			limits := row.Payload.(cephdomain.RGWStatus).GlobalRateLimit
			if len(limits) != 3 || limits["user_ratelimit"].(map[string]any)["enabled"] != true {
				t.Fatalf("limits: %#v", limits)
			}
			return
		}
	}
	t.Fatal("status missing")
}
