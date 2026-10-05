package ceph

import (
	"context"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestAccountCollectionRequiresMatchingNativeID(t *testing.T) {
	for _, raw := range []string{
		`{"id":"RGW123","name":"Account"}`,
		`{"id":"other","name":"Account"}`,
		`{"id":" RGW123 ","name":"Account"}`,
		`{"name":"Account"}`, `{"id":null}`, `{"id":123}`, `{"id":false}`, `{"id":{}}`,
	} {
		t.Run(raw, func(t *testing.T) {
			valid := raw == `{"id":"RGW123","name":"Account"}`
			calls := []executor.CommandSpec{}
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			p := NativeProvider{Executor: recordingExecutor{calls: &calls, base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_account":        []byte(`["RGW123"]`),
				"collect.rgw_account_detail": []byte(raw),
			}}}}
			count := 0
			for _, row := range p.collectRGWOptional(ctx, ClusterAccess{}, time.Now()) {
				if row.Kind != "rgw_account" {
					continue
				}
				count++
				if row.Payload.(map[string]any)["account_id"] != "RGW123" {
					t.Fatal("wrong account identity")
				}
			}
			if (count == 1) != valid {
				t.Fatalf("unexpected account count: %d", count)
			}
			if !valid {
				for _, kind := range []string{"rgw_account", "rgw_role"} {
					if _, marked := trace.unavailable[kind]; !marked {
						t.Fatalf("missing unavailable %s", kind)
					}
				}
				for _, call := range calls {
					if call.ID == "collect.rgw_account_stats" {
						t.Fatal("collected stats after mismatched identity")
					}
					for _, arg := range call.Args {
						if call.ID == "collect.rgw_role" && arg == "--account-id" {
							t.Fatal("collected account roles after mismatched identity")
						}
					}
				}
			}
		})
	}
}
