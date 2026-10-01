package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestSubvolumeQuotaReadbackExecution(t *testing.T) {
	for _, tt := range []struct {
		name, size, post, failID string
		unlimited, valid         bool
	}{
		{"numeric exact", "9007199254740993", `{"bytes_quota":9007199254740993}`, "", false, true},
		{"string exact", "9007199254740993", `{"bytes_quota":"9007199254740993"}`, "", false, true},
		{"rounded", "9007199254740993", `{"bytes_quota":9007199254740992}`, "", false, false},
		{"unchanged", "2048", `{"bytes_quota":1024}`, "", false, false},
		{"missing", "2048", `{}`, "", false, false},
		{"null", "2048", `{"bytes_quota":null}`, "", false, false},
		{"boolean", "2048", `{"bytes_quota":true}`, "", false, false},
		{"trailing", "2048", `{"bytes_quota":2048} {}`, "", false, false},
		{"malformed", "2048", `{`, "", false, false},
		{"unlimited", "", `{"bytes_quota":"infinite"}`, "", true, true},
		{"unlimited unchanged", "", `{"bytes_quota":2048}`, "", true, false},
		{"read failed", "2048", `{"bytes_quota":2048}`, "subvolume.update.post_check", false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"subvolume.update.post_check": tt.post}, failID: tt.failID}
			s.executor = runner
			parameters := map[string]any{"group": "team"}
			size := tt.size
			if tt.unlimited {
				parameters["unlimited"] = true
				size = "inf"
			} else {
				parameters["size"] = size
			}
			_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "subvolume.update", ResourceKey: "filesystem/cephfs/subvolume/home", Parameters: parameters})
			if tt.valid {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || !actionErr.Retryable {
					t.Fatalf("error=%v", err)
				}
			}
			if len(runner.specs) != 2 || !reflect.DeepEqual(runner.specs[0].Args, []string{"fs", "subvolume", "resize", "cephfs", "home", size, "team"}) || !reflect.DeepEqual(runner.specs[1].Args, []string{"fs", "subvolume", "info", "cephfs", "home", "team", "--format", "json"}) {
				t.Fatalf("specs=%+v", runner.specs)
			}
			if !runner.specs[0].Mutating || runner.specs[1].Mutating {
				t.Fatal("readback mutability is invalid")
			}
		})
	}
}
