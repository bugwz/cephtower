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

type syncFlowUpdateExecutor struct {
	calls   []executor.CommandSpec
	bodies  map[string]string
	failure string
}

func (e *syncFlowUpdateExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_bucket.sync_flow_update.")
	if stage == e.failure {
		return executor.CommandResult{}, errors.New("injected failure")
	}
	return executor.CommandResult{Stdout: []byte(e.bodies[stage])}, nil
}

func TestBucketSyncFlowUpdate(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{"symmetrical":[{"id":"f","zones":["Zeta","Alpha"]}],"directional":[{"source_zone":"Zeta","dest_zone":"Gamma"}]},"pipes":[{"id":"p","priority":9007199254740993}]}`
	policy := func(g string) string {
		return `{"groups":[` + g + `,{"id":"other","status":"allowed","data_flow":{},"pipes":[]}]}`
	}
	union := strings.Replace(group, `["Zeta","Alpha"]`, `["Zeta","Alpha","Gamma"]`, 1)
	final := strings.Replace(group, `["Zeta","Alpha"]`, `["Alpha","Gamma"]`, 1)
	for _, tenant := range []string{"", "team"} {
		for _, tc := range []struct {
			name, failure, before, post, expected, code string
			middle                                      string
			zones                                       []any
			stages                                      []string
		}{
			{name: "both", zones: []any{"c", "b"}, stages: []string{"pre_check", "zones", "add", "add_post_check", "remove", "remove_post_check"}},
			{name: "add only", zones: []any{"a", "b", "c"}, stages: []string{"pre_check", "zones", "add", "add_post_check"}},
			{name: "remove only", zones: []any{"b"}, post: strings.Replace(group, `["Zeta","Alpha"]`, `["Alpha"]`, 1), stages: []string{"pre_check", "zones", "remove", "remove_post_check"}},
			{name: "unchanged", zones: []any{"b", "a"}, code: "pre_check_failed", stages: []string{"pre_check", "zones"}},
			{name: "stale", expected: union, code: "pre_check_failed", stages: []string{"pre_check"}},
			{name: "missing", before: strings.Replace(group, `"id":"f"`, `"id":"other"`, 1), code: "pre_check_failed", stages: []string{"pre_check"}},
			{name: "unknown old zone", before: strings.Replace(group, "Zeta", "Unknown", 1), code: "pre_check_failed", stages: []string{"pre_check", "zones"}},
			{name: "pre failure", failure: "pre_check", code: "pre_check_failed", stages: []string{"pre_check"}},
			{name: "mapping failure", failure: "zones", code: "pre_check_failed", stages: []string{"pre_check", "zones"}},
			{name: "add failure", failure: "add", code: "command_failed", stages: []string{"pre_check", "zones", "add"}},
			{name: "add read failure", failure: "add_post_check", code: "post_check_failed", stages: []string{"pre_check", "zones", "add", "add_post_check"}},
			{name: "add unchanged", middle: policy(group), code: "post_check_failed", stages: []string{"pre_check", "zones", "add", "add_post_check"}},
			{name: "add malformed", middle: `broken`, code: "post_check_failed", stages: []string{"pre_check", "zones", "add", "add_post_check"}},
			{name: "add changed other data", middle: policy(strings.Replace(union, "9007199254740993", "9007199254740992", 1)), code: "post_check_failed", stages: []string{"pre_check", "zones", "add", "add_post_check"}},
			{name: "remove failure", failure: "remove", code: "command_failed", stages: []string{"pre_check", "zones", "add", "add_post_check", "remove"}},
			{name: "remove read failure", failure: "remove_post_check", code: "post_check_failed", stages: []string{"pre_check", "zones", "add", "add_post_check", "remove", "remove_post_check"}},
			{name: "unexpected other mutation", post: strings.Replace(final, "9007199254740993", "9007199254740992", 1), code: "post_check_failed", stages: []string{"pre_check", "zones", "add", "add_post_check", "remove", "remove_post_check"}},
		} {
			t.Run(tc.name+"/"+tenant, func(t *testing.T) {
				service, _, clusterID := newCephUserService(t)
				before := tc.before
				if before == "" {
					before = group
				}
				expected := tc.expected
				if expected == "" {
					expected = before
				}
				post := tc.post
				if post == "" {
					post = final
				}
				zones := tc.zones
				if zones == nil {
					zones = []any{"b", "c"}
				}
				runner := &syncFlowUpdateExecutor{failure: tc.failure, bodies: map[string]string{"pre_check": policy(before), "zones": `{"zones":[{"id":"a","name":"Zeta"},{"id":"b","name":"Alpha"},{"id":"c","name":"Gamma"}]}`, "add_post_check": policy(union), "remove_post_check": policy(post)}}
				service.executor = runner
				if tc.middle != "" {
					runner.bodies["add_post_check"] = tc.middle
				}
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_flow_update", Parameters: map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos")), "group_id": "g", "flow_id": "f", "expected_group": expected, "zones": zones}})
				if tc.code == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Code != tc.code || actionErr.Retryable {
						t.Fatalf("error: %v", err)
					}
				}
				stages := []string{}
				for _, call := range runner.calls {
					stage := strings.TrimPrefix(call.ID, "rgw_bucket.sync_flow_update.")
					stages = append(stages, stage)
					if call.Mutating != (stage == "add" || stage == "remove") {
						t.Fatalf("wrong mutating flag: %+v", call)
					}
					if stage == "add" || stage == "remove" {
						verb, delta := "create", "c"
						if stage == "remove" {
							verb, delta = "remove", "a"
						}
						if call.Args[3] != verb {
							t.Fatalf("wrong verb: %v", call.Args)
						}
						for flag, value := range map[string]string{"--zone-ids": delta, "--bucket": "photos", "--tenant": tenant, "--flow-id": "f"} {
							found := false
							for i, arg := range call.Args {
								if arg == flag && i+1 < len(call.Args) && call.Args[i+1] == value {
									found = true
								}
							}
							if !found {
								t.Fatalf("missing %s=%s in %v", flag, value, call.Args)
							}
						}
					}
				}
				if !reflect.DeepEqual(stages, tc.stages) {
					t.Fatalf("stages %v want %v", stages, tc.stages)
				}
			})
		}
	}
}
