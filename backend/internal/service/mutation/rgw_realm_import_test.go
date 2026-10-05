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

type realmImportExecutor struct {
	responses map[string]string
	fail      string
	specs     []executor.CommandSpec
}

func (e *realmImportExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	spec.Stdin = append([]byte(nil), spec.Stdin...)
	e.specs = append(e.specs, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_realm.import.")
	if stage == e.fail {
		return executor.CommandResult{Stderr: []byte("fixture-secret")}, errors.New("fixture-secret")
	}
	return executor.CommandResult{Stdout: []byte(e.responses[stage])}, nil
}
func realmImportResponses() map[string]string {
	return map[string]string{
		"system_user": `{"user_id":"sys","system":true,"keys":[{"access_key":"fixture-access","secret_key":"fixture-secret"}]}`,
		"zones":       `{"zones":[]}`, "realms": `{"realms":[]}`, "service_absence": `[]`,
		"realm_post_check":   `{"id":"id","name":"realm","current_period":"period"}`,
		"zone_post_check":    `{"id":"zone-id","name":"secondary","realm_id":"id"}`,
		"period_post_check":  `{"id":"period","realm_id":"id","master_zonegroup":"zg","period_map":{"zonegroups":[{"id":"zg","name":"group","master_zone":"master","zones":[{"id":"zone-id","name":"secondary"}]}]}}`,
		"service_post_check": `[{"service_name":"rgw.realm.secondary","service_id":"realm.secondary","service_type":"rgw","spec":{"rgw_realm":"realm","rgw_zone":"secondary","rgw_zonegroup":"group","rgw_frontend_port":80}}]`,
		"deployment_service": `[{"service_name":"rgw.realm.secondary","service_id":"realm.secondary","service_type":"rgw","spec":{"rgw_realm":"realm","rgw_zone":"secondary","rgw_zonegroup":"group","rgw_frontend_port":80},"status":{"size":1,"running":1,"last_refresh":"2026-01-01T00:03:00Z"}}]`,
		"deployment_daemons": `[{"daemon_id":"realm.secondary.host.id","daemon_type":"rgw","service_name":"rgw.realm.secondary","hostname":"host","status":1,"started":"2026-01-01T00:02:00Z","last_refresh":"2026-01-01T00:03:00Z"}]`,
	}
}
func realmImportParameters(t *testing.T) map[string]any {
	return map[string]any{"realm_token": realmTokenFixture(t, "id", "realm"), "name": "secondary", "port": 80, "placement": map[string]any{}, "confirm_import": true}
}
func TestRealmImportCommand(t *testing.T) {
	p := realmImportParameters(t)
	c, err := buildRealmImport(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.binary != executor.BinaryCeph || !reflect.DeepEqual(c.args, []string{"rgw", "zone", "create", "-i", "-"}) {
		t.Fatal("wrong native command")
	}
	var doc map[string]any
	if json.Unmarshal(c.stdin, &doc) != nil || doc["rgw_realm_token"] != p["realm_token"] || doc["service_id"] != "realm.secondary" || doc["rgw_zone"] != "secondary" {
		t.Fatal("invalid stdin spec")
	}
	for _, bad := range []map[string]any{{"port": 0}, {"port": 65536}, {"port": 1.5}, {"port": "80"}, {"name": "-bad"}, {"confirm_import": false}, {"realm_token": "secret"}, {"placement": nil}, {"placement": map[string]any{"count": 0}}, {"placement": map[string]any{"label": "a", "hosts": []string{"h"}}}, {"placement": map[string]any{"tier_type": "archive"}}} {
		p := realmImportParameters(t)
		for k, v := range bad {
			p[k] = v
		}
		if _, err := buildRealmImport(p); err == nil {
			t.Fatal("invalid import accepted")
		}
	}
}
func TestRealmImportExecution(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	e := &realmImportExecutor{responses: realmImportResponses()}
	s.executor = e
	result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: realmImportParameters(t)})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(result)
	if result.Details.(map[string]any)["daemons_verified"] != true || result.Details.(map[string]any)["replication_verified"] != false {
		t.Fatal("incorrect readiness result")
	}
	if strings.Contains(string(encoded), "fixture-secret") || strings.Contains(string(encoded), "realm_token") {
		t.Fatal("result leaks credentials")
	}
	writes := 0
	for _, spec := range e.specs {
		if strings.HasSuffix(spec.ID, ".deployment_service") || strings.HasSuffix(spec.ID, ".deployment_daemons") {
			command := "ls"
			if strings.HasSuffix(spec.ID, ".deployment_daemons") {
				command = "ps"
			}
			if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"orch", command, "--service-name", "rgw.realm.secondary", "--refresh", "--format", "json"}) {
				t.Fatal("invalid readiness query")
			}
		}
		if spec.Mutating {
			writes++
			if len(spec.Stdin) == 0 {
				t.Fatal("missing stdin")
			}
		}
		if strings.HasSuffix(spec.ID, ".system_user") {
			if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"user", "info", "--access-key", "fixture-access", "--zone-id", "zone-id", "--format", "json"}) {
				t.Fatal("unscoped user query")
			}
			if _, ok := spec.SensitiveArgs[3]; !ok {
				t.Fatal("unmarked access key")
			}
		} else if strings.Contains(strings.Join(spec.Args, " "), "fixture") {
			t.Fatal("credential in argv")
		}
	}
	if writes != 1 || len(e.specs) != 11 {
		t.Fatalf("commands=%d writes=%d", len(e.specs), writes)
	}
}

