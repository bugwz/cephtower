package ceph

import (
	"context"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWLifecycleProgressCollection(t *testing.T) {
	for _, tc := range []struct {
		name, payload    string
		available, found bool
		status           string
	}{
		{"exact identity", `[{"bucket":"other:photos:m1","status":"FAILED"},{"bucket":"team:photos-old:m1","status":"FAILED"},{"bucket":"team:photos:old","status":"FAILED"},{"bucket":"team:photos:m1","status":"PROCESSING","started":"Fri, 02 Oct 2026 00:00:00 GMT"}]`, true, true, "PROCESSING"},
		{"unknown status", `[{"bucket":"team:photos:m1","status":"FUTURE"}]`, true, true, "FUTURE"},
		{"empty list", `[]`, true, false, ""},
		{"other generation", `[{"bucket":"team:photos:m0","status":"COMPLETE"}]`, true, false, ""},
		{"duplicate key", `[{"bucket":"team:photos:m1","status":"COMPLETE"},{"bucket":"team:photos:m1","status":"FAILED"}]`, false, false, ""},
		{"null", `null`, false, false, ""}, {"object", `{}`, false, false, ""},
		{"malformed", `broken`, false, false, ""}, {"missing status", `[{"bucket":"team:photos:m1"}]`, false, false, ""},
		{"invalid timestamp", `[{"bucket":"team:photos:m1","status":"COMPLETE","started":12}]`, false, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := []executor.CommandSpec{}
			provider := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_bucket":             []byte(`["team/photos"]`),
				"collect.rgw_bucket_detail":      []byte(`{"bucket":"photos","tenant":"team","marker":"m1"}`),
				"collect.rgw_lifecycle_progress": []byte(tc.payload),
			}}, calls: &calls}}
			seen := false
			for _, row := range provider.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
				if row.Kind != "rgw_bucket" {
					continue
				}
				seen = true
				value := row.Payload.(map[string]any)["lifecycle_progress"]
				if !tc.available {
					if value != nil {
						t.Fatalf("unavailable became %v", value)
					}
					continue
				}
				progress := value.(map[string]any)
				if progress["found"] != tc.found {
					t.Fatalf("wrong record state: %v", value)
				}
				if tc.found && progress["entry"].(map[string]any)["status"] != tc.status {
					t.Fatalf("wrong association: %v", value)
				}
				if !tc.found && progress["entry"] != nil {
					t.Fatalf("unexpected entry: %v", value)
				}
			}
			if !seen {
				t.Fatal("bucket omitted")
			}
			count := 0
			for _, call := range calls {
				if call.ID == "collect.rgw_lifecycle_progress" {
					count++
					if call.Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(call.Args, []string{"lc", "list", "--format", "json"}) {
						t.Fatalf("wrong command: %+v", call)
					}
				}
			}
			if count != 1 {
				t.Fatalf("lc list calls=%d", count)
			}
		})
	}
}

func TestRGWLifecycleProgressMarkerRequired(t *testing.T) {
	if value := rgwBucketLifecycleProgress(map[string]any{"tenant": "team", "bucket": "photos"}, nil, true); value != nil {
		t.Fatal("missing marker treated as no record")
	}
	entries := map[string]map[string]any{":photos:m1": {"status": "COMPLETE"}}
	value := rgwBucketLifecycleProgress(map[string]any{"tenant": "", "bucket": "photos", "marker": "m1"}, entries, true).(map[string]any)
	if value["found"] != true {
		t.Fatal("global tenant identity lost")
	}
}
