package ceph

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

type currentPeriodExecutor struct {
	t            *testing.T
	body, period string
	fail         bool
	calls        []executor.CommandSpec
}

func (e *currentPeriodExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	switch spec.ID {
	case "collect.rgw_realm":
		return executor.CommandResult{Stdout: []byte(`{"realms":["named"]}`)}, nil
	case "collect.rgw_realm_detail":
		return executor.CommandResult{Stdout: []byte(`{"id":"realm-id","name":"named","current_period":"` + e.period + `"}`)}, nil
	case "collect.rgw_current_period":
		e.calls = append(e.calls, spec)
		if e.fail {
			return executor.CommandResult{}, errors.New("unavailable")
		}
		return executor.CommandResult{Stdout: []byte(e.body)}, nil
	}
	return fixtureExecutor{e.t}.Run(ctx, access, spec)
}
func TestCollectCurrentPeriodScopedAndOptional(t *testing.T) {
	for _, tc := range []struct {
		name, body, period string
		fail, want         bool
		calls              int
	}{
		{"valid", `{"id":"p-id","realm_id":"realm-id","epoch":2,"realm_epoch":1,"master_zone":"z","period_map":{"zonegroups":[]},"extension":9007199254740993}`, "p-id", false, true, 1},
		{"wrong realm", `{"id":"p-id","realm_id":"other"}`, "p-id", false, false, 1},
		{"wrong period", `{"id":"other","realm_id":"realm-id"}`, "p-id", false, false, 1},
		{"null", `null`, "p-id", false, false, 1},
		{"malformed", `broken`, "p-id", false, false, 1},
		{"failure", ``, "p-id", true, false, 1},
		{"no current", ``, "", false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &currentPeriodExecutor{t: t, body: tc.body, period: tc.period, fail: tc.fail}
			provider := NativeProvider{Executor: runner}
			found := false
			for _, row := range provider.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
				if row.Kind != "rgw_realm" {
					continue
				}
				found = true
				data := row.Payload.(map[string]any)
				period, ok := data["current_period_details"].(map[string]any)
				if ok != tc.want {
					t.Fatalf("period: %#v", data)
				}
				if ok {
					if period["extension"] != json.Number("9007199254740993") || period["id"] != "p-id" {
						t.Fatalf("period changed: %#v", period)
					}
				}
				if row.NaturalKey != "named" || data["id"] != "realm-id" {
					t.Fatal("realm identity changed")
				}
			}
			if !found || len(runner.calls) != tc.calls {
				t.Fatalf("found %v calls %v", found, runner.calls)
			}
			for _, call := range runner.calls {
				if call.Mutating || call.Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(call.Args, []string{"period", "get", "--realm-id", "realm-id", "--period", "p-id", "--format", "json"}) {
					t.Fatalf("command: %+v", call)
				}
			}
		})
	}
	if !reflect.DeepEqual(collectionFailureKinds["collect.rgw_current_period"], []string{"rgw_realm"}) {
		t.Fatal("collection coverage missing")
	}
}
