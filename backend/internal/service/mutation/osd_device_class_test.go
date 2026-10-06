package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"fmt"
	"reflect"
	"testing"
)

type osdClassExecutor struct {
	specs               []executor.CommandSpec
	before, after, fail string
}

func (e *osdClassExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.ID == "osd.device_class."+e.fail {
		return executor.CommandResult{}, fmt.Errorf("offline")
	}
	if spec.ID == "osd.device_class.pre_check" {
		return executor.CommandResult{Stdout: []byte(e.before)}, nil
	}
	return executor.CommandResult{Stdout: []byte(e.after)}, nil
}

func TestOSDDeviceClassMutation(t *testing.T) {
	for _, scenario := range []string{"change", "same", "unclassified", "conflict", "invalid-read", "remove-failure", "set-failure", "wrong-after"} {
		t.Run(scenario, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &osdClassExecutor{before: `[{"osd":0,"device_class":"hdd"}]`, after: `[{"osd":0,"device_class":"ssd"}]`}
			service.executor = runner
			p := map[string]any{"device_class": "ssd", "expected_class": "hdd"}
			stages := []string{"pre_check", "remove", "set", "post_check"}
			success := scenario == "change" || scenario == "same" || scenario == "unclassified"
			switch scenario {
			case "same":
				p["device_class"] = "hdd"
				stages = stages[:1]
			case "unclassified":
				p["expected_class"] = ""
				runner.before = `[{"osd":0,"device_class":""}]`
				stages = []string{"pre_check", "set", "post_check"}
			case "conflict":
				p["expected_class"] = "nvme"
				stages = stages[:1]
			case "invalid-read":
				runner.before = `[{"device_class":"hdd"}]`
				stages = stages[:1]
			case "remove-failure":
				runner.fail = "remove"
				stages = stages[:2]
			case "set-failure":
				runner.fail = "set"
				stages = stages[:3]
			case "wrong-after":
				runner.after = `[{"osd":0,"device_class":"hdd"}]`
			}
			result, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.device_class", ResourceKey: "osd/0", Parameters: p})
			if (err == nil) != success {
				t.Fatalf("result=%v err=%v", result, err)
			}
			if len(runner.specs) != len(stages) {
				t.Fatalf("specs=%+v", runner.specs)
			}
			for i, stage := range stages {
				spec := runner.specs[i]
				if spec.ID != "osd.device_class."+stage || spec.Mutating != (stage == "remove" || stage == "set") {
					t.Fatalf("spec=%+v", spec)
				}
				args := []string{"osd", "crush", "get-device-class", "0", "--format", "json"}
				if stage == "remove" {
					args = []string{"osd", "crush", "rm-device-class", "0"}
				}
				if stage == "set" {
					args = []string{"osd", "crush", "set-device-class", "ssd", "0"}
				}
				if !reflect.DeepEqual(spec.Args, args) {
					t.Fatalf("args=%v", spec.Args)
				}
			}
		})
	}
}

func TestOSDDeviceClassRejectsInvalidInput(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &osdClassExecutor{}
	service.executor = runner
	for _, target := range []string{"-1", "01", "--help", "2147483648"} {
		if _, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.device_class", ResourceKey: "osd/" + target, Parameters: map[string]any{"device_class": "ssd", "expected_class": "hdd"}}); err == nil {
			t.Fatal("invalid id accepted")
		}
	}
	for _, class := range []string{"", "--help", "ssd;id", " ssd"} {
		if _, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.device_class", ResourceKey: "osd/0", Parameters: map[string]any{"device_class": class, "expected_class": "hdd"}}); err == nil {
			t.Fatal("invalid class accepted")
		}
	}
	if len(runner.specs) != 0 {
		t.Fatal("invalid input executed")
	}
}
