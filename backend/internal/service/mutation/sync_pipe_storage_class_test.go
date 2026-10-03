package mutation

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
)

func TestSyncPipeStorageClass(t *testing.T) {
	for _, value := range []any{"", "COLD", "归档", " spaced "} {
		got, present, err := syncPipeStorageClass(map[string]any{"storage_class": value})
		if err != nil || !present || got != value {
			t.Fatalf("valid %v: %v", value, err)
		}
	}
	for _, value := range []any{nil, 0, true, "-bad", "x\n", strings.Repeat("x", 513), string([]byte{0xff})} {
		if _, _, err := syncPipeStorageClass(map[string]any{"storage_class": value}); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
	if _, present, err := syncPipeStorageClass(map[string]any{}); present || err != nil {
		t.Fatal("omission changed class")
	}
}

func TestSyncPipeStorageClassAbsentToEmpty(t *testing.T) {
	dest := map[string]any{"unknown": "keep"}
	pipe := map[string]any{"id": "p", "source": map[string]any{"bucket": "*"}, "dest": map[string]any{"bucket": "*"}, "params": map[string]any{"mode": "system", "dest": dest}}
	group := map[string]any{"pipes": []any{pipe}}
	p := map[string]any{"pipe_id": "p", "source_bucket": "*", "dest_bucket": "*", "mode": "system", "storage_class": ""}
	if err := updateBucketSyncPipe(group, p); err != nil {
		t.Fatal(err)
	}
	if value, exists := dest["storage_class"]; !exists || value != "" || dest["unknown"] != "keep" {
		t.Fatalf("destination %+v", dest)
	}
	if err := updateBucketSyncPipe(group, p); err == nil {
		t.Fatal("accepted unchanged explicit empty class")
	}
}

func TestBucketSyncPipeStorageClassOnly(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","source":{"bucket":"*","zones":["A"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"system","priority":7,"source":{"filter":{"prefix":"keep"}},"dest":{"storage_class":"COLD","acl_translation":{"owner":"u"}}}}]}`
	policy := func(g string) string { return `{"groups":[` + g + `]}` }
	for _, value := range []string{"ARCHIVE", ""} {
		for _, scenario := range []string{"success", "wrong class", "other changed", "unchanged"} {
			t.Run(value+scenario, func(t *testing.T) {
				service, _, cluster := newCephUserService(t)
				after := strings.Replace(group, `"storage_class":"COLD"`, `"storage_class":"`+value+`"`, 1)
				if scenario == "wrong class" {
					after = group
				}
				if scenario == "other changed" {
					after = strings.Replace(after, `"owner":"u"`, `"owner":"other"`, 1)
				}
				runner := &syncGroupExecutor{before: policy(group), after: policy(after)}
				service.executor = runner
				p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket")), "group_id": "g", "pipe_id": "p", "expected_group": group, "source_bucket": "*", "dest_bucket": "*", "mode": "system", "storage_class": value}
				if scenario == "unchanged" {
					p["storage_class"] = "COLD"
				}
				_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_bucket.sync_pipe_update", Parameters: p})
				if (err == nil) != (scenario == "success") {
					t.Fatalf("error %v", err)
				}
				if scenario == "unchanged" {
					if len(runner.calls) != 1 {
						t.Fatal("unchanged wrote")
					}
					return
				}
				if len(runner.calls) != 3 {
					t.Fatalf("calls %+v", runner.calls)
				}
				found := false
				for i, arg := range runner.calls[1].Args {
					if arg == "--storage-class" {
						found = true
						if runner.calls[1].Args[i+1] != value {
							t.Fatal("class changed")
						}
					}
				}
				if !found {
					t.Fatal("missing flag")
				}
			})
		}
	}
}
