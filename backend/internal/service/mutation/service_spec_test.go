package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
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

func TestServiceCreationNeverOverwritesExistingSpec(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, output := range []string{"Scheduled rgw.a update...", "Skipped rgw.a service spec. To change rgw.a spec omit --no-overwrite flag", "", "Scheduled rgw.other update...", "Failed to apply spec"} {
		e := &directoryRenameExecutor{outputs: map[string]string{"service.create": output}}
		s.executor = e
		_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "service.create", ResourceKey: "service/rgw.a", Parameters: map[string]any{"service_type": "rgw", "service_id": "a"}})
		if output == "Scheduled rgw.a update..." {
			if err != nil || len(e.specs) != 2 {
				t.Fatalf("%v %+v", err, e.specs)
			}
		} else {
			var ae *cephdomain.ActionError
			if !errors.As(err, &ae) || ae.Code != "post_check_failed" || ae.Retryable || len(e.specs) != 1 {
				t.Fatalf("unconfirmed create accepted: %s %v", output, err)
			}
		}
		if !reflect.DeepEqual(e.specs[0].Args, []string{"orch", "apply", "-i", "-", "--no-overwrite"}) || !e.specs[0].Mutating {
			t.Fatalf("unsafe create: %+v", e.specs[0])
		}
	}
	cmd, err := build(Request{Action: "service.update", ResourceKey: "service/rgw.a"}, map[string]any{"service_type": "rgw"})
	if err != nil || !reflect.DeepEqual(cmd.args, []string{"orch", "apply", "-i", "-"}) {
		t.Fatalf("update changed: %+v %v", cmd, err)
	}
}

func TestServiceManagementModeUpdatesPreserveSpec(t *testing.T) {
	for _, unmanaged := range []bool{true, false} {
		cmd, err := build(Request{Action: "service.update", ResourceKey: "service/rgw.a"}, map[string]any{"service_type": "rgw", "unmanaged": unmanaged})
		if err != nil {
			t.Fatal(err)
		}
		merged, err := mergeServiceSpec([]byte(`[{"service_name":"rgw.a","service_type":"rgw","service_id":"a","unmanaged":true,"placement":{"count":2},"spec":{"ssl":true}}]`), cmd.stdin, "rgw.a")
		if err != nil {
			t.Fatal(err)
		}
		var value map[string]json.RawMessage
		if err := json.Unmarshal(merged, &value); err != nil {
			t.Fatal(err)
		}
		want, _ := json.Marshal(unmanaged)
		if string(value["unmanaged"]) != string(want) || string(value["spec"]) != `{"ssl":true}` || string(value["placement"]) != `{"count":2}` {
			t.Fatalf("unexpected merged spec: %s", merged)
		}
	}
	for _, value := range []any{nil, "false", 0} {
		if _, err := build(Request{Action: "service.create", ResourceKey: "service"}, map[string]any{"service_type": "mgr", "unmanaged": value}); err == nil {
			t.Fatalf("invalid management mode accepted: %#v", value)
		}
	}
}

func TestServiceNetworkUpdates(t *testing.T) {
	for _, networks := range [][]string{{}, {"192.0.2.0/24", "2001:db8::/64"}} {
		for _, action := range []string{"service.create", "service.update"} {
			cmd, err := build(Request{Action: action, ResourceKey: "service/mgr"}, map[string]any{"service_type": "mgr", "networks": networks})
			if err != nil {
				t.Fatal(err)
			}
			data := cmd.stdin
			if action == "service.update" {
				data, err = mergeServiceSpec([]byte(`[{"service_name":"mgr","service_type":"mgr","networks":["10.0.0.0/8"],"spec":{"custom":true}}]`), data, "mgr")
				if err != nil {
					t.Fatal(err)
				}
			}
			var parsed map[string]json.RawMessage
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(networks)
			if string(parsed["networks"]) != string(want) {
				t.Fatalf("networks lost: %s", data)
			}
			if action == "service.update" && string(parsed["spec"]) != `{"custom":true}` {
				t.Fatalf("spec lost: %s", data)
			}
		}
	}
	for _, networks := range []any{nil, "192.0.2.0/24", []any{false}, []string{" "}} {
		if _, err := build(Request{Action: "service.create"}, map[string]any{"service_type": "mgr", "networks": networks}); err == nil {
			t.Fatalf("accepted invalid networks: %#v", networks)
		}
	}
}

func TestServiceDeletionRequiresNativeConfirmationAndAbsence(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, tc := range []struct {
		ack, inventory string
		ok             bool
	}{
		{"Removed service rgw.a", "[]", true},
		{"Removed service rgw.a\n", "No services reported\n", true},
		{"Unable to remove rgw.a service.", "[]", false},
		{"Invalid service 'rgw.a'.", "[]", false},
		{"Removed service rgw.b", "[]", false},
		{"Removed service rgw.a", `[{"service_name":"rgw.a"}]`, false},
		{"Removed service rgw.a", "null", false},
		{"Removed service rgw.a", "[] {}", false},
		{"Removed service rgw.a", "", false},
	} {
		e := &directoryRenameExecutor{outputs: map[string]string{"service.delete": tc.ack, "service.delete.post_check": tc.inventory}}
		s.executor = e
		_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "service.delete", ResourceKey: "service/rgw.a"})
		if tc.ok {
			if err != nil {
				t.Fatal(err)
			}
		} else {
			var ae *cephdomain.ActionError
			if !errors.As(err, &ae) || ae.Code != "post_check_failed" || ae.Retryable {
				t.Fatalf("unconfirmed removal accepted: %+v %v", tc, err)
			}
		}
		if len(e.specs) > 1 && (!reflect.DeepEqual(e.specs[1].Args, []string{"orch", "ls", "--service-name", "rgw.a", "--export", "--format", "json"}) || e.specs[1].Mutating) {
			t.Fatalf("incorrect verification: %+v", e.specs)
		}
	}
}
