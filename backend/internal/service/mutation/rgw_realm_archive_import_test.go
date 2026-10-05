package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func archiveImportResponses(t *testing.T) map[string]string {
	t.Helper()
	encode := func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	responses := realmImportResponses()
	master := map[string]any{"id": "master", "name": "primary"}
	zone := map[string]any{"id": "zone-id", "name": "secondary", "tier_type": "archive", "sync_from_all": false, "sync_from": []string{"primary"}}
	group := map[string]any{"id": "zg", "name": "group", "realm_id": "id", "master_zone": "master", "zones": []any{master}}
	period := map[string]any{"id": "period", "epoch": 1, "realm_id": "id", "master_zonegroup": "zg", "period_map": map[string]any{"zonegroups": []any{group}}}
	responses["archive.pull"] = responses["realm_post_check"]
	responses["archive.period"] = encode(period)
	group["zones"] = []any{master, zone}
	responses["archive.group"] = encode(group)
	responses["archive.commit"] = encode(period)
	responses["archive.period_check"] = encode(period)
	responses["period_post_check"] = encode(period)
	responses["archive.realm_check"] = responses["realm_post_check"]
	responses["archive.create"] = `{"id":"zone-id","name":"secondary","realm_id":"id","system_key":{"access_key":"fixture-access","secret_key":"fixture-secret"}}`
	return responses
}

func TestArchiveImportNativeSequence(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	e := &realmImportExecutor{responses: archiveImportResponses(t)}
	s.executor = e
	p := realmImportParameters(t)
	p["tier_type"] = "archive"
	result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: p})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result)
	if strings.Contains(string(raw), "fixture-") || strings.Contains(string(raw), "token") {
		t.Fatal("secret in result")
	}
	var stages []string
	var deploy executor.CommandSpec
	for _, c := range e.specs {
		stages = append(stages, strings.TrimPrefix(c.ID, "rgw_realm.import."))
		for i, arg := range c.Args {
			if arg == "fixture-access" || arg == "fixture-secret" {
				if _, ok := c.SensitiveArgs[i]; !ok {
					t.Fatal("unmarked credential")
				}
			}
		}
		if c.ID == "rgw_realm.import.archive.create" {
			want := []string{"zone", "create", "--realm-id", "id", "--zonegroup-id", "zg", "--rgw-zone", "secondary", "--access-key", "fixture-access", "--secret", "fixture-secret", "--tier-type", "archive", "--sync-from-all=false", "--sync-from", "primary", "--format", "json"}
			if !c.Mutating || c.Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(c.Args, want) {
				t.Fatal("archive create args")
			}
		}
		if c.ID == "rgw_realm.import.archive.pull" {
			want := []string{"realm", "pull", "--rgw-realm", "realm", "--url", "https://master.example", "--access-key", "fixture-access", "--secret", "fixture-secret", "--format", "json"}
			if !c.Mutating || !reflect.DeepEqual(c.Args, want) || len(c.SensitiveArgs) != 2 {
				t.Fatal("unsafe realm pull")
			}
		}
		if c.ID == "rgw_realm.import.archive.commit" {
			want := []string{"period", "update", "--commit", "--realm-id", "id", "--zonegroup-id", "zg", "--zone-id", "zone-id", "--format", "json"}
			if !c.Mutating || !reflect.DeepEqual(c.Args, want) {
				t.Fatal("unscoped period commit")
			}
		}
		if c.ID == "rgw_realm.import.archive.deploy" {
			deploy = c
		}
	}
	want := []string{"zones", "realms", "service_absence", "archive.pull", "archive.period", "archive.create", "archive.group", "archive.commit", "archive.realm_check", "archive.period_check", "archive.deploy", "realm_post_check", "zone_post_check", "period_post_check", "service_post_check", "deployment_service", "deployment_daemons"}
	if !reflect.DeepEqual(stages, want) {
		t.Fatalf("unexpected stages %v", stages)
	}
	if !deploy.Mutating || deploy.Binary != executor.BinaryCeph || !reflect.DeepEqual(deploy.Args, []string{"orch", "apply", "-i", "-"}) {
		t.Fatal("invalid deploy")
	}
	var spec map[string]any
	if json.Unmarshal(deploy.Stdin, &spec) != nil || spec["rgw_realm"] != "realm" || spec["rgw_zonegroup"] != "group" || spec["rgw_zone"] != "secondary" || spec["update_endpoints"] != true {
		t.Fatal("invalid secondary spec")
	}
	tokenRaw, err := base64.StdEncoding.DecodeString(rawText(spec, "rgw_realm_token"))
	if err != nil {
		t.Fatal(err)
	}
	var token map[string]any
	if json.Unmarshal(tokenRaw, &token) != nil || token["endpoint"] != nil || token["realm_id"] != "id" || token["access_key"] != "fixture-access" || token["secret"] != "fixture-secret" {
		t.Fatal("wrong secondary token")
	}
}

