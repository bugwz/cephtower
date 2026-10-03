package mutation

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRBDConfigurationMutationReadback(t *testing.T) {
	for _, tc := range []struct {
		verb, output string
		want         bool
	}{
		{"config-set", `[{"name":"rbd_qos_iops_limit","value":"1000","source":"image"}]`, true},
		{"config-set", `[{"name":"rbd_qos_iops_limit","value":"1000","source":"pool"}]`, false},
		{"config-set", `[{"name":"rbd_qos_iops_limit","value":"2000","source":"image"}]`, false},
		{"config-remove", `[{"name":"rbd_qos_iops_limit","value":"1000","source":"image"}]`, false},
		{"config-remove", `[{"name":"rbd_qos_iops_limit","value":"1000","source":"pool"}]`, true},
		{"config-remove", `[{"name":"rbd_qos_iops_limit","value":"0","source":"config"}]`, true},
		{"config-remove", `[{"name":"rbd_qos_iops_limit","value":"0","source":"unknown (4)"}]`, false},
		{"config-remove", `[]`, false},
		{"config-set", `null`, false},
		{"config-remove", `[{}]`, false},
		{"config-remove", `[{"name":"rbd_qos_iops_limit","source":"pool"}]`, false},
		{"config-set", `[{"name":"rbd_qos_iops_limit","value":1000,"source":"image"}]`, false},
		{"config-remove", `[{"name":"other","value":"0","source":"pool"}]`, false},
		{"config-set", `[{"name":"rbd_qos_iops_limit","value":"1000","source":"image"},{"name":"rbd_qos_iops_limit","value":"0","source":"pool"}]`, false},
		{"config-set", `[{"name":"rbd_qos_iops_limit","value":"1000","source":"image"},null]`, false},
	} {
		service, _, clusterID := newCephUserService(t)
		runner := &directoryRenameExecutor{outputs: map[string]string{"rbd_image.action.post_check": tc.output}}
		service.executor = runner
		params := map[string]any{"action": tc.verb, "config_name": "rbd_qos_iops_limit", "config_value": "1000"}
		_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rbd_image.action", ResourceKey: "rbd/image/" + base64.RawURLEncoding.EncodeToString([]byte("images/team/vm")) + "/action", Parameters: params})
		if tc.want {
			if err != nil {
				t.Fatalf("%s %s: %v", tc.verb, tc.output, err)
			}
		} else {
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
				t.Fatalf("%s %s: err=%v", tc.verb, tc.output, err)
			}
		}
		want := []string{"config", "image", "list", "images/team/vm", "--format", "json"}
		if len(runner.specs) != 2 || runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, want) {
			t.Fatalf("incorrect readback commands: %#v", runner.specs)
		}
	}
}
