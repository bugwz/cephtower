package mutation

import (
	"encoding/json"
	"testing"
)

func TestServiceUpdatePreservesIdentity(t *testing.T) {
	for _, tc := range []struct{ name, kind, id string }{
		{"mon", "mon", ""}, {"mgr", "mgr", ""}, {"rgw.foo", "rgw", "foo"}, {"rgw.realm.zone", "rgw", "realm.zone"}, {"mds.fs", "mds", "fs"},
	} {
		for _, explicit := range []bool{false, true} {
			p := map[string]any{"service_type": tc.kind}
			if explicit {
				p["service_id"] = tc.id
			}
			cmd, err := build(Request{Action: "service.update", ResourceKey: "service/" + tc.name}, p)
			if err != nil {
				t.Fatal(err)
			}
			var spec map[string]any
			if err := json.Unmarshal(cmd.stdin, &spec); err != nil {
				t.Fatal(err)
			}
			if spec["service_type"] != tc.kind {
				t.Fatal(spec)
			}
			if tc.id == "" {
				if _, ok := spec["service_id"]; ok {
					t.Fatal(spec)
				}
			} else if spec["service_id"] != tc.id {
				t.Fatal(spec)
			}
		}
	}
	for _, p := range []map[string]any{{"service_type": "nfs"}, {"service_type": "rgw", "service_id": "other"}, {"service_type": "rgw", "service_id": ""}, {"service_type": "rgw", "service_id": "rgw.foo"}} {
		if _, err := build(Request{Action: "service.update", ResourceKey: "service/rgw.foo"}, p); err == nil {
			t.Fatalf("accepted changed target: %v", p)
		}
	}
}

func TestServiceCreationNativeIdentityRequirements(t *testing.T) {
	for _, kind := range []string{"mds", "rgw", "nfs", "smb", "mon", "mgr", "rbd-mirror", "cephfs-mirror", "prometheus", "alertmanager", "grafana", "loki", "promtail", "node-exporter", "crash"} {
		requires := kind == "mds" || kind == "rgw" || kind == "nfs" || kind == "smb"
		for _, id := range []string{"", "realm.zone-1_a", "bad id", "bad/id", "bad:id"} {
			_, err := build(Request{Action: "service.create"}, map[string]any{"service_type": kind, "service_id": id})
			valid := (requires && id == "realm.zone-1_a") || (!requires && id == "")
			if (err == nil) != valid {
				t.Fatalf("%s id %q: %v", kind, id, err)
			}
		}
	}
}
