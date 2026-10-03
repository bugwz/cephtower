package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNamespaceSnapshotsAndTrashKeepDistinctIdentity(t *testing.T) {
	base := malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_namespace":         []byte(`[{"name":"team"}]`),
		"collect.rbd_image_detail":      []byte(`[{"image":"image","size":1024,"format":2},{"image":"image","snapshot":"snap","size":512,"format":2}]`),
		"collect.rbd_image_info":        []byte(`{"name":"image","features":["fast-diff"]}`),
		"collect.rbd_image_usage":       []byte(`{"images":[{"name":"image","snapshot":"snap","used_size":256},{"name":"image","used_size":512}]}`),
		"collect.rbd_snapshot":          []byte(`[{"name":"snap","id":1,"size":1024,"protected":"true","timestamp":"Tue Sep 23 02:03:04 2026"}]`),
		"collect.rbd_snapshot_children": []byte(`[{"pool":"child-pool","pool_namespace":"apps","image":"child","id":"child-id","trash":true}]`),
		"collect.rbd_trash":             []byte(`[{"name":"image","id":"abc123","source":"USER","status":"protected until Sat Oct  3 12:00:00 2026"}]`),
	}}
	var calls []executor.CommandSpec
	provider := NativeProvider{Executor: recordingExecutor{base: base, calls: &calls}}
	rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
	keys := map[string]bool{}
	counts := map[string]int{}
	for _, row := range rows {
		if row.Kind == "rbd_image" {
			counts[row.Kind]++
			continue
		}
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
			if payload["trash_source"] != "USER" || payload["deferment_status"] != "protected until Sat Oct  3 12:00:00 2026" || payload["status"] != nil || payload["source"] != nil {
				t.Fatalf("native trash state conflicts with inventory metadata: %+v", payload)
			}
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
			if payload["protected"] != true || payload["is_protected"] != true || payload["used_bytes"] != uint64(256) || payload["disk_usage"] != uint64(256) || payload["timestamp"] != "Tue Sep 23 02:03:04 2026" {
				t.Fatalf("snapshot details lost: %+v", payload)
			}
			children := objectList(payload["children"])
			if len(children) != 1 || children[0]["pool"] != "child-pool" || children[0]["pool_namespace"] != "apps" || children[0]["image"] != "child" || children[0]["id"] != "child-id" || children[0]["trash"] != true {
				t.Fatalf("snapshot children lost: %+v", payload)
			}
		}
	}
	if counts["rbd_image"] != 2 || counts["rbd_snapshot"] != 2 || counts["rbd_trash"] != 2 {
		t.Fatalf("missing namespace resources: %v", counts)
	}
	imageLists := map[string]int{}
	usageCalls := map[string]int{}
	childCalls := map[string]int{}
	for _, call := range calls {
		switch call.ID {
		case "collect.rbd_image_detail":
			imageLists[strings.Join(call.Args, " ")]++
		case "collect.rbd_image_usage":
			usageCalls[strings.Join(call.Args, " ")]++
		case "collect.rbd_snapshot_children":
			childCalls[strings.Join(call.Args, " ")]++
		}
	}
	for _, command := range []string{
		"ls --long --pool pool --format json",
		"ls --long --pool pool --format json --namespace team",
	} {
		if imageLists[command] != 1 {
			t.Fatalf("image list command %q count=%d, all=%v", command, imageLists[command], imageLists)
		}
	}
	for _, command := range []string{
		"du pool/image --format json",
		"du pool/team/image --format json",
	} {
		if usageCalls[command] != 1 {
			t.Fatalf("usage command %q count=%d, all=%v", command, usageCalls[command], usageCalls)
		}
	}
	for _, command := range []string{
		"children pool/image@snap --all --format json",
		"children pool/team/image@snap --all --format json",
	} {
		if childCalls[command] != 1 {
			t.Fatalf("children command %q count=%d, all=%v", command, childCalls[command], childCalls)
		}
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
	base := malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_image_config": []byte(`[{"name":"rbd_qos_iops_limit","value":"1000","source":"image"}]`),
		"collect.rbd_image_status": []byte(`{"watchers":[{"address":"10.0.0.1:0/1","client":9007199254740993,"cookie":18446744073709551615}],"migration":{"state":"executed"},"persistent_cache":{"clean":true,"cached_bytes":18446744073709551615,"dirty_bytes":0,"hits_full_percent":-1}}`),
		"collect.rbd_image_info":   []byte(`{"name":"image","size":1073741824,"objects":256,"object_size":4194304,"order":22,"stripe_unit":4096,"stripe_count":2,"create_timestamp":"2026-09-23T02:03:04Z","data_pool":"rbd-data","block_name_prefix":"rbd_data.1","features":["layering","exclusive-lock","fast-diff"],"parent":{"pool":"parent-pool","image":"parent","snapshot":"base"},"mirroring":{"mode":"snapshot","state":"enabled","global_id":"global-1","primary":false}}`),
		"collect.rbd_image_usage":  []byte(`{"images":[{"name":"image","snapshot":"base","provisioned_size":1073741824,"used_size":1048576},{"name":"image","provisioned_size":1073741824,"used_size":2097152}]}`),
	}}
	var calls []executor.CommandSpec
	provider := NativeProvider{Executor: recordingExecutor{base: base, calls: &calls}}
	payload := cephdomain.RBDImage{ImagePath: "pool/ns/image"}
	provider.enrichRBDImage(context.Background(), ClusterAccess{}, payload.ImagePath, &payload)
	watchers := objectList(payload.RuntimeStatus["watchers"])
	cache := payload.RuntimeStatus["persistent_cache"].(map[string]any)
	if cache["cached_bytes"] != "18446744073709551615" || cache["dirty_bytes"] != "0" || cache["hits_full_percent"] != "-1" || cache["clean"] != true {
		t.Fatalf("cache metrics lost precision or boolean type: %#v", cache)
	}
	if len(watchers) != 1 || watchers[0]["client"] != "9007199254740993" || watchers[0]["cookie"] != "18446744073709551615" {
		t.Fatalf("watcher identifiers lost precision: %v", watchers)
	}
	if len(payload.Configuration) != 1 || payload.Configuration[0].Source != "image" || payload.Configuration[0].Value != "1000" {
		t.Fatalf("missing effective configuration: %+v", payload)
	}
	if payload.RuntimeStatus["watchers"] == nil || payload.RuntimeStatus["migration"] == nil || payload.RuntimeStatus["persistent_cache"] == nil {
		t.Fatalf("missing runtime status: %+v", payload)
	}
	if len(payload.Features) != 3 || payload.Features[0] != "layering" || payload.Features[2] != "fast-diff" || payload.Parent["snapshot"] != "base" || payload.Details["stripe_unit"] == nil {
		t.Fatalf("missing details: %+v", payload)
	}
	if payload.MirrorMode != "snapshot" || payload.MirrorState != "enabled" || payload.MirrorGlobalID != "global-1" || payload.Primary == nil || *payload.Primary {
		t.Fatalf("missing mirroring state: %+v", payload)
	}
	if payload.SizeBytes == nil || *payload.SizeBytes != 1073741824 || payload.ObjectCount == nil || *payload.ObjectCount != 256 || payload.ObjectSize == nil || *payload.ObjectSize != 4194304 {
		t.Fatalf("missing image capacity details: %+v", payload)
	}
	if payload.StripeUnit == nil || *payload.StripeUnit != 4096 || payload.StripeCount == nil || *payload.StripeCount != 2 || payload.Order == nil || *payload.Order != 22 {
		t.Fatalf("missing image layout details: %+v", payload)
	}
	if payload.CreatedAt != "2026-09-23T02:03:04Z" || payload.DataPool != "rbd-data" || payload.BlockPrefix != "rbd_data.1" {
		t.Fatalf("missing image identity details: %+v", payload)
	}
	encodedImage, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var wireImage map[string]json.RawMessage
	if err := json.Unmarshal(encodedImage, &wireImage); err != nil {
		t.Fatal(err)
	}
	if string(wireImage["image_created_at"]) != `"2026-09-23T02:03:04Z"` || wireImage["created_at"] != nil {
		t.Fatalf("image timestamp collides with inventory metadata: %s", encodedImage)
	}
	if payload.UsedBytes == nil || *payload.UsedBytes != 2097152 || payload.TotalUsedBytes == nil || *payload.TotalUsedBytes != 3145728 {
		t.Fatalf("missing image usage details: %+v", payload)
	}
	foundUsage := false
	for _, call := range calls {
		if call.ID == "collect.rbd_image_usage" {
			foundUsage = strings.Join(call.Args, " ") == "du pool/ns/image --format json"
		}
	}
	if !foundUsage {
		t.Fatalf("missing exact rbd du command: %+v", calls)
	}
}

