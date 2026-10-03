package mutation

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type syncFlowExecutor struct {
	*syncGroupExecutor
	zoneCalls int
	zoneBody  *string
	zoneErr   error
}

func (e *syncFlowExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.ID == "rgw_bucket.sync_flow_create.zones" || spec.ID == "rgw_bucket.sync_flow_delete.zones" || spec.ID == "rgw_bucket.sync_pipe_create.zones" {
		e.zoneCalls++
		if spec.Mutating || spec.Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(spec.Args, []string{"zonegroup", "get", "--format", "json"}) {
			return executor.CommandResult{}, errors.New("wrong zone lookup")
		}
		if e.zoneErr != nil {
			return executor.CommandResult{}, e.zoneErr
		}
		if e.zoneBody != nil {
			return executor.CommandResult{Stdout: []byte(*e.zoneBody)}, nil
		}
		return executor.CommandResult{Stdout: []byte(`{"zones":[{"id":"a","name":"Zeta"},{"id":"b","name":"Alpha"}]}`)}, nil
	}
	return e.syncGroupExecutor.Run(ctx, access, spec)
}

func TestBucketSyncFlowMappingFailuresNeverWrite(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[]}`
	for _, tc := range []struct {
		name, body string
		err        error
	}{
		{"command unavailable", "", errors.New("zonegroup unavailable")},
		{"malformed", "broken", nil},
		{"missing zones", `{}`, nil},
		{"null zones", `{"zones":null}`, nil},
		{"empty zones", `{"zones":[]}`, nil},
		{"unknown requested ID", `{"zones":[{"id":"a","name":"A"}]}`, nil},
		{"duplicate ID", `{"zones":[{"id":"a","name":"A"},{"id":"a","name":"B"},{"id":"b","name":"C"}]}`, nil},
		{"duplicate name", `{"zones":[{"id":"a","name":"A"},{"id":"b","name":"A"}]}`, nil},
		{"missing name", `{"zones":[{"id":"a"},{"id":"b","name":"B"}]}`, nil},
		{"wrong field type", `{"zones":[{"id":4,"name":"A"},{"id":"b","name":"B"}]}`, nil},
		{"trailing document", `{"zones":[{"id":"a","name":"A"},{"id":"b","name":"B"}]}{}`, nil},
	} {
		for _, kind := range []string{"symmetrical", "directional"} {
			for _, tenant := range []string{"", "team"} {
				t.Run(tc.name+"/"+kind+"/"+tenant, func(t *testing.T) {
					service, _, clusterID := newCephUserService(t)
					runner := &syncFlowExecutor{syncGroupExecutor: &syncGroupExecutor{before: `{"groups":[` + group + `]}`}, zoneBody: &tc.body, zoneErr: tc.err}
					service.executor = runner
					p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos")), "group_id": "g", "expected_group": group, "flow_type": kind}
					if kind == "symmetrical" {
						p["flow_id"] = "f"
						p["zones"] = []any{"a", "b"}
					} else {
						p["source_zone"] = "a"
						p["dest_zone"] = "b"
					}
					_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_flow_create", Parameters: p})
					var actionError *cephdomain.ActionError
					if !errors.As(err, &actionError) || actionError.Code != "pre_check_failed" || actionError.Retryable {
						t.Fatalf("unexpected error: %v", err)
					}
					if runner.zoneCalls != 1 || len(runner.calls) != 1 || runner.calls[0].Mutating {
						t.Fatalf("unexpected execution: zones=%d commands=%+v", runner.zoneCalls, runner.calls)
					}
				})
			}
		}
	}
}

func TestBucketSyncFlowCreation(t *testing.T) {
	for _, kind := range []string{"symmetrical", "directional"} {
		flow := `{"id":" new ","zones":["Zeta","Alpha"]}`
		extra := []string{"--flow-id", " new ", "--zone-ids", "a,b"}
		params := map[string]any{"flow_type": kind, "flow_id": " new ", "zones": []any{"b", "a"}}
		if kind == "directional" {
			flow = `{"source_zone":"Zeta","dest_zone":"Alpha"}`
			extra = []string{"--flow-id", "directional", "--source-zone-id", "a", "--dest-zone-id", "b"}
			params = map[string]any{"flow_type": kind, "source_zone": "a", "dest_zone": "b"}
		}
		group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"priority":9007199254740993}]}`
		changed := strings.Replace(group, `"data_flow":{}`, `"data_flow":{"`+kind+`":[`+flow+`]}`, 1)
		oldFlow := strings.ReplaceAll(strings.ReplaceAll(flow, " new ", "old"), "Zeta", "Gamma")
		existing := strings.Replace(group, `"data_flow":{}`, `"data_flow":{"`+kind+`":[`+oldFlow+`]}`, 1)
		appended := strings.Replace(group, `"data_flow":{}`, `"data_flow":{"`+kind+`":[`+oldFlow+`,`+flow+`]}`, 1)
		policy := func(g string) string {
			return `{"groups":[` + g + `,{"id":"other","status":"allowed","data_flow":{},"pipes":[]}]}`
		}
		for _, tc := range []struct {
			name, before, expected, after, code string
			fail, count                         int
		}{
			{"create", group, group, policy(changed), "", 0, 3},
			{"append preserves old flow", existing, existing, policy(appended), "", 0, 3},
			{"exists", changed, changed, policy(changed), "pre_check_failed", 0, 1},
			{"stale", group, changed, policy(changed), "pre_check_failed", 0, 1},
			{"malformed expectation", group, "broken", policy(changed), "pre_check_failed", 0, 1},
			{"pre read", group, group, policy(changed), "pre_check_failed", 1, 1},
			{"write", group, group, policy(changed), "command_failed", 2, 2},
			{"post read", group, group, policy(changed), "post_check_failed", 3, 3},
			{"unchanged", group, group, policy(group), "post_check_failed", 0, 3},
			{"other removed", group, group, `{"groups":[` + changed + `]}`, "post_check_failed", 0, 3},
			{"pipe changed", group, group, strings.Replace(policy(changed), "9007199254740993", "9007199254740992", 1), "post_check_failed", 0, 3},
		} {
			for _, tenant := range []string{"", "team"} {
				t.Run(kind+"/"+tc.name+"/"+tenant, func(t *testing.T) {
					s, _, clusterID := newCephUserService(t)
					runner := &syncGroupExecutor{before: policy(tc.before), after: tc.after, failAt: tc.fail}
					wrapped := &syncFlowExecutor{syncGroupExecutor: runner}
					s.executor = wrapped
					p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos")), "group_id": "g", "expected_group": tc.expected}
					for k, v := range params {
						p[k] = v
					}
					_, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_flow_create", Parameters: p})
					if len(runner.calls) != tc.count {
						t.Fatalf("calls=%d", len(runner.calls))
					}
					if tc.count > 1 && wrapped.zoneCalls != 1 {
						t.Fatal("zone mapping not checked")
					}
					for i, call := range runner.calls {
						args := []string{"sync", "policy", "get"}
						if i == 1 {
							args = append([]string{"sync", "group", "flow", "create", "--group-id", "g", "--flow-type", kind}, extra...)
						}
						args = append(args, "--bucket", "photos", "--tenant", tenant, "--format", "json")
						if call.Binary != executor.BinaryRGWAdmin || call.Mutating != (i == 1) || !reflect.DeepEqual(args, call.Args) {
							t.Fatalf("wrong command: %+v", call)
						}
					}
					if tc.code == "" {
						if err != nil {
							t.Fatal(err)
						}
						return
					}
					var e *cephdomain.ActionError
					if !errors.As(err, &e) || e.Code != tc.code || e.Retryable {
						t.Fatalf("error: %v", err)
					}
				})
			}
		}
	}
}

