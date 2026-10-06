package mutation

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type deviceLightExecutor struct {
	specs     []executor.CommandSpec
	raw, fail string
}

func (e *deviceLightExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.ID == "device.identify."+e.fail {
		return executor.CommandResult{}, fmt.Errorf("offline")
	}
	return executor.CommandResult{Stdout: []byte(e.raw)}, nil
}

func TestDeviceIdentifyNativeTarget(t *testing.T) {
	for _, state := range []string{"on", "off"} {
		for _, light := range []string{"ident", "fault"} {
			for _, scenario := range []string{"success", "alias", "wrong-host", "wrong-path", "wrong-id", "multiple", "empty", "malformed", "missing-id", "padded-id", "option-id", "relative-path", "pre_check", "apply"} {
				t.Run(state+"/"+light+"/"+scenario, func(t *testing.T) {
					s, _, cluster := newCephUserService(t)
					e := &deviceLightExecutor{raw: `{"devid":"serial","location":[{"host":"node1","dev":"sda","path":"/dev/disk/by-id/serial"}]}`}
					s.executor = e
					p := map[string]any{"device_id": "serial", "host": "node1", "device": "/dev/sda", "state": state, "light": light}
					switch scenario {
					case "alias":
						p["device"] = "/dev/disk/by-id/serial"
					case "wrong-host":
						p["host"] = "node2"
					case "wrong-path":
						p["device"] = "/dev/sdb"
					case "wrong-id":
						p["device_id"] = "other"
					case "multiple":
						e.raw = `{"devid":"serial","location":[{"host":"node1","dev":"sda"},{"host":"node2","dev":"sda"}]}`
					case "empty":
						e.raw = `{"devid":"serial","location":[]}`
					case "malformed":
						e.raw = `{`
					case "missing-id":
						delete(p, "device_id")
					case "padded-id":
						p["device_id"] = " serial "
					case "option-id":
						p["device_id"] = "--help"
					case "relative-path":
						p["device"] = "sda"
					case "pre_check", "apply":
						e.fail = scenario
					}
					result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "device.identify", ResourceKey: "device/node1/serial/identify", Parameters: p})
					success := scenario == "success" || scenario == "alias"
					if (err == nil) != success {
						t.Fatalf("result=%v err=%v", result, err)
					}
					want := 1
					if success || scenario == "apply" {
						want = 2
					}
					if scenario == "missing-id" || scenario == "padded-id" || scenario == "option-id" || scenario == "relative-path" {
						want = 0
					}
					if len(e.specs) != want {
						t.Fatalf("specs=%+v", e.specs)
					}
					for i, spec := range e.specs {
						args := []string{"device", "info", p["device_id"].(string), "--format", "json"}
						if i == 1 {
							args = []string{"device", "light", state, "serial", light}
						}
						if !reflect.DeepEqual(spec.Args, args) || spec.Mutating != (i == 1) {
							t.Fatalf("spec=%+v", spec)
						}
					}
					if success && result.Details.(map[string]any)["physical_state_verified"] != false {
						t.Fatalf("physical state must remain unverified: %v", result)
					}
				})
			}
		}
	}
}
