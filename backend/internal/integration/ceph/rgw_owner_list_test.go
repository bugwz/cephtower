package ceph

import (
	"context"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWOwnerListsRequireCompleteNativeArrays(t *testing.T) {
	for _, noun := range []string{"user", "account"} {
		kind := "rgw_" + noun
		for _, raw := range []string{`[]`, `{}`, `null`, `false`, `{"accounts":[]}`, `{"keys":[],"truncated":false}`, `{"keys":["owner"],"truncated":true}`, `[null]`, `["owner",1]`, `[""]`, `["owner","owner"]`} {
			t.Run(noun+raw, func(t *testing.T) {
				trace := &collectionTrace{unavailable: map[string]struct{}{}}
				ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
				calls := []executor.CommandSpec{}
				p := NativeProvider{Executor: recordingExecutor{calls: &calls, base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect." + kind: []byte(raw)}}}}
				for _, row := range p.collectRGWOptional(ctx, ClusterAccess{}, time.Now()) {
					if row.Kind == kind {
						t.Fatal("unexpected owner observation")
					}
				}
				_, unavailable := trace.unavailable[kind]
				if unavailable != (raw != `[]`) {
					t.Fatalf("wrong list availability: %+v", trace.unavailable)
				}
				found := false
				for _, call := range calls {
					if call.ID == "collect."+kind+"_detail" {
						t.Fatal("detail requested for invalid list")
					}
					if call.ID == "collect."+kind {
						found = true
						if call.Mutating || !reflect.DeepEqual(call.Args, []string{noun, "list", "--format", "json"}) {
							t.Fatalf("wrong list command: %+v", call)
						}
					}
				}
				if !found {
					t.Fatal("missing list command")
				}
			})
		}
	}
}
