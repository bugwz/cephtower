package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

type zonegroupPolicyExecutor struct {
	t        *testing.T
	policies map[string]string
	calls    []string
}

func (e *zonegroupPolicyExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.ID == "collect.rgw_zonegroup" {
		return executor.CommandResult{Stdout: []byte(`{"zonegroups":["east","west","empty","missing"],"default_info":"east-id"}`)}, nil
	}
	if spec.ID == "collect.rgw_zonegroup_detail" {
		if spec.Mutating || spec.Binary != executor.BinaryRGWAdmin || len(spec.Args) != 6 {
			e.t.Fatalf("unsafe detail command: %+v", spec)
		}
		name := spec.Args[3]
		if !reflect.DeepEqual(spec.Args, []string{"zonegroup", "get", "--rgw-zonegroup", name, "--format", "json"}) {
			e.t.Fatalf("wrong command: %+v", spec)
		}
		e.calls = append(e.calls, name)
		extra := ""
		if policy, ok := e.policies[name]; ok {
			extra = `,"sync_policy":` + policy
		}
		return executor.CommandResult{Stdout: []byte(`{"name":"` + name + `","id":"` + name + `-id","zones":[]` + extra + `}`)}, nil
	}
	return fixtureExecutor{e.t}.Run(ctx, access, spec)
}

func TestZonegroupSyncPolicyPreservesScopedNativeData(t *testing.T) {
	east := `{"groups":[{"id":"g","status":"enabled","data_flow":{"symmetrical":[{"id":"flow","zones":["zone-id-a","zone-id-b"]}]},"pipes":[{"id":"p","source":{"bucket":"*","zones":["*"]},"dest":{"bucket":"tenant/photos","zones":["zone-id-b"]},"params":{"priority":9007199254740993,"mode":"system","source":{"filter":{"tags":[]}},"dest":{}}}]}]}`
	west := `{"groups":[{"id":"g","status":"forbidden","data_flow":{"directional":[{"source_zone":"west-id-a","dest_zone":"west-id-b"}]},"pipes":[]}]}`
	runner := &zonegroupPolicyExecutor{t: t, policies: map[string]string{"east": east, "west": west, "empty": `{"groups":[]}`}}
	provider := NativeProvider{Executor: runner}
	found := map[string]bool{}
	for _, row := range provider.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
		if row.Kind != "rgw_zonegroup" {
			continue
		}
		found[row.NaturalKey] = true
		payload := row.Payload.(map[string]any)
		expected, exists := runner.policies[row.NaturalKey]
		if !exists {
			if _, ok := payload["sync_policy"]; ok {
				t.Fatal("missing policy synthesized")
			}
			continue
		}
		var want any
		decoder := json.NewDecoder(strings.NewReader(expected))
		decoder.UseNumber()
		if err := decoder.Decode(&want); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(payload["sync_policy"], want) {
			t.Fatalf("policy drift for %s: %#v", row.NaturalKey, payload["sync_policy"])
		}
		if payload["is_default"] != (row.NaturalKey == "east") {
			t.Fatal("default identity drift")
		}
	}
	if len(found) != 4 || !reflect.DeepEqual(runner.calls, []string{"east", "west", "empty", "missing"}) {
		t.Fatalf("scope drift: %v %v", found, runner.calls)
	}
}