func TestRBDFeatureAvailabilitySurvivesJSON(t *testing.T) {
	for _, tc := range []struct{ info, want string }{
		{`{"features":[]}`, `[]`},
		{`{"features":["layering"]}`, `["layering"]`},
		{`{}`, `null`},
		{`{"features":null}`, `null`},
		{`{"features":["fast-diff",123]}`, `null`},
		{`{"features":{"unexpected":["fast-diff"]}}`, `null`},
		{`{"features":[""]}`, `null`},
	} {
		var calls []executor.CommandSpec
		provider := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rbd_image_info": []byte(tc.info)}}, calls: &calls}}
		image := cephdomain.RBDImage{ImagePath: "pool/image"}
		provider.enrichRBDImage(context.Background(), ClusterAccess{}, image.ImagePath, &image)
		data, err := json.Marshal(image)
		if err != nil {
			t.Fatal(err)
		}
		var wire map[string]json.RawMessage
		if err := json.Unmarshal(data, &wire); err != nil {
			t.Fatal(err)
		}
		if string(wire["features"]) != tc.want {
			t.Fatalf("%s: %s", tc.info, data)
		}
		for _, call := range calls {
			if call.ID == "collect.rbd_image_usage" {
				t.Fatalf("invalid or absent fast-diff triggered usage: %s", tc.info)
			}
		}
	}
}

