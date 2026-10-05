package ceph

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestUserCollectionRequiresExactFullUID(t *testing.T) {
	for _, uid := range []string{"user", "team$user", "team$namespace$user"} {
		for _, identity := range []any{uid, "other", " " + uid + " ", nil, 1, false, map[string]any{}} {
			observed, ok := identity.(string)
			valid := ok && observed == uid
			list, _ := json.Marshal([]string{uid})
			details := map[string]any{"user_id": "user", "account_id": "RGW123", "type": "rgw"}
			if identity != nil {
				details["full_user_id"] = identity
			}
			raw, _ := json.Marshal(details)
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			calls := []executor.CommandSpec{}
			p := NativeProvider{Executor: recordingExecutor{calls: &calls, base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_user": list, "collect.rgw_user_detail": raw,
			}}}}
			count := 0
			for _, row := range p.collectRGWOptional(ctx, ClusterAccess{}, time.Now()) {
				if row.Kind != "rgw_user" {
					continue
				}
				count++
				payload := row.Payload.(map[string]any)
				if payload["uid"] != uid || payload["full_user_id"] != uid || payload["user_id"] != "user" {
					t.Fatalf("identity changed: %+v", payload)
				}
			}
			if (count == 1) != valid {
				t.Fatalf("uid=%s identity=%v count=%d", uid, identity, count)
			}
			if !valid {
				if _, marked := trace.unavailable["rgw_user"]; !marked {
					t.Fatal("invalid identity not marked unavailable")
				}
				for _, call := range calls {
					switch call.ID {
					case "collect.rgw_user_stats", "collect.rgw_user_policies", "collect.rgw_user_ratelimit":
						t.Fatalf("dependent read after invalid identity: %+v", call)
					}
				}
			}
		}
	}
}
