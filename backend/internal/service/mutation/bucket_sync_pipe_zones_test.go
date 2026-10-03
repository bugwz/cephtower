package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

const pipeZonesMapping = `{"zones":[{"id":"a","name":"Zeta"},{"id":"b","name":"Alpha"},{"id":"c","name":"Gamma"}]}`

func TestBucketSyncPipeZonesSingleStage(t *testing.T) {
	for _, tc := range []struct {
		name, before, after, verb, delta string
		desired                          []any
		count                            int
	}{
		{"wildcard to list", `["*"]`, `["Zeta","Alpha"]`, "modify", "a,b", []any{"b", "a"}, 4},
		{"list to wildcard", `["Zeta","Alpha"]`, `["*"]`, "modify", "*", []any{"*"}, 4},
		{"only remove", `["Zeta","Alpha"]`, `["Alpha"]`, "remove", "a", []any{"b"}, 4},
		{"only add", `["Zeta"]`, `["Zeta","Alpha"]`, "modify", "b", []any{"a", "b"}, 4},
		{"no change", `["Zeta","Alpha"]`, `["Zeta","Alpha"]`, "", "", []any{"b", "a"}, 2},
	} {
		for _, side := range []string{"source", "dest"} {
			t.Run(tc.name+"/"+side, func(t *testing.T) {
				group := func(zones string) string {
					return `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","` + side + `":{"bucket":"*","zones":` + zones + `},"` + map[string]string{"source": "dest", "dest": "source"}[side] + `":{"bucket":"keep","zones":["*"]},"params":{"mode":"system"}}]}`
				}
				policy := func(g string) string { return `{"groups":[` + g + `]}` }
				service, _, clusterID := newCephUserService(t)
				runner := &pipeZonesExecutor{bodies: map[string]string{"pre_check": policy(group(tc.before)), "zones": pipeZonesMapping, "add_post_check": policy(group(tc.after)), "remove_post_check": policy(group(tc.after))}}
				service.executor = runner
				p := map[string]any{"bucket_id": "AGJ1Y2tldA", "group_id": "g", "pipe_id": "p", "expected_group": group(tc.before), "source_zones": []any{"*"}, "dest_zones": []any{"*"}}
				p[side+"_zones"] = tc.desired
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_pipe_zones", Parameters: p})
				if tc.count == 2 {
					var ae *cephdomain.ActionError
					if !errors.As(err, &ae) || ae.Code != "pre_check_failed" {
						t.Fatalf("error %v", err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				if len(runner.calls) != tc.count {
					t.Fatalf("calls %v", runner.calls)
				}
				if tc.count == 4 {
					args := runner.calls[2].Args
					if args[3] != tc.verb || !reflect.DeepEqual(args[len(args)-2:], []string{"--" + side + "-zone-ids", tc.delta}) {
						t.Fatalf("args %v", args)
					}
					for _, arg := range args {
						if arg == "--source-bucket" || arg == "--dest-bucket" || arg == "--mode" {
							t.Fatalf("unrelated flag %s", arg)
						}
					}
				}
			})
		}
	}
}

func TestPipeZonePlanning(t *testing.T) {
	for _, tc := range []struct {
		name, old     string
		desired       []any
		add, remove   []string
		middle, final []any
	}{
		{"replace", `["Zeta","Alpha"]`, []any{"c", "b"}, []string{"c"}, []string{"a"}, []any{"Zeta", "Alpha", "Gamma"}, []any{"Alpha", "Gamma"}},
		{"add", `["Zeta"]`, []any{"b", "a"}, []string{"b"}, nil, []any{"Zeta", "Alpha"}, []any{"Zeta", "Alpha"}},
		{"remove", `["Zeta","Alpha"]`, []any{"b"}, nil, []string{"a"}, []any{"Zeta", "Alpha"}, []any{"Alpha"}},
		{"wildcard to list", `["*"]`, []any{"b", "a"}, []string{"a", "b"}, nil, []any{"Zeta", "Alpha"}, []any{"Zeta", "Alpha"}},
		{"list to wildcard", `["Zeta","Alpha"]`, []any{"*"}, []string{"*"}, nil, []any{"*"}, []any{"*"}},
		{"same wildcard", `["*"]`, []any{"*"}, nil, nil, []any{"*"}, []any{"*"}},
		{"same list", `["Zeta","Alpha"]`, []any{"b", "a"}, nil, nil, []any{"Zeta", "Alpha"}, []any{"Zeta", "Alpha"}},
		{"empty to list", `[]`, []any{"a"}, []string{"a"}, nil, []any{"Zeta"}, []any{"Zeta"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var old []any
			if err := json.Unmarshal([]byte(tc.old), &old); err != nil {
				t.Fatal(err)
			}
			entity := map[string]any{"bucket": "keep", "zones": old}
			change, err := planPipeZones(entity, tc.desired, []byte(pipeZonesMapping))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(change.added, tc.add) || !reflect.DeepEqual(change.removed, tc.remove) || !reflect.DeepEqual(change.intermediate, tc.middle) || !reflect.DeepEqual(change.final, tc.final) {
				t.Fatalf("plan: %+v", change)
			}
			if !reflect.DeepEqual(entity["zones"], old) {
				t.Fatal("planning mutated entity")
			}
		})
	}
	for _, old := range []any{nil, "A", []any{nil}, []any{"missing"}, []any{"*", "Zeta"}, []any{"Zeta", "Zeta"}} {
		if _, err := planPipeZones(map[string]any{"zones": old}, []any{"a"}, []byte(pipeZonesMapping)); err == nil {
			t.Fatalf("accepted old %v", old)
		}
	}
	for _, desired := range []any{nil, []any{}, []any{"unknown"}, []any{"a", "a"}, []any{"*", "a"}, []any{"a;b"}, []any{"a=b"}, []any{"a b"}} {
		if _, err := planPipeZones(map[string]any{"zones": []any{"Zeta"}}, desired, []byte(pipeZonesMapping)); err == nil {
			t.Fatalf("accepted desired %v", desired)
		}
	}
	for _, body := range []string{`broken`, `{}`, `{"zones":[{"id":"a","name":"*"}]}`, `{"zones":[{"id":"a","name":"A"},{"id":"b","name":"A"}]}`} {
		if _, err := planPipeZones(map[string]any{"zones": []any{"*"}}, []any{"*"}, []byte(body)); err == nil {
			t.Fatalf("accepted mapping %s", body)
		}
	}
}

type pipeZonesExecutor struct {
	calls   []executor.CommandSpec
	bodies  map[string]string
	failure string
}

func (e *pipeZonesExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_bucket.sync_pipe_zones.")
	if stage == e.failure {
		return executor.CommandResult{}, errors.New("injected failure")
	}
	return executor.CommandResult{Stdout: []byte(e.bodies[stage])}, nil
}

func TestBucketSyncPipeZonesExecution(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{},"pipes":[{"id":"p","source":{"bucket":"photos","zones":["Zeta","Alpha"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"user","user":"keep","priority":9007199254740993}}]}`
	policy := func(g string) string {
		return `{"groups":[` + g + `,{"id":"other","status":"allowed","data_flow":{},"pipes":[]}]}`
	}
	middle := strings.Replace(strings.Replace(group, `["Zeta","Alpha"]`, `["Zeta","Alpha","Gamma"]`, 1), `["*"]`, `["Zeta"]`, 1)
	final := strings.Replace(middle, `["Zeta","Alpha","Gamma"]`, `["Alpha","Gamma"]`, 1)
	stages := []string{"pre_check", "zones", "add", "add_post_check", "remove", "remove_post_check"}
	for _, tenant := range []string{"", "team"} {
		for _, tc := range []struct {
			name, failure, middle, after, expected, code string
			count                                        int
		}{
			{name: "success", count: 6},
			{name: "stale", expected: final, code: "pre_check_failed", count: 1},
			{name: "pre failure", failure: "pre_check", code: "pre_check_failed", count: 1},
			{name: "mapping failure", failure: "zones", code: "pre_check_failed", count: 2},
			{name: "add failure", failure: "add", code: "command_failed", count: 3},
			{name: "add read failure", failure: "add_post_check", code: "post_check_failed", count: 4},
			{name: "add mismatch", middle: policy(group), code: "post_check_failed", count: 4},
			{name: "remove failure", failure: "remove", code: "command_failed", count: 5},
			{name: "remove read failure", failure: "remove_post_check", code: "post_check_failed", count: 6},
			{name: "other params changed", after: policy(strings.Replace(final, "9007199254740993", "9007199254740992", 1)), code: "post_check_failed", count: 6},
		} {
			t.Run(tenant+"/"+tc.name, func(t *testing.T) {
				service, _, clusterID := newCephUserService(t)
				runner := &pipeZonesExecutor{failure: tc.failure, bodies: map[string]string{"pre_check": policy(group), "zones": pipeZonesMapping, "add_post_check": policy(middle), "remove_post_check": policy(final)}}
				if tc.middle != "" {
					runner.bodies["add_post_check"] = tc.middle
				}
				if tc.after != "" {
					runner.bodies["remove_post_check"] = tc.after
				}
				expected := tc.expected
				if expected == "" {
					expected = group
				}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_pipe_zones", Parameters: map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos")), "group_id": "g", "pipe_id": "p", "expected_group": expected, "source_zones": []any{"b", "c"}, "dest_zones": []any{"a"}}})
				if tc.code == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var ae *cephdomain.ActionError
					if !errors.As(err, &ae) || ae.Code != tc.code || ae.Retryable {
						t.Fatalf("error %v", err)
					}
				}
				if len(runner.calls) != tc.count {
					t.Fatalf("calls %v", runner.calls)
				}
				for i, call := range runner.calls {
					stage := stages[i]
					if call.ID != "rgw_bucket.sync_pipe_zones."+stage || call.Mutating != (stage == "add" || stage == "remove") {
						t.Fatalf("call %+v", call)
					}
					if !call.Mutating {
						continue
					}
					verb := "modify"
					if stage == "remove" {
						verb = "remove"
					}
					want := []string{"sync", "group", "pipe", verb, "--group-id", "g", "--pipe-id", "p", "--bucket", "photos", "--tenant", tenant, "--format", "json"}
					if stage == "add" {
						want = append(want, "--source-zone-ids", "c", "--dest-zone-ids", "a")
					} else {
						want = append(want, "--source-zone-ids", "a")
					}
					if !reflect.DeepEqual(call.Args, want) {
						t.Fatalf("args %v want %v", call.Args, want)
					}
				}
			})
		}
	}
}
