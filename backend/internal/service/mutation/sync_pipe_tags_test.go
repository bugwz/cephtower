package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestSyncPipeTagValidation(t *testing.T) {
	pair := func(k, v string) map[string]any { return map[string]any{"key": k, "value": v} }
	for _, tag := range []map[string]any{pair("", ""), pair("-key", "v=more"), pair(" 标签 ", " value ")} {
		args, err := syncPipeTagArgs(map[string]any{"tags_add": []any{tag}})
		if err != nil || len(args) != 1 || args[0] != "--tags-add="+tag["key"].(string)+"="+tag["value"].(string) {
			t.Fatalf("args %+v %v", args, err)
		}
	}
	for _, raw := range []any{nil, "x", []any{nil}, []any{pair("k,", "v")}, []any{pair("k=", "v")}, []any{pair("k", "v,")}, []any{pair("k", "v\n")}, []any{pair("k", strings.Repeat("x", 1025))}, []any{pair("k", string([]byte{0xff}))}, []any{pair("k", "v"), pair("k", "v")}, []any{map[string]any{"key": "k"}}} {
		if _, err := syncPipeTagArgs(map[string]any{"tags_add": raw}); err == nil {
			t.Fatalf("accepted %+v", raw)
		}
	}
	if _, err := syncPipeTagArgs(map[string]any{"tags_add": []any{pair("k", "v")}, "tags_remove": []any{pair("k", "v")}}); err == nil {
		t.Fatal("conflicting pair")
	}
	for _, tags := range []any{nil, []any{pair("k", "v"), pair("k", "v")}, []any{map[string]any{"key": "k", "value": 1}}} {
		if _, err := updateSyncPipeTags(map[string]any{"source": map[string]any{"filter": map[string]any{"tags": tags}}}, map[string]any{"tags_add": []any{pair("x", "y")}}); err == nil {
			t.Fatal("invalid current tags")
		}
	}
}

func TestBucketSyncPipeTagsOnly(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","source":{"bucket":"*","zones":["A"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"system","priority":7,"source":{"filter":{"prefix":"keep","tags":[{"key":"k","value":"v"}]}},"dest":{"storage_class":"COLD"}}}]}`
	policy := func(g string) string { return `{"groups":[` + g + `]}` }
	for _, change := range []struct{ name, add, remove, want string }{
		{"add", `[{"key":"k","value":"other"}]`, `[]`, `[{"key":"k","value":"other"},{"key":"k","value":"v"}]`},
		{"remove", `[]`, `[{"key":"k","value":"v"}]`, `[]`},
		{"replace", `[{"key":"k","value":""}]`, `[{"key":"k","value":"v"}]`, `[{"key":"k","value":""}]`},
	} {
		for _, scenario := range []string{"success", "unchanged", "wrong tags", "prefix changed", "write", "post"} {
			t.Run(change.name+scenario, func(t *testing.T) {
				service, _, cluster := newCephUserService(t)
				var add, remove []any
				json.Unmarshal([]byte(change.add), &add)
				json.Unmarshal([]byte(change.remove), &remove)
				before := group
				after := strings.Replace(group, `[{"key":"k","value":"v"}]`, change.want, 1)
				if scenario == "unchanged" {
					before = after
				}
				if scenario == "wrong tags" {
					after = group
				}
				if scenario == "prefix changed" {
					after = strings.Replace(after, `"prefix":"keep"`, `"prefix":"other"`, 1)
				}
				runner := &syncGroupExecutor{before: policy(before), after: policy(after)}
				if scenario == "write" {
					runner.failAt = 2
				}
				if scenario == "post" {
					runner.failAt = 3
				}
				service.executor = runner
				p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket")), "group_id": "g", "pipe_id": "p", "expected_group": before, "source_bucket": "*", "dest_bucket": "*", "mode": "system", "tags_add": add, "tags_remove": remove}
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
					args, _ := syncPipeTagArgs(p)
					joined := strings.Join(runner.calls[1].Args, "|")
					for _, arg := range args {
						if !strings.Contains(joined, arg) {
							t.Fatal(joined)
						}
					}
				}
			})
		}
	}
}
