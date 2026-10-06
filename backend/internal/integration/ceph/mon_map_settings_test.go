package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestMonitorMapSettings(t *testing.T) {
	for _, raw := range []string{`{}`, `{"stretch_mode":null}`, `{"min_mon_release":0,"min_mon_release_name":"unknown","election_strategy":1,"stretch_mode":false,"tiebreaker_mon":"","disallowed_leaders":"[]","removed_ranks":"[]"}`, `{"min_mon_release":19,"min_mon_release_name":"squid","election_strategy":3,"stretch_mode":true,"tiebreaker_mon":"c","disallowed_leaders":"[a,b]","removed_ranks":"[1,2]"}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.quorum": []byte(`{"quorum_names":[],"monmap":` + raw + `}`)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		var want map[string]any
		if err := json.Unmarshal([]byte(raw), &want); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "mon_status" {
				continue
			}
			found = true
			data, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"min_mon_release", "min_mon_release_name", "election_strategy", "stretch_mode", "tiebreaker_mon", "disallowed_leaders", "removed_ranks"} {
				if !reflect.DeepEqual(got[field], want[field]) {
					t.Fatalf("%s got=%v want=%v", field, got[field], want[field])
				}
			}
		}
		if !found {
			t.Fatal("missing mon status")
		}
	}
}