func TestSnapshotChildrenFailureDoesNotPublishEmptyDependencies(t *testing.T) {
	for _, output := range []string{`[]`, `null`, `invalid`, `{}`} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.rbd_namespace":         []byte(`[]`),
			"collect.rbd_image_detail":      []byte(`[{"image":"image","size":1024,"format":2}]`),
			"collect.rbd_image_info":        []byte(`{"name":"image","features":[]}`),
			"collect.rbd_snapshot":          []byte(`[{"name":"snap","protected":false}]`),
			"collect.rbd_snapshot_children": []byte(output),
		}}}
		trace := &collectionTrace{unavailable: map[string]struct{}{}}
		ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
		rows := provider.collectStorageOptional(ctx, ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
		count := 0
		for _, row := range rows {
			if row.Kind == "rbd_snapshot" {
				count++
			}
		}
		if output == `[]` {
			if count != 1 {
				t.Fatalf("valid empty dependencies lost: %v", rows)
			}
		} else {
			if count != 0 {
				t.Fatalf("invalid dependency read published snapshot: %s", output)
			}
			if _, ok := trace.unavailable["rbd_snapshot"]; !ok {
				t.Fatalf("snapshot inventory not marked unavailable: %s", output)
			}
		}
	}
}

func TestSnapshotMirroringStillReadsUserSnapshotChildren(t *testing.T) {
	for _, output := range []string{`[{"pool":"p","pool_namespace":"","image":"child","id":"id","trash":true}]`, `invalid`} {
		var calls []executor.CommandSpec
		provider := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.rbd_namespace":         []byte(`[]`),
			"collect.rbd_image_detail":      []byte(`[{"image":"image","size":1024,"format":2}]`),
			"collect.rbd_image_info":        []byte(`{"features":[],"mirroring":{"mode":"snapshot","state":"enabled"}}`),
			"collect.rbd_snapshot":          []byte(`[{"name":"user-snap","protected":false}]`),
			"collect.rbd_snapshot_children": []byte(output),
		}}, calls: &calls}}
		rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
		queried := false
		for _, call := range calls {
			if call.ID == "collect.rbd_snapshot_children" {
				queried = strings.Join(call.Args, " ") == "children pool/image@user-snap --all --format json"
			}
		}
		if !queried {
			t.Fatal("mirroring mode skipped user snapshot dependencies")
		}
		count := 0
		for _, row := range rows {
			if row.Kind == "rbd_snapshot" {
				count++
				children := objectList(row.Payload.(map[string]any)["children"])
				if len(children) != 1 || children[0]["trash"] != true {
					t.Fatal(row)
				}
			}
		}
		if (output == "invalid" && count != 0) || (output != "invalid" && count != 1) {
			t.Fatalf("%s: snapshot count %d", output, count)
		}
	}
}