func TestRealmImportExistingRealmAndPlacement(t *testing.T) {
	for _, match := range []bool{true, false} {
		s, _, cluster := newCephUserService(t)
		e := &realmImportExecutor{responses: realmImportResponses()}
		e.responses["realms"] = `{"realms":["realm"]}`
		e.responses["existing_realm"] = `{"id":"id","name":"realm"}`
		if !match {
			e.responses["existing_realm"] = `{"id":"other","name":"realm"}`
		}
		s.executor = e
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: realmImportParameters(t)})
		if (err == nil) != match {
			t.Fatal("realm identity check failed")
		}
		if !match {
			for _, c := range e.specs {
				if c.Mutating {
					t.Fatal("conflicting realm mutated")
				}
			}
		}
	}
	for _, replacement := range []string{`"rgw_frontend_port":81`, `"rgw_frontend_port":80,"rgw_zone":"wrong"`} {
		s, _, cluster := newCephUserService(t)
		e := &realmImportExecutor{responses: realmImportResponses()}
		e.responses["service_post_check"] = strings.Replace(e.responses["service_post_check"], `"rgw_frontend_port":80`, replacement, 1)
		s.executor = e
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: realmImportParameters(t)}); err == nil {
			t.Fatal("wrong service accepted")
		}
	}
	s, _, cluster := newCephUserService(t)
	e := &realmImportExecutor{responses: realmImportResponses()}
	s.executor = e
	p := realmImportParameters(t)
	p["placement"] = map[string]any{"count": 2}
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: p}); err == nil {
		t.Fatal("unverified placement accepted")
	}
}
func TestRealmImportStopsAndSanitizesFailures(t *testing.T) {
	for _, stage := range []string{"zones", "realms", "service_absence", "rgw_realm.import", "realm_post_check", "zone_post_check", "period_post_check", "service_post_check", "deployment_service", "deployment_daemons", "system_user"} {
		t.Run(stage, func(t *testing.T) {
			s, _, cluster := newCephUserService(t)
			e := &realmImportExecutor{responses: realmImportResponses(), fail: stage}
			s.executor = e
			_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: realmImportParameters(t)})
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Retryable || strings.Contains(err.Error(), "fixture-secret") {
				t.Fatal("unsafe error", err)
			}
			writes := 0
			for _, spec := range e.specs {
				if spec.Mutating {
					writes++
				}
			}
			if writes > 1 || ((stage == "zones" || stage == "realms" || stage == "service_absence") && writes != 0) {
				t.Fatal("unexpected mutation")
			}
		})
	}
	for stage, bad := range map[string]string{"zones": `{"zones":["secondary"]}`, "realms": `{"realms":null}`, "service_absence": `[{}]`, "realm_post_check": `{"id":"wrong"}`, "zone_post_check": `{"id":"zone-id","name":"secondary","realm_id":"wrong"}`, "period_post_check": `{"id":"period","realm_id":"id","period_map":{"zonegroups":[]}}`, "service_post_check": `[{"service_name":"wrong"}]`} {
		s, _, cluster := newCephUserService(t)
		responses := realmImportResponses()
		responses[stage] = bad
		e := &realmImportExecutor{responses: responses}
		s.executor = e
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: realmImportParameters(t)}); err == nil {
			t.Fatalf("accepted bad %s", stage)
		}
	}
}
