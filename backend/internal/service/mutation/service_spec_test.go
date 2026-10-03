package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestServiceUpdateMergesExportedSpec(t *testing.T) {
	original := `[{"service_type":"rgw","service_id":"realm.zone","service_name":"rgw.realm.zone","placement":{"count":2},"networks":["10.0.0.0/24"],"unmanaged":true,"spec":{"ssl":true,"rgw_frontend_port":8443,"future_counter":18446744073709551615}}]`
	patch := []byte(`{"service_type":"rgw","service_id":"realm.zone","placement":{"count":3}}`)
	merged, err := mergeServiceSpec([]byte(original), patch, "rgw.realm.zone")
	if err != nil {
		t.Fatal(err)
	}
	var before []map[string]json.RawMessage
	var after map[string]json.RawMessage
	json.Unmarshal([]byte(original), &before)
	json.Unmarshal(merged, &after)
	for _, key := range []string{"spec", "networks", "unmanaged", "service_id", "service_name", "service_type"} {
		if string(before[0][key]) != string(after[key]) {
			t.Fatalf("changed %s: %s", key, merged)
		}
	}
	if string(after["placement"]) != `{"count":3}` {
		t.Fatal(string(merged))
	}
	for _, bad := range []string{`null`, `[]`, `[null]`, `[{},{}]`, `[{}]`, `[{"service_name":"rgw.other","service_type":"rgw","service_id":"other"}]`, original + ` {}`} {
		if _, err := mergeServiceSpec([]byte(bad), patch, "rgw.realm.zone"); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	s, _, id := newCephUserService(t)
	e := &directoryRenameExecutor{outputs: map[string]string{"service.update.pre_check": original}}
	s.executor = e
	_, err = s.Execute(context.Background(), Request{ClusterID: id, Action: "service.update", ResourceKey: "service/rgw.realm.zone", Parameters: map[string]any{"service_type": "rgw", "placement": map[string]any{"count": 3}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(e.specs) != 3 || e.specs[0].Mutating || !e.specs[1].Mutating || e.specs[2].Mutating {
		t.Fatalf("%+v", e.specs)
	}
	if !reflect.DeepEqual(e.specs[0].Args, []string{"orch", "ls", "--service-name", "rgw.realm.zone", "--export", "--format", "json"}) || string(e.specs[1].Stdin) != string(merged) {
		t.Fatalf("unexpected chain: %+v", e.specs)
	}
	e.specs = nil
	e.outputs["service.update.pre_check"] = `[]`
	if _, err = s.Execute(context.Background(), Request{ClusterID: id, Action: "service.update", ResourceKey: "service/rgw.realm.zone", Parameters: map[string]any{"service_type": "rgw"}}); err == nil || len(e.specs) != 1 {
		t.Fatalf("missing service written: %v %+v", err, e.specs)
	}
}
