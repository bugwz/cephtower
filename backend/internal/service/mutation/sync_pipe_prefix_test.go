package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestSyncPipePrefixValidation(t *testing.T) {
	for _, value := range []string{"", "--leading", "目录/ ", "a=b", strings.Repeat("x", 1024)} {
		mode, got, err := syncPipePrefix(map[string]any{"prefix_mode": "set", "source_prefix": value})
		if err != nil || mode != "set" || got != value {
			t.Fatalf("valid %q: %v", value, err)
		}
	}
	for _, p := range []map[string]any{{"source_prefix": "x"}, {"prefix_mode": "unknown"}, {"prefix_mode": "set"}, {"prefix_mode": "set", "source_prefix": nil}, {"prefix_mode": "set", "source_prefix": "x\n"}, {"prefix_mode": "set", "source_prefix": strings.Repeat("x", 1025)}, {"prefix_mode": "set", "source_prefix": string([]byte{0xff})}, {"prefix_mode": "remove", "source_prefix": ""}} {
		if _, _, err := syncPipePrefix(p); err == nil {
			t.Fatalf("accepted %+v", p)
		}
	}
}

func TestBucketSyncPipePrefixOnly(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","source":{"bucket":"*","zones":["A"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"system","priority":7,"source":{"filter":{"prefix":"old","tags":[{"key":"k","value":"v"}]}},"dest":{"storage_class":"COLD"}}}]}`
	policy := func(g string) string { return `{"groups":[` + g + `]}` }
	for _, mode := range []string{"set", "empty", "remove"} {
		for _, scenario := range []string{"success", "unchanged", "wrong prefix", "tags changed", "write", "post"} {
			t.Run(mode+scenario, func(t *testing.T) {
				service, _, cluster := newCephUserService(t)
				p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket")), "group_id": "g", "pipe_id": "p", "expected_group": group, "source_bucket": "*", "dest_bucket": "*", "mode": "system", "prefix_mode": "set", "source_prefix": "--目录/ "}
				before := group
				if mode == "empty" {
					p["source_prefix"] = ""
				}
				if mode == "remove" {
					p["prefix_mode"] = "remove"
					delete(p, "source_prefix")
				}
				if scenario == "unchanged" {
					if mode == "remove" {
						before = strings.Replace(group, `"prefix":"old",`, "", 1)
					} else {
						p["source_prefix"] = "old"
					}
					p["expected_group"] = before
				}
				after := strings.Replace(before, `"prefix":"old",`, "", 1)
				if mode != "remove" {
					encoded, _ := json.Marshal(p["source_prefix"])
					after = strings.Replace(before, `"prefix":"old"`, `"prefix":`+string(encoded), 1)
				}
				if scenario == "wrong prefix" {
					after = before
				}
				if scenario == "tags changed" {
					after = strings.Replace(after, `"value":"v"`, `"value":"other"`, 1)
				}
				runner := &syncGroupExecutor{before: policy(before), after: policy(after)}
				if scenario == "write" {
					runner.failAt = 2
				}
				if scenario == "post" {
					runner.failAt = 3
				}
				service.executor = runner
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
					for i, arg := range runner.calls[1].Args {
						if mode == "remove" && arg == "--prefix-rm" {
							found = runner.calls[1].Args[i+1] == "true"
						}
						if mode != "remove" && arg == "--prefix="+p["source_prefix"].(string) {
							found = true
						}
					}
					if !found {
						t.Fatalf("prefix args %+v", runner.calls[1].Args)
					}
				}
			})
		}
	}
}
