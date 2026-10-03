package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestBucketSyncPolicyCollection(t *testing.T) {
	for _, tenant := range []string{"", "team"} {
		for _, tc := range []struct {
			name, payload string
			valid         bool
		}{
			{"pipe details", `{"groups":[{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","source":{"bucket":"team/photos:marker","zones":["*"]},"dest":{"bucket":"*","zones":["Zone B"]},"params":{"source":{"filter":{"prefix":"","tags":[{"key":"<key>","value":""}]}},"dest":{"acl_translation":{"owner":"team$user"},"storage_class":"COLD"},"priority":0,"mode":"user","user":"team$user","future":{"enabled":true}}}]}]}`, true},
			{"both flow types", `{"groups":[{"id":"g","status":"enabled","data_flow":{"symmetrical":[{"id":"双向","zones":[" a ","<b>"]}],"directional":[{"source_zone":"b","dest_zone":"c"}],"future":{"option":true}},"pipes":[]}]}`, true},
			{"enabled", `{"groups":[{"id":"g","status":"enabled","data_flow":{"directional":[{"source_zone":"a","dest_zone":"b"}]},"pipes":[{"id":"p"}]}]}`, true},
			{"allowed", `{"groups":[{"id":"g","status":"allowed","data_flow":{},"pipes":[]}]}`, true},
			{"forbidden", `{"groups":[{"id":"g","status":"forbidden","data_flow":{},"pipes":[]}]}`, true},
			{"unknown", `{"groups":[{"id":"g","status":"future","data_flow":{},"pipes":[]}]}`, true},
			{"empty", `{"groups":[]}`, true},
			{"missing groups", `{}`, false},
			{"command unavailable", "", false},
			{"null", `null`, false}, {"broken", `broken`, false},
			{"missing status", `{"groups":[{"id":"g","data_flow":{},"pipes":[]}]}`, false},
			{"bad pipes", `{"groups":[{"id":"g","status":"enabled","data_flow":{},"pipes":{}}]}`, false},
			{"duplicate id", `{"groups":[{"id":"g","status":"enabled","data_flow":{},"pipes":[]},{"id":"g","status":"forbidden","data_flow":{},"pipes":[]}]}`, false},
		} {
			t.Run(tenant+"/"+tc.name, func(t *testing.T) {
				calls := []executor.CommandSpec{}
				entry := "photos"
				if tenant != "" {
					entry = tenant + "/" + entry
				}
				overrides := map[string][]byte{
					"collect.rgw_bucket":             []byte(`["` + entry + `"]`),
					"collect.rgw_bucket_detail":      []byte(`{"bucket":"photos","tenant":"` + tenant + `","marker":"m1"}`),
					"collect.rgw_bucket_sync_policy": []byte(tc.payload),
				}
				if tc.name == "command unavailable" {
					delete(overrides, "collect.rgw_bucket_sync_policy")
				}
				provider := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: overrides}, calls: &calls}}
				found := false
				for _, row := range provider.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
					if row.Kind != "rgw_bucket" {
						continue
					}
					found = true
					value := row.Payload.(map[string]any)["bucket_sync_policy"]
					if (value != nil) != tc.valid {
						t.Fatalf("wrong availability: %#v", value)
					}
					if tc.valid {
						var original map[string]any
						decoder := json.NewDecoder(strings.NewReader(tc.payload))
						decoder.UseNumber()
						if err := decoder.Decode(&original); err != nil || !reflect.DeepEqual(value, original) {
							t.Fatalf("native policy fields changed: %#v", value)
						}
						policy, ok := rgwBucketSyncPolicy(value)
						if !ok {
							t.Fatal("invalid observed policy")
						}
						groups := policy["groups"].([]any)
						if tc.name == "enabled" && len(groups[0].(map[string]any)["pipes"].([]any)) != 1 {
							t.Fatal("raw pipes lost")
						}
					}
				}
				if !found {
					t.Fatal("bucket omitted")
				}
				count := 0
				for _, call := range calls {
					if call.ID != "collect.rgw_bucket_sync_policy" {
						continue
					}
					count++
					if call.Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(call.Args, []string{"sync", "policy", "get", "--bucket", "photos", "--tenant", tenant, "--format", "json"}) {
						t.Fatalf("wrong scoped command: %+v", call)
					}
				}
				if count != 1 {
					t.Fatalf("policy reads: %d", count)
				}
			})
		}
	}
}
