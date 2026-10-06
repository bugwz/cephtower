package ceph

import (
	"context"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestOverviewManagerCountsRetainUnknown(t *testing.T) {
	for _, tc := range []struct {
		raw             string
		active, standby *int
	}{
		{`{}`, nil, nil},
		{`{"available":null,"num_standbys":null}`, nil, nil},
		{`{"available":false,"num_standbys":0}`, intPointer(0), intPointer(0)},
		{`{"available":true,"num_standbys":2}`, intPointer(1), intPointer(2)},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.status": []byte(`{"fsid":"test","health":{"status":"HEALTH_OK"},"mgrmap":` + tc.raw + `}`)}}}
			rows, err := provider.Collect(context.Background(), ClusterAccess{}, "fast")
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				if row.Kind != "overview" {
					continue
				}
				got := row.Payload.(cephdomain.Overview).Services["mgr"]
				if !reflect.DeepEqual(got.Active, tc.active) || !reflect.DeepEqual(got.Standby, tc.standby) {
					t.Fatalf("unexpected manager counts: %+v", got)
				}
				return
			}
			t.Fatal("overview missing")
		})
	}
}
