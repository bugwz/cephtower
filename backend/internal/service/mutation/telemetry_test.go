package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestTelemetryRequiresExplicitLicense(t *testing.T) {
	for _, parameters := range []map[string]any{{}, {"enabled": "true"}, {"enabled": true}, {"enabled": true, "license": "other"}, {"enabled": false, "license": "sharing-1-0"}} {
		if _, err := telemetryCommand(parameters); err == nil {
			t.Fatalf("accepted invalid consent: %v", parameters)
		}
	}
	for _, enabled := range []bool{true, false} {
		parameters := map[string]any{"enabled": enabled}
		expected := []string{"telemetry", "off"}
		if enabled {
			parameters["license"] = "sharing-1-0"
			expected = []string{"telemetry", "on", "--license", "sharing-1-0"}
		}
		command, err := build(Request{Action: "telemetry.update", ResourceKey: "manager-module/telemetry"}, parameters)
		if err != nil || !reflect.DeepEqual(command.args, expected) || !reflect.DeepEqual(command.check, []string{"telemetry", "status", "--format", "json"}) {
			t.Fatalf("command=%v err=%v", command, err)
		}
	}
}

func TestTelemetryMutationVerifiesRequestedState(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, enabled := range []bool{false, true} {
		parameters := map[string]any{"enabled": enabled}
		if enabled {
			parameters["license"] = "sharing-1-0"
		}
		for _, output := range []string{`{"enabled":true}`, `{"enabled":false}`, `{}`, `null`, `{"enabled":"true"}`, `{"enabled":true} {}`} {
			e := &directoryRenameExecutor{outputs: map[string]string{"telemetry.update.post_check": output}}
			s.executor = e
			_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "telemetry.update", ResourceKey: "manager-module/telemetry", Parameters: parameters})
			valid := (enabled && output == `{"enabled":true}`) || (!enabled && output == `{"enabled":false}`)
			if valid && err != nil {
				t.Fatal(err)
			}
			if !valid {
				var actionError *cephdomain.ActionError
				if !errors.As(err, &actionError) || actionError.Code != "post_check_failed" || actionError.Retryable {
					t.Fatalf("unverified state accepted: %s err=%v", output, err)
				}
			}
			if len(e.specs) != 2 || !e.specs[0].Mutating || e.specs[1].Mutating {
				t.Fatalf("unexpected command chain: %+v", e.specs)
			}
		}
	}
}
