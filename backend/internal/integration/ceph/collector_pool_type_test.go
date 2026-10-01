package ceph

import (
	"context"
	"fmt"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestPoolTypeMapping(t *testing.T) {
	for _, tc := range []struct{ field, want string }{
		{`,"type":1`, "replicated"},
		{`,"type":3`, "erasure"},
		{``, "unknown"},
		{`,"type":null`, "unknown"},
		{`,"type":0`, "unknown"},
		{`,"type":99`, "unknown"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.pool": []byte(fmt.Sprintf(`[{"pool":7,"pool_name":"type-pool"%s}]`, tc.field)),
			}}}
			rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				if row.Kind == "pool" {
					if got := row.Payload.(cephdomain.Pool).Type; got != tc.want {
						t.Fatalf("type = %q, want %q", got, tc.want)
					}
					return
				}
			}
			t.Fatal("missing pool observation")
		})
	}
}
