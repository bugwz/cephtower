package ceph

import (
	"context"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestBucketCollectionRequiresExplicitIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, entry, raw string
		valid            bool
	}{
		{"default tenant", "photos", `{"bucket":"photos","tenant":""}`, true},
		{"named tenant", "team/photos", `{"bucket":"photos","tenant":"team"}`, true},
		{"missing tenant", "photos", `{"bucket":"photos"}`, false},
		{"null tenant", "photos", `{"bucket":"photos","tenant":null}`, false},
		{"number tenant", "photos", `{"bucket":"photos","tenant":0}`, false},
		{"wrong tenant", "photos", `{"bucket":"photos","tenant":"team"}`, false},
		{"whitespace tenant", "team/photos", `{"bucket":"photos","tenant":" team "}`, false},
		{"wrong bucket", "photos", `{"bucket":"other","tenant":""}`, false},
		{"missing bucket", "photos", `{"tenant":""}`, false},
		{"null bucket", "photos", `{"bucket":null,"tenant":""}`, false},
		{"null response", "photos", `null`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := []executor.CommandSpec{}
			p := NativeProvider{Executor: recordingExecutor{calls: &calls, base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_bucket":        []byte(`["` + tc.entry + `"]`),
				"collect.rgw_bucket_detail": []byte(tc.raw),
			}}}}
			count := 0
			for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
				if row.Kind == "rgw_bucket" {
					count++
				}
			}
			if (count == 1) != tc.valid {
				t.Fatalf("unexpected bucket count: %d", count)
			}
			for _, call := range calls {
				if !tc.valid && (call.ID == "collect.rgw_bucket_ratelimit" || call.ID == "collect.rgw_bucket_sync_policy") {
					t.Fatalf("continued collection after invalid identity: %+v", call)
				}
			}
		})
	}
}
