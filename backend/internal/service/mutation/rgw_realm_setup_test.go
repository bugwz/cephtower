package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type realmSetupExecutor struct {
	responses map[string]string
	fail      string
	specs     []executor.CommandSpec
}

func (e *realmSetupExecutor) Run(_ context.Context, _ executor.ClusterAccess, c executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, c)
	stage := strings.TrimPrefix(c.ID, "rgw_realm.setup.")
	if stage == e.fail {
		return executor.CommandResult{Stderr: []byte("generated-secret")}, errors.New("generated-secret")
	}
	return executor.CommandResult{Stdout: []byte(e.responses[stage])}, nil
}
func setupParameters() map[string]any {
	return map[string]any{"name": "realm", "zonegroup": "group", "zone": "primary", "username": "sys", "zonegroup_endpoints": []string{"https://group.example"}, "zone_endpoints": []string{"https://zone.example"}, "expected_services": []string{"rgw.gateway"}, "confirm_setup": true}
}
func setupResponses(t *testing.T) map[string]string {
	t.Helper()
	period := `{"id":"period","epoch":2,"realm_id":"r","master_zonegroup":"g","master_zone":"z","period_map":{"zonegroups":[{"id":"g","name":"group","master_zone":"z","endpoints":["https://group.example"],"zones":[{"id":"z","name":"primary","tier_type":"","endpoints":["https://zone.example"]}]}]}}`
	zone := `{"id":"z","name":"primary","realm_id":"r","system_key":{"access_key":"generated-access","secret_key":"generated-secret"}}`
	user := `{"user_id":"sys","system":true,"keys":[{"user":"sys","access_key":"generated-access","secret_key":"generated-secret"}]}`
	return map[string]string{"realm_absence": `{"realms":[]}`, "zonegroup_absence": `{"zonegroups":[]}`, "zone_absence": `{"zones":[]}`,
		"daemons_before":  `[{"daemon_id":"gateway.host.id","daemon_type":"rgw","service_name":"rgw.gateway","hostname":"host","status":1,"started":"2026-01-01T00:00:00Z","last_refresh":"2026-01-01T00:01:00Z"}]`,
		"daemons_ready":   `[{"daemon_id":"gateway.host.id","daemon_type":"rgw","service_name":"rgw.gateway","hostname":"host","status":1,"started":"2026-01-01T00:02:00Z","last_refresh":"2026-01-01T00:03:00Z"}]`,
		"hosts":           `[]`,
		"services_before": `[{"service_name":"rgw.gateway","service_type":"rgw"}]`, "services_check": `[{"service_name":"rgw.gateway","service_type":"rgw"}]`,
		"realm_create": `{"id":"r","name":"realm"}`, "group_create": `{"id":"g","name":"group","realm_id":"r"}`, "zone_create": `{"id":"z","name":"primary","realm_id":"r"}`,
		"initial_commit": period, "user_absence": `[]`, "user_create": user, "zone_credentials": zone, "commit": period,
		"realm_check": `{"id":"r","name":"realm","current_period":"period"}`, "period_check": period,
		"realm_default": `{"default_info":"r"}`, "zonegroup_default": `{"default_info":"g"}`, "zone_default": `{"default_info":"z"}`, "zone_check": zone, "user_check": user,
	}
}
func TestRealmSetupValidatesBeforeWrites(t *testing.T) {
	for _, endpoint := range []string{"https://host:0", "https://host:65536"} {
		p := setupParameters()
		p["zone_endpoints"] = []string{endpoint}
		if _, err := buildRealmSetup(p); err == nil {
			t.Fatal("invalid endpoint port accepted")
		}
	}
	for _, bad := range []map[string]any{{"name": "-bad"}, {"confirm_setup": false}, {"tier_type": "wrong"}, {"expected_services": nil}, {"expected_services": []string{"rgw.one", "rgw.one"}}, {"zone_endpoints": []string{}}, {"zone_endpoints": []string{"file:///secret"}}, {"zone_endpoints": []string{"https://u:secret@host"}}, {"zone_endpoints": []string{"https://host/?a=b"}}, {"zone_endpoints": []string{"https://a.example", "https://a.example"}}} {
		p := setupParameters()
		for k, v := range bad {
			p[k] = v
		}
		if _, err := buildRealmSetup(p); err == nil {
			t.Fatal("invalid setup accepted")
		}
	}
}
func TestRealmSetupNativeSequence(t *testing.T) {
	for _, archive := range []bool{false, true} {
		s, _, cluster := newCephUserService(t)
		responses := setupResponses(t)
		p := setupParameters()
		if archive {
			p["tier_type"] = "archive"
			for _, stage := range []string{"initial_commit", "commit", "period_check"} {
				responses[stage] = strings.ReplaceAll(responses[stage], `"tier_type":""`, `"tier_type":"archive"`)
			}
		}
		e := &realmSetupExecutor{responses: responses}
		s.executor = e
		result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.setup", Parameters: p})
		if err != nil {
			t.Fatal(err)
		}
		if result.Details.(map[string]any)["daemons_verified"] != true || result.Details.(map[string]any)["replication_verified"] != false {
			t.Fatal("incorrect verification result")
		}
		raw, _ := json.Marshal(result)
		if strings.Contains(string(raw), "generated-") {
			t.Fatal("secret leaked")
		}
		var stages []string
		for _, c := range e.specs {
			stage := strings.TrimPrefix(c.ID, "rgw_realm.setup.")
			stages = append(stages, stage)
			if stage == "daemons_before" || stage == "daemons_ready" {
				if c.Mutating || !reflect.DeepEqual(c.Args, []string{"orch", "ps", "--service-name", "rgw.gateway", "--daemon-type", "rgw", "--refresh", "--format", "json"}) {
					t.Fatal("unscoped or cached daemon query")
				}
			}
			for i, arg := range c.Args {
				if strings.HasPrefix(arg, "generated-") {
					if _, ok := c.SensitiveArgs[i]; !ok {
						t.Fatal("unmarked secret")
					}
				}
			}
			if stage == "realm_create" && !reflect.DeepEqual(c.Args, []string{"realm", "create", "--rgw-realm", "realm", "--default", "--format", "json"}) {
				t.Fatal("default realm missing")
			}
			if stage == "commit" && !reflect.DeepEqual(c.Args, []string{"period", "update", "--commit", "--realm-id", "r", "--zonegroup-id", "g", "--zone-id", "z", "--format", "json"}) {
				t.Fatal("unscoped commit")
			}
			if stage == "restart" && (!c.Mutating || !reflect.DeepEqual(c.Args, []string{"orch", "restart", "rgw.gateway"})) {
				t.Fatal("wrong restart")
			}
		}
		want := []string{"realm_absence", "zonegroup_absence", "zone_absence", "services_before", "hosts", "realm_create", "group_create", "zone_create", "initial_commit", "user_absence", "user_create", "zone_credentials", "commit", "realm_check", "period_check", "realm_default", "zonegroup_default", "zone_default", "zone_check", "user_check", "services_check", "daemons_before", "restart", "daemons_ready"}
		if !reflect.DeepEqual(stages, want) {
			t.Fatalf("unexpected sequence %v", stages)
		}
	}
}
func TestRealmSetupStopsAtEveryStage(t *testing.T) {
	for _, stage := range []string{"realm_absence", "zonegroup_absence", "zone_absence", "services_before", "hosts", "realm_create", "group_create", "zone_create", "initial_commit", "user_absence", "user_create", "zone_credentials", "commit", "realm_check", "period_check", "realm_default", "zonegroup_default", "zone_default", "zone_check", "user_check", "services_check", "daemons_before", "restart", "daemons_ready"} {
		t.Run(stage, func(t *testing.T) {
			s, _, cluster := newCephUserService(t)
			e := &realmSetupExecutor{responses: setupResponses(t), fail: stage}
			s.executor = e
			_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.setup", Parameters: setupParameters()})
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Retryable || strings.Contains(err.Error(), "generated-secret") {
				t.Fatal("unsafe failure")
			}
			if e.specs[len(e.specs)-1].ID != "rgw_realm.setup."+stage {
				t.Fatal("continued after failure")
			}
		})
	}
}
func TestRealmSetupRejectsChangedState(t *testing.T) {
	for stage, bad := range map[string]string{"realm_absence": `{"realms":["realm"]}`, "services_before": `[]`, "user_absence": `["sys"]`, "user_create": `{"user_id":"sys","system":false}`, "commit": `{"id":"period","epoch":2}`, "realm_default": `{"default_info":"other"}`, "zone_check": `{"id":"z","name":"primary","realm_id":"r"}`, "services_check": `[{"service_type":"rgw","service_name":"rgw.new"}]`} {
		s, _, cluster := newCephUserService(t)
		r := setupResponses(t)
		r[stage] = bad
		e := &realmSetupExecutor{responses: r}
		s.executor = e
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.setup", Parameters: setupParameters()}); err == nil {
			t.Fatal("changed state accepted", stage)
		}
		if e.specs[len(e.specs)-1].ID != "rgw_realm.setup."+stage {
			t.Fatal("continued after invalid state")
		}
	}
}
