package mutation

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestOSDPreviewReturnsOutputWithoutPostCheck(t *testing.T) {
	service, runner, id := newCephUserService(t)
	result, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd_deployment.preview", ResourceKey: "osd-deployment/preview", Parameters: map[string]any{"data_devices": map[string]any{"all": true}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 1 {
		t.Fatalf("preview executed extra commands: %+v", runner.specs)
	}
	spec := runner.specs[0]
	if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"orch", "apply", "osd", "-i", "-", "--dry-run"}) {
		t.Fatalf("wrong preview command: %+v", spec)
	}
	details, ok := result.Details.(map[string]any)
	if !ok {
		t.Fatalf("missing details: %+v", result)
	}
	output, ok := details["preview"].(string)
	if !ok || !strings.Contains(output, "[osd]") || strings.Contains(output, "synthetic-secret") {
		t.Fatalf("preview output lost or unredacted: %q", output)
	}
	runner.fail = true
	if _, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd_deployment.preview", Parameters: map[string]any{"data_devices": map[string]any{"all": true}}}); err == nil {
		t.Fatal("failed preview reported success")
	}
}