func TestArchiveImportStopsAtEveryFailure(t *testing.T) {
	for _, stage := range []string{"archive.pull", "archive.period", "archive.create", "archive.group", "archive.commit", "archive.realm_check", "archive.period_check", "archive.deploy"} {
		t.Run(stage, func(t *testing.T) {
			s, _, cluster := newCephUserService(t)
			e := &realmImportExecutor{responses: archiveImportResponses(t), fail: stage}
			s.executor = e
			p := realmImportParameters(t)
			p["tier_type"] = "archive"
			_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: p})
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Retryable || strings.Contains(err.Error(), "fixture-secret") {
				t.Fatal("unsafe failure")
			}
			if e.specs[len(e.specs)-1].ID != "rgw_realm.import."+stage {
				t.Fatal("continued after failure")
			}
		})
	}
}

func TestArchiveImportRejectsChangedIdentityAndPolicy(t *testing.T) {
	for _, tc := range []struct{ stage, from, to string }{
		{"archive.pull", `"id":"id"`, `"id":"other"`},
		{"archive.period", `"master_zone":"master"`, `"master_zone":"missing"`},
		{"archive.period", `"name":"primary"`, `"name":"secondary"`},
		{"archive.period", `"name":"primary"`, `"name":"one,two"`},
		{"archive.create", `"secret_key":"fixture-secret"`, `"secret_key":"wrong"`},
		{"archive.group", `"sync_from_all":false`, `"sync_from_all":true`},
		{"archive.group", `"name":"primary"`, `"name":"renamed"`},
		{"archive.group", `"tier_type":"archive"`, `"tier_type":""`},
		{"archive.group", `"sync_from":["primary"]`, `"sync_from":["other"]`},
		{"archive.commit", `"tier_type":"archive"`, `"tier_type":""`},
		{"archive.commit", `"epoch":1`, `"epoch":0`},
		{"archive.realm_check", `"current_period":"period"`, `"current_period":"other"`},
		{"archive.period_check", `"epoch":1`, `"epoch":2`},
		{"period_post_check", `"tier_type":"archive"`, `"tier_type":""`},
	} {
		t.Run(tc.stage+tc.from, func(t *testing.T) {
			s, _, cluster := newCephUserService(t)
			responses := archiveImportResponses(t)
			if !strings.Contains(responses[tc.stage], tc.from) {
				t.Fatal("invalid test fixture")
			}
			responses[tc.stage] = strings.ReplaceAll(responses[tc.stage], tc.from, tc.to)
			e := &realmImportExecutor{responses: responses}
			s.executor = e
			p := realmImportParameters(t)
			p["tier_type"] = "archive"
			if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: p}); err == nil {
				t.Fatal("invalid archive import accepted")
			}
			if e.specs[len(e.specs)-1].ID != "rgw_realm.import."+tc.stage {
				t.Fatal("continued after invalid response")
			}
		})
	}
}
