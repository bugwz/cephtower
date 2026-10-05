package ceph

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestRoleCollectionRequiresNativeStringIdentity(t *testing.T) {
	for _, account := range []string{"", "123", "RGW123"} {
		for _, field := range []string{"RoleName", "AccountId"} {
			for _, value := range []any{nil, 123.0, json.Number("123"), false, []any{}, map[string]any{}} {
				role := map[string]any{"RoleName": "reader", "AccountId": account}
				role[field] = value
				trace := &collectionTrace{unavailable: map[string]struct{}{}}
				ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
				if rows := rgwRoleObservations(ctx, []any{role}, account, time.Now()); len(rows) != 0 {
					t.Fatalf("published invalid %s=%v: %+v", field, value, rows)
				}
				if _, marked := trace.unavailable["rgw_role"]; !marked {
					t.Fatal("invalid identity reported as complete inventory")
				}
			}
		}
	}
}

func TestNativeRoleCollectionRejectsMissingAccountIdentity(t *testing.T) {
	for _, raw := range []string{
		`[{"RoleName":"reader","AccountId":""}]`,
		`[{"RoleName":"reader"}]`, `[{"RoleName":123,"AccountId":""}]`,
		`[{"RoleName":"reader","AccountId":null}]`, `null`, `[]`,
	} {
		t.Run(raw, func(t *testing.T) {
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_role":    []byte(raw),
				"collect.rgw_account": []byte(`[]`),
			}}}
			count := 0
			for _, row := range p.collectRGWOptional(ctx, ClusterAccess{}, time.Now()) {
				if row.Kind == "rgw_role" {
					count++
					if row.NaturalKey != "reader" || row.Payload.(map[string]any)["AccountId"] != "" {
						t.Fatalf("altered identity: %+v", row)
					}
				}
			}
			valid := raw == `[{"RoleName":"reader","AccountId":""}]`
			if (count == 1) != valid {
				t.Fatalf("unexpected role count: %d", count)
			}
			_, unavailable := trace.unavailable["rgw_role"]
			if unavailable == (valid || raw == `[]`) {
				t.Fatal("incorrect collection completeness")
			}
		})
	}
}