func TestTrashCollectionIncludesAllSourcesInEachNamespace(t *testing.T) {
	var calls []executor.CommandSpec
	provider := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_namespace":    []byte(`[{"name":"team"}]`),
		"collect.rbd_image_detail": []byte(`[]`),
		"collect.rbd_trash":        []byte(`[{"id":"a","name":"user","source":"USER"},{"id":"b","name":"parent","source":"USER_PARENT"},{"id":"c","name":"mirror","source":"MIRRORING"},{"id":"d","name":"migration","source":"MIGRATION"},{"id":"e","name":"removing","source":"REMOVING"}]`),
	}}, calls: &calls}}
	rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
	counts := map[string]int{}
	for _, row := range rows {
		if row.Kind != "rbd_trash" {
			continue
		}
		payload := row.Payload.(map[string]any)
		counts[payload["trash_source"].(string)]++
	}
	for _, source := range []string{"USER", "USER_PARENT", "MIRRORING", "MIGRATION", "REMOVING"} {
		if counts[source] != 2 {
			t.Fatalf("missing source %s: %v", source, counts)
		}
	}
	commands := map[string]bool{}
	for _, call := range calls {
		if call.ID == "collect.rbd_trash" {
			if call.Mutating || call.Binary != executor.BinaryRBD {
				t.Fatal(call)
			}
			commands[strings.Join(call.Args, " ")] = true
		}
	}
	for _, args := range []string{"trash ls --all --long --pool pool --format json", "trash ls --all --long --pool pool --format json --namespace team"} {
		if !commands[args] {
			t.Fatalf("missing complete source query %s: %v", args, commands)
		}
	}
}

func TestRBDImageInfoDefaultsToDisabledMirroring(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_image_info": []byte(`{"name":"image","features":[]}`),
	}}}
	payload := cephdomain.RBDImage{ImagePath: "pool/image"}
	provider.enrichRBDImage(context.Background(), ClusterAccess{}, payload.ImagePath, &payload)
	if payload.MirrorState != "disabled" || payload.MirrorMode != "" || payload.Primary != nil {
		t.Fatalf("disabled mirroring state=%+v", payload)
	}
}

func TestRBDImageUsageRequiresCurrentImageRow(t *testing.T) {
	trace := &collectionTrace{unavailable: map[string]struct{}{}}
	ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_image_info":  []byte(`{"name":"image","features":["fast-diff"]}`),
		"collect.rbd_image_usage": []byte(`{"images":[{"name":"image","snapshot":"base","used_size":1024}]}`),
	}}}
	payload := cephdomain.RBDImage{ImagePath: "pool/image"}
	provider.enrichRBDImage(ctx, ClusterAccess{}, payload.ImagePath, &payload)
	if payload.UsedBytes != nil || payload.TotalUsedBytes != nil {
		t.Fatalf("malformed usage was accepted: %+v", payload)
	}
	if _, ok := trace.unavailable["rbd_image"]; !ok {
		t.Fatalf("malformed usage did not protect cached image state: %+v", trace.unavailable)
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

func TestRGWRealmDetails(t *testing.T) {
	for _, tc := range []struct {
		name string
		want bool
	}{{"east", true}, {"wrong", false}} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.rgw_realm":        []byte(`{"realms":["east"],"default_info":"realm-id"}`),
			"collect.rgw_realm_detail": []byte(fmt.Sprintf(`{"name":%q,"id":"realm-id","current_period":"period-id","epoch":4}`, tc.name)),
		}}}
		found := false
		for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
			if row.Kind == "rgw_realm" {
				found = true
				if row.Payload.(map[string]any)["current_period"] != "period-id" || row.Payload.(map[string]any)["is_default"] != true {
					t.Fatal("period missing")
				}
			}
		}
		if found != tc.want {
			t.Fatalf("realm %s found=%v", tc.name, found)
		}
	}
}

func TestRGWZonegroupDetails(t *testing.T) {
	for _, tc := range []struct {
		name, detail string
		want         bool
	}{
		{"valid", `{"name":"east","id":"zg-id","zones":[{"id":"zone-id","name":"zone-a"}],"endpoints":["https://rgw.example"],"placement_targets":[{"name":"default-placement"}]}`, true},
		{"wrong name", `{"name":"west","id":"zg-id"}`, false},
		{"missing id", `{"name":"east"}`, false},
		{"null", `null`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_zonegroup":        []byte(`{"zonegroups":["east"]}`),
				"collect.rgw_zonegroup_detail": []byte(tc.detail),
			}}}
			found := false
			for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
				if row.Kind != "rgw_zonegroup" {
					continue
				}
				found = true
				data := row.Payload.(map[string]any)
				if row.NaturalKey != "east" || len(data["zones"].([]any)) != 1 || len(data["endpoints"].([]any)) != 1 || len(data["placement_targets"].([]any)) != 1 {
					t.Fatal("zonegroup detail lost")
				}
			}
			if found != tc.want {
				t.Fatalf("found=%v want=%v", found, tc.want)
			}
		})
	}
}

