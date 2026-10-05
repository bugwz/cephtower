package ceph

import (
	"context"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestMonitorQuorumRequiresExplicitNativeNames(t *testing.T) {
	for _, raw := range []string{
		`{"quorum_names":[]}`, `{"quorum_names":["ceph-node-1"]}`,
		`{}`, `null`, `{"quorum_names":null}`, `{"quorum_names":{}}`,
		`{"quorum_names":[null]}`, `{"quorum_names":[123]}`, `{"quorum_names":[false]}`,
		`{"quorum_names":[""]}`, `{"quorum_names":[" ceph-node-1"]}`,
		`{"quorum_names":["ceph-node-1","ceph-node-1"]}`,
	} {
		t.Run(raw, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.quorum": []byte(raw)}}}
			rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
			valid := raw == `{"quorum_names":[]}` || raw == `{"quorum_names":["ceph-node-1"]}`
			if (err == nil) != valid {
				t.Fatalf("unexpected result: %v", err)
			}
			if !valid {
				if len(rows) != 0 {
					t.Fatal("published invented quorum state")
				}
				return
			}
			found := false
			for _, row := range rows {
				if row.Kind == "mon" && row.Name == "ceph-node-1" {
					found = true
					if row.Payload.(cephdomain.Monitor).InQuorum != (raw != `{"quorum_names":[]}`) {
						t.Fatal("incorrect membership")
					}
				}
			}
			if !found {
				t.Fatal("missing monitor")
			}
		})
	}
}
