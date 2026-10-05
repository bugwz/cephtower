package mutation

import (
	"context"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestGroupPlacementPreservesPrimarySensitiveArguments(t *testing.T) {
	for _, failure := range []string{"", "add"} {
		t.Run("failure_"+failure, func(t *testing.T) {
			p := targetParams()
			spec, err := buildCloudTarget(p)
			if err != nil {
				t.Fatal(err)
			}
			spec.sensitive = map[int]struct{}{len(spec.args) - 1: {}}
			r := targetFixture()
			r.fail = failure
			s := &Service{executor: r}
			_, err = s.executeZonegroupStorageClass(context.Background(), executor.ClusterAccess{}, Request{Action: "rgw_zonegroup.cloud_target", Parameters: p}, spec)
			if (err != nil) != (failure != "") {
				t.Fatal(err)
			}
			found := false
			for _, call := range r.calls {
				if call.ID == "rgw_zonegroup.cloud_target.add" {
					found = true
					if !reflect.DeepEqual(call.SensitiveArgs, spec.sensitive) || !reflect.DeepEqual(call.Args, spec.args) {
						t.Fatal("primary argument metadata changed")
					}
				} else if len(call.SensitiveArgs) != 0 {
					t.Fatal("primary argument indices leaked into another command", call.ID)
				}
			}
			if !found {
				t.Fatal("primary command not executed")
			}
		})
	}
}
