package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestResolveRealmSetupEndpoints(t *testing.T) {
	for _, tc := range []struct{ addr, input, want string }{
		{"192.0.2.10", "https://node-a:8443/node-a/%2F", "https://192.0.2.10:8443/node-a/%2F"},
		{"2001:db8::1", "http://node-a:8080/path", "http://[2001:db8::1]:8080/path"},
		{"[2001:db8::1]", "https://NODE-A/path", "https://[2001:db8::1]/path"},
		{"rgw.internal.example", "https://node-a/path", "https://rgw.internal.example/path"},
		{"192.0.2.10", "https://unknown.example/node-a", "https://unknown.example/node-a"},
	} {
		p := setupParameters()
		p["zone_endpoints"] = []string{tc.input}
		raw, _ := json.Marshal([]map[string]string{{"hostname": "node-a", "addr": tc.addr}})
		got, err := resolveRealmSetupEndpoints(raw, p)
		if err != nil || !reflect.DeepEqual(got["zone_endpoints"], []string{tc.want}) {
			t.Fatalf("unexpected mapping for %s: %v", tc.input, err)
		}
		if !reflect.DeepEqual(p["zone_endpoints"], []string{tc.input}) {
			t.Fatal("request snapshot was modified")
		}
	}
	for _, raw := range []string{`null`, `{}`, `[{}]`, `[{"hostname":"node-a","addr":""}]`, `[{"hostname":"node-a","addr":"user:secret@host"}]`, `[{"hostname":"node-a","addr":"host:8080"}]`, `[{"hostname":"node-a","addr":"192.0.2.1/path"}]`, `[{"hostname":"node-a","addr":"192.0.2.1"},{"hostname":"NODE-A","addr":"192.0.2.2"}]`, `[{"hostname":true,"addr":"192.0.2.1"}]`, `[] {}`} {
		if _, err := resolveRealmSetupEndpoints([]byte(raw), setupParameters()); err == nil {
			t.Fatal("malformed host inventory accepted")
		}
	}
	p := setupParameters()
	p["zone_endpoints"] = []string{"http://node-a:80", "http://node-b:80"}
	if _, err := resolveRealmSetupEndpoints([]byte(`[{"hostname":"node-a","addr":"192.0.2.1"},{"hostname":"node-b","addr":"192.0.2.1"}]`), p); err == nil {
		t.Fatal("collapsed duplicate endpoints accepted")
	}
}

func TestRealmSetupUsesResolvedEndpoints(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := setupParameters()
	p["zonegroup_endpoints"] = []string{"https://node-a:8443"}
	p["zone_endpoints"] = []string{"https://node-a:8443"}
	r := setupResponses(t)
	r["hosts"] = `[{"hostname":"node-a","addr":"2001:db8::1"}]`
	for _, stage := range []string{"initial_commit", "commit", "period_check"} {
		r[stage] = strings.ReplaceAll(strings.ReplaceAll(r[stage], "https://group.example", "https://[2001:db8::1]:8443"), "https://zone.example", "https://[2001:db8::1]:8443")
	}
	e := &realmSetupExecutor{responses: r}
	s.executor = e
	result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.setup", Parameters: p})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Details.(map[string]any)["zone_endpoints"], []string{"https://[2001:db8::1]:8443"}) {
		t.Fatal("effective endpoints missing from result")
	}
	for _, c := range e.specs {
		if c.ID == "rgw_realm.setup.hosts" && (c.Mutating || c.Binary != executor.BinaryCeph || !reflect.DeepEqual(c.Args, []string{"orch", "host", "ls", "--format", "json"})) {
			t.Fatal("wrong host query")
		}
		if c.ID == "rgw_realm.setup.group_create" || c.ID == "rgw_realm.setup.zone_create" {
			found := false
			for i, arg := range c.Args {
				if arg == "--endpoints" && i+1 < len(c.Args) {
					found = c.Args[i+1] == "https://[2001:db8::1]:8443"
				}
			}
			if !found {
				t.Fatal("unresolved endpoint used for mutation")
			}
		}
	}
	if !reflect.DeepEqual(p["zone_endpoints"], []string{"https://node-a:8443"}) {
		t.Fatal("original request was changed")
	}
}

func TestRealmSetupInvalidHostInventoryDoesNotWrite(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	r := setupResponses(t)
	r["hosts"] = `[{"hostname":"node-a","addr":"invalid/addr"}]`
	e := &realmSetupExecutor{responses: r}
	s.executor = e
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.setup", Parameters: setupParameters()}); err == nil {
		t.Fatal("invalid inventory accepted")
	}
	for _, c := range e.specs {
		if c.Mutating {
			t.Fatal("wrote before host mapping was verified")
		}
	}
}