func TestRGWZoneDetails(t *testing.T) {
	for _, tc := range []struct {
		name, detail string
		want         bool
	}{
		{"valid", `{"name":"east","id":"zone-id","realm_id":"realm-id","control_pool":"east.rgw.control","placement_pools":[{"key":"default-placement","val":{"index_pool":"east.rgw.buckets.index","storage_classes":{"STANDARD":{"data_pool":"east.rgw.buckets.data"}}}}],"system_key":{"access_key":"private-access","secret_key":"private-secret"}}`, true},
		{"wrong name", `{"name":"west","id":"zone-id"}`, false},
		{"missing id", `{"name":"east"}`, false},
		{"null", `null`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_zone":        []byte(`{"zones":["east"]}`),
				"collect.rgw_zone_detail": []byte(tc.detail),
			}}}
			found := false
			for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
				if row.Kind != "rgw_zone" {
					continue
				}
				found = true
				data := row.Payload.(map[string]any)
				if row.NaturalKey != "east" || data["control_pool"] != "east.rgw.control" || len(data["placement_pools"].([]any)) != 1 {
					t.Fatal("zone detail lost")
				}
				redacted, err := security.RedactJSON(data)
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := json.Marshal(redacted)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(encoded), "private-") || !strings.Contains(string(encoded), "east.rgw.buckets.data") {
					t.Fatal("zone redaction must hide credentials and preserve placement pools")
				}
			}
			if found != tc.want {
				t.Fatalf("found=%v want=%v", found, tc.want)
			}
		})
	}
}

func TestMultisiteDefaultIdentity(t *testing.T) {
	for _, resource := range []struct{ kind, key string }{{"rgw_realm", "realms"}, {"rgw_zonegroup", "zonegroups"}, {"rgw_zone", "zones"}} {
		for _, tc := range []struct {
			label, value string
			known, want  bool
		}{{"match", `,"default_info":"id"`, true, true}, {"other", `,"default_info":"other"`, true, false}, {"empty", `,"default_info":""`, true, false}, {"missing", "", false, false}, {"invalid", `,"default_info":42`, false, false}} {
			t.Run(resource.kind+"/"+tc.label, func(t *testing.T) {
				p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
					"collect." + resource.kind:             []byte(fmt.Sprintf(`{"%s":["east"]%s}`, resource.key, tc.value)),
					"collect." + resource.kind + "_detail": []byte(`{"name":"east","id":"id"}`),
				}}}
				found := false
				for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
					if row.Kind != resource.kind {
						continue
					}
					found = true
					value, known := row.Payload.(map[string]any)["is_default"]
					if known != tc.known || (known && value != tc.want) {
						t.Fatalf("default=%v known=%v", value, known)
					}
				}
				if !found {
					t.Fatal("missing multisite observation")
				}
			})
		}
	}
}

func TestZoneMembershipJoinUsesNativeID(t *testing.T) {
	target := map[string]any{"id": "zone-id", "name": "renamed"}
	other := map[string]any{"id": "other-id", "name": "old-name"}
	rows := []Observation{
		{Kind: "rgw_zone", Payload: target}, {Kind: "rgw_zone", Payload: other},
		{Kind: "rgw_zonegroup", Name: "group-a", Payload: map[string]any{"id": "group-id", "master_zone": "zone-id", "zones": []any{map[string]any{"id": "zone-id", "name": "old-name", "tier_type": "archive", "sync_from_all": false, "endpoints": []any{"https://rgw.example"}}}}},
		{Kind: "rgw_zonegroup", Name: "group-b", Payload: map[string]any{"id": "group-b-id", "zones": []any{map[string]any{"id": "zone-id", "name": "renamed", "read_only": true}}}},
	}
	attachZoneMemberships(rows)
	matches := target["zonegroup_memberships"].([]any)
	if len(matches) != 2 {
		t.Fatal("lost group membership")
	}
	first := matches[0].(map[string]any)
	if first["zonegroup_name"] != "group-a" || first["is_master"] != true || first["tier_type"] != "archive" || first["sync_from_all"] != false {
		t.Fatalf("wrong membership %v", first)
	}
	if _, present := other["zonegroup_memberships"]; present {
		t.Fatal("joined unrelated zone by name")
	}
	if _, present := target["tier_type"]; present {
		t.Fatal("group context must remain explicit")
	}
}
