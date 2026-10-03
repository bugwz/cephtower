package ceph

import (
	"encoding/json"
	"testing"
)

func TestOverviewPGStatesRequireCompleteCounts(t *testing.T) {
	for _, data := range []string{
		`{}`, `{"num_pgs":0}`, `{"num_pgs":0,"pgs_by_state":null}`,
		`{"pgs_by_state":[]}`,
		`{"num_pgs":1,"pgs_by_state":[{"state_name":"active"}]}`,
		`{"num_pgs":0,"pgs_by_state":[{"state_name":"active","count":null}]}`,
		`{"num_pgs":1,"pgs_by_state":[{"state_name":"","count":1}]}`,
		`{"num_pgs":2,"pgs_by_state":[{"state_name":"active","count":1},{"state_name":"active","count":1}]}`,
		`{"num_pgs":2,"pgs_by_state":[{"state_name":"active","count":1}]}`,
		`{"num_pgs":0,"pgs_by_state":[{"state_name":"active","count":18446744073709551615},{"state_name":"clean","count":1}]}`,
	} {
		var status statusWire
		if err := json.Unmarshal([]byte(`{"pgmap":`+data+`}`), &status); err != nil {
			t.Fatal(err)
		}
		if got := validatedOverviewPGStates(status); got != nil {
			t.Fatalf("invalid counts accepted: %s %#v", data, got)
		}
	}
	for _, data := range []string{
		`{"num_pgs":0,"pgs_by_state":[]}`,
		`{"num_pgs":0,"pgs_by_state":[{"state_name":"active","count":0}]}`,
		`{"num_pgs":3,"pgs_by_state":[{"state_name":"active+clean","count":2},{"state_name":"new_state","count":1}]}`,
	} {
		var status statusWire
		if err := json.Unmarshal([]byte(`{"pgmap":`+data+`}`), &status); err != nil {
			t.Fatal(err)
		}
		if got := validatedOverviewPGStates(status); got == nil {
			t.Fatalf("valid counts rejected: %s", data)
		}
	}
	if got := overviewScrubStatus(nil, nil, true); got != "" {
		t.Fatalf("missing PG data reported %q", got)
	}
	if got := overviewScrubStatus(nil, []string{"noscrub"}, true); got != "disabled" {
		t.Fatalf("known flag lost: %q", got)
	}
}
