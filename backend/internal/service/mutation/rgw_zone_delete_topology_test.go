package mutation

import (
	"context"
	"testing"
)

func TestZoneDeleteRejectsInvalidRealmTopologyBeforeWriting(t *testing.T) {
	service, _, cluster := newCephUserService(t)
	for _, raw := range []string{
		`{"id":"other","name":"other","realm_id":"r","is_master":false,"master_zone":"missing","zones":[{"id":"remote"}]}`,
		`{"id":"other","name":"other","realm_id":"r","is_master":false,"master_zone":"","zones":[]}`,
		`{"id":"other","name":"other","realm_id":"r","is_master":true,"master_zone":"remote","zones":[{"id":"remote"}]}`,
		`{"id":"other","name":"other","realm_id":"r","master_zone":"remote","zones":[{"id":"remote"}]}`,
	} {
		runner := zoneDeleteFixture("r")
		runner.outputs["groups_before"] = `{"zonegroups":["group","other"],"default_info":"g"}`
		runner.outputs["group_before_1"] = raw
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zone.delete", Parameters: map[string]any{"zone_id": "z", "name": "secondary", "realm_id": "r", "confirm_delete": true}})
		if err == nil {
			t.Fatal("invalid Realm topology accepted")
		}
		for _, call := range runner.calls {
			if call.Mutating {
				t.Fatal("deleted before discovering invalid Realm group")
			}
		}
		if runner.calls[len(runner.calls)-1].ID != "rgw_zone.delete.period_before" {
			t.Fatal("did not exercise Realm preflight")
		}
	}
}

func TestZoneDeleteRealmTopology(t *testing.T) {
	primary := periodDocument([]byte(`{"id":"g","realm_id":"r","is_master":true,"master_zone":"a","zones":[{"id":"a"}]}`))
	secondary := periodDocument([]byte(`{"id":"s","realm_id":"r","is_master":false,"master_zone":"b","zones":[{"id":"b"}]}`))
	published := map[string]any{"master_zonegroup": "g", "master_zone": "a"}
	if !zoneDeleteRealmTopology([]map[string]any{primary, secondary}, "r", published) {
		t.Fatal("valid topology rejected")
	}
	if !zoneDeleteRealmTopology([]map[string]any{primary, {"realm_id": "outside"}}, "r", published) {
		t.Fatal("unrelated Realm constrained")
	}
	for _, groups := range [][]map[string]any{nil, {secondary}, {primary, primary}, {primary, {"id": "s", "realm_id": "r", "is_master": false, "master_zone": "b", "zones": []any{map[string]any{"id": "b"}, map[string]any{"id": "b"}}}}} {
		if zoneDeleteRealmTopology(groups, "r", published) {
			t.Fatal("invalid topology accepted")
		}
	}
	for _, p := range []map[string]any{nil, {"master_zonegroup": "s", "master_zone": "b"}, {"master_zonegroup": "g", "master_zone": "changed"}} {
		if zoneDeleteRealmTopology([]map[string]any{primary, secondary}, "r", p) {
			t.Fatal("implicit primary migration accepted")
		}
	}
}
