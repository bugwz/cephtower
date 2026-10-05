package ceph

import (
	"context"
	"encoding/json"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestMonitorQuorumStatusSummary(t *testing.T) {
	for _, raw := range []string{
		`{"quorum_names":["ceph-node-1"],"quorum_leader_name":"ceph-node-1","election_epoch":18446744073709551615,"quorum_age":0}`,
		`{"quorum_names":[],"quorum_leader_name":""}`,
		`{"quorum_names":[]}`,
	} {
		t.Run(raw, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.quorum": []byte(raw)}}}
			rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				if row.Kind != "mon_status" {
					continue
				}
				data, err := json.Marshal(row.Payload)
				if err != nil {
					t.Fatal(err)
				}
				var payload map[string]any
				if err := json.Unmarshal(data, &payload); err != nil {
					t.Fatal(err)
				}
				members, ok := payload["quorum_names"].([]any)
				if !ok {
					t.Fatalf("missing member array: %s", data)
				}
				if len(members) > 0 {
					if payload["quorum_leader_name"] != "ceph-node-1" || payload["election_epoch"] != "18446744073709551615" || payload["quorum_age"] != "0" {
						t.Fatalf("lost native fields: %s", data)
					}
				} else {
					if payload["election_epoch"] != nil || payload["quorum_age"] != nil {
						t.Fatalf("invented counters: %s", data)
					}
					var native map[string]any
					_ = json.Unmarshal([]byte(raw), &native)
					if payload["quorum_leader_name"] != native["quorum_leader_name"] {
						t.Fatalf("lost absent/empty leader: %s", data)
					}
				}
				return
			}
			t.Fatal("missing mon_status")
		})
	}
	for _, field := range []string{"election_epoch", "quorum_age"} {
		for _, value := range []string{`-1`, `1.5`, `"12"`, `18446744073709551616`} {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.quorum": []byte(`{"quorum_names":[],"` + field + `":` + value + `}`)}}}
			if rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology"); err == nil || len(rows) != 0 {
				t.Fatalf("accepted invalid %s %s", field, value)
			}
		}
	}
}

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
