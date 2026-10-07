package ceph

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestInventoryJSONDocumentBoundary(t *testing.T) {
	for _, raw := range []string{`{"count":18446744073709551615}`, " {\"count\":18446744073709551615}\n\t", `{"count":1} {}`, `{"count":1} null`, `{"count":1} SECRET-DIAGNOSTIC`, `{"count":`, "", " \n", `{"count":01}`} {
		t.Run(raw, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rbd_mirror_service_status": []byte(raw)}}}
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			var result map[string]any
			err := p.runBinaryInto(ctx, ClusterAccess{}, executor.BinaryCeph, "collect.rbd_mirror_service_status", []string{"service", "status", "--format", "json"}, &result)
			if json.Valid([]byte(raw)) {
				if err != nil || result["count"] != json.Number("18446744073709551615") || len(trace.unavailable) != 0 {
					t.Fatalf("result=%#v err=%v trace=%v", result, err, trace.unavailable)
				}
			} else {
				if err == nil || result != nil || strings.Contains(err.Error(), "SECRET") {
					t.Fatalf("invalid response accepted or leaked: result=%#v err=%v", result, err)
				}
				if _, ok := trace.unavailable["rbd_mirroring"]; !ok {
					t.Fatal("missing failure classification")
				}
			}
		})
	}
}
