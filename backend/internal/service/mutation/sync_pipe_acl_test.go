package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestSyncPipeDestinationOwner(t *testing.T) {
	for _, uid := range []string{"", "u", "team$u", "$ns$u", "team$ns$u", "team$ns$u$extra"} {
		got, present, err := syncPipeDestinationOwner(map[string]any{"dest_owner": uid})
		if err != nil || !present || got != uid {
			t.Fatalf("valid %q: %v", uid, err)
		}
	}
	for _, uid := range []any{nil, 1, "-x", "u\n", "u x", "$u", "team$", "team$$u", "$ns$", strings.Repeat("x", 513), string([]byte{0xff})} {
		if _, _, err := syncPipeDestinationOwner(map[string]any{"dest_owner": uid}); err == nil {
			t.Fatalf("accepted %v", uid)
		}
	}
	params := map[string]any{"dest": map[string]any{"storage_class": "COLD", "unknown": true}}
	if changed, err := updateSyncPipeACL(params, map[string]any{"dest_owner": ""}); changed || err != nil {
		t.Fatal("absent removal changed")
	}
	if changed, err := updateSyncPipeACL(params, map[string]any{"dest_owner": "team$ns$u"}); !changed || err != nil {
		t.Fatal("absent set failed")
	}
	if changed, err := updateSyncPipeACL(params, map[string]any{"dest_owner": "team$ns$u"}); changed || err != nil {
		t.Fatal("unchanged set changed")
	}
	if params["dest"].(map[string]any)["unknown"] != true {
		t.Fatal("lost unknown destination field")
	}
}

func TestBucketSyncPipeACLOnly(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","source":{"bucket":"*","zones":["A"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"system","priority":7,"source":{"filter":{"prefix":"keep","tags":[]}},"dest":{"acl_translation":{"owner":"old"},"storage_class":"COLD"}}}]}`
	policy := func(g string) string { return `{"groups":[` + g + `]}` }
	for _, owner := range []string{"u", "team$u", "$ns$u", "team$ns$u", ""} {
		for _, scenario := range []string{"success", "unchanged", "wrong owner", "storage changed", "write", "post"} {
			t.Run(owner+scenario, func(t *testing.T) {
				service, _, cluster := newCephUserService(t)
				encoded, _ := json.Marshal(owner)
				after := strings.Replace(group, `"owner":"old"`, `"owner":`+string(encoded), 1)
				if owner == "" {
					after = strings.Replace(group, `"acl_translation":{"owner":"old"},`, "", 1)
				}
				before := group
				if scenario == "unchanged" {
					before = after
				}
				if scenario == "wrong owner" {
					after = group
				}
				if scenario == "storage changed" {
					after = strings.Replace(after, "COLD", "HOT", 1)
				}
				runner := &syncGroupExecutor{before: policy(before), after: policy(after)}
				if scenario == "write" {
					runner.failAt = 2
				}
				if scenario == "post" {
					runner.failAt = 3
				}
				service.executor = runner
				p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket")), "group_id": "g", "pipe_id": "p", "expected_group": before, "source_bucket": "*", "dest_bucket": "*", "mode": "system", "dest_owner": owner}
				_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_bucket.sync_pipe_update", Parameters: p})
				if (err == nil) != (scenario == "success") {
					t.Fatalf("error %v", err)
				}
				count := 3
				if scenario == "unchanged" {
					count = 1
				}
				if scenario == "write" {
					count = 2
				}
				if len(runner.calls) != count {
					t.Fatalf("calls %+v", runner.calls)
				}
				if count > 1 {
					found := false
					for _, arg := range runner.calls[1].Args {
						if arg == "--dest-owner="+owner {
							found = true
						}
					}
					if !found {
						t.Fatal(runner.calls[1].Args)
					}
				}
			})
		}
	}
}
