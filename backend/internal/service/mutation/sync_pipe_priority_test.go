package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestSyncPipePriority(t *testing.T) {
	for _, value := range []any{-2147483648, 0, 2147483647, json.Number("0"), float64(-1), float64(2147483647), float64(-2147483648)} {
		if _, present, err := syncPipePriority(map[string]any{"priority": value}); err != nil || !present {
			t.Fatalf("valid %v: %v", value, err)
		}
	}
	for _, value := range []any{int64(-2147483649), int64(2147483648), 0.5, "0", nil, true, json.Number("1.5"), json.Number("1e2")} {
		if _, _, err := syncPipePriority(map[string]any{"priority": value}); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
	if _, present, err := syncPipePriority(map[string]any{}); present || err != nil {
		t.Fatal("omission changed priority")
	}
}

func TestBucketSyncPipePriorityOnly(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","source":{"bucket":"*","zones":["A"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"system","priority":7,"source":{"filter":{"prefix":"keep"}},"dest":{"storage_class":"COLD"}}}]}`
	policy := func(g string) string { return `{"groups":[` + g + `]}` }
	for _, scenario := range []string{"success", "unchanged", "wrong priority", "other changed"} {
		t.Run(scenario, func(t *testing.T) {
			service, _, cluster := newCephUserService(t)
			after := strings.Replace(group, `"priority":7`, `"priority":0`, 1)
			if scenario == "wrong priority" {
				after = group
			}
			if scenario == "other changed" {
				after = strings.Replace(after, "COLD", "HOT", 1)
			}
			runner := &syncGroupExecutor{before: policy(group), after: policy(after)}
			service.executor = runner
			p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket")), "group_id": "g", "pipe_id": "p", "expected_group": group, "source_bucket": "*", "dest_bucket": "*", "mode": "system", "priority": 0}
			if scenario == "unchanged" {
				p["priority"] = 7
			}
			_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_bucket.sync_pipe_update", Parameters: p})
			if (err == nil) != (scenario == "success") {
				t.Fatalf("error: %v", err)
			}
			if scenario == "unchanged" {
				if len(runner.calls) != 1 {
					t.Fatal("unchanged wrote")
				}
				return
			}
			if len(runner.calls) != 3 {
				t.Fatalf("calls %v", runner.calls)
			}
			args := strings.Join(runner.calls[1].Args, " ")
			if !strings.Contains(args, "--priority 0 --bucket bucket --tenant team") {
				t.Fatal(args)
			}
		})
	}
}