func TestBucketSyncFlowValidation(t *testing.T) {
	base := map[string]any{"expected_group": "{}", "flow_type": "symmetrical", "flow_id": "f", "zones": []any{"a", "b"}}
	for _, change := range []map[string]any{
		{"expected_group": ""}, {"flow_type": "unknown"}, {"flow_id": "-bad"}, {"flow_id": "a\nb"}, {"zones": []any{}}, {"zones": []any{"a", "a"}}, {"zones": []any{"a,b"}}, {"zones": []any{"*"}}, {"zones": []any{" a"}}, {"zones": []any{3}}, {"source_zone": "a"},
	} {
		p := map[string]any{}
		for k, v := range base {
			p[k] = v
		}
		for k, v := range change {
			p[k] = v
		}
		if _, err := bucketSyncFlowArgs(p); err == nil {
			t.Fatalf("accepted: %+v", p)
		}
	}
	for _, p := range []map[string]any{
		{"expected_group": "{}", "flow_type": "directional", "source_zone": "a", "dest_zone": "a"},
		{"expected_group": "{}", "flow_type": "directional", "source_zone": "a", "dest_zone": "b", "flow_id": "f"},
		{"expected_group": "{}", "flow_type": "directional", "source_zone": "a", "dest_zone": "b", "zones": []any{}},
	} {
		if _, err := bucketSyncFlowArgs(p); err == nil {
			t.Fatalf("accepted: %+v", p)
		}
	}
}

func TestBucketSyncFlowZoneMapping(t *testing.T) {
	p := map[string]any{"flow_type": "symmetrical", "zones": []any{"b", "a"}}
	for _, body := range []string{"broken", `{}`, `{"zones":[{"id":"a","name":"A"}]}`, `{"zones":[{"id":"a","name":"A"},{"id":"b","name":"A"}]}`, `{"zones":[{"id":"a","name":"A"},{"id":"a","name":"B"}]}`} {
		if _, err := resolveBucketSyncFlow(p, []byte(body)); err == nil {
			t.Fatalf("accepted bad mapping: %s", body)
		}
	}
	result, err := resolveBucketSyncFlow(p, []byte(`{"zones":[{"id":"a","name":"Zeta"},{"id":"b","name":"Alpha"}]}`))
	if err != nil || !reflect.DeepEqual(result["zones"], []any{"Zeta", "Alpha"}) || !reflect.DeepEqual(p["zones"], []any{"b", "a"}) {
		t.Fatalf("mapping/order: %+v %v", result, err)
	}
	for _, id := range []string{"a;b", "a=b"} {
		if _, err := bucketSyncFlowArgs(map[string]any{"expected_group": "{}", "flow_type": "symmetrical", "flow_id": "f", "zones": []any{id}}); err == nil {
			t.Fatalf("accepted delimiter: %s", id)
		}
	}
}
