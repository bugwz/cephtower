package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWUserPlacementExecution(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, tc := range []struct {
		name, response, failID string
		wantSuccess            bool
	}{
		{"verified", `{"default_placement":"custom","default_storage_class":"","placement_tags":["fast"," raw "]}`, "", true},
		{"wrong rule", `{"default_placement":"old","default_storage_class":"","placement_tags":["fast"," raw "]}`, "", false},
		{"null class", `{"default_placement":"custom","default_storage_class":null,"placement_tags":["fast"," raw "]}`, "", false},
		{"missing class", `{"default_placement":"custom","placement_tags":["fast"," raw "]}`, "", false},
		{"wrong tags", `{"default_placement":"custom","default_storage_class":"","placement_tags":["fast"]}`, "", false},
		{"empty object", `{}`, "", false},
		{"null", `null`, "", false},
		{"malformed", `{`, "", false},
		{"trailing data", `{} {}`, "", false},
		{"read failed", `{}`, "rgw_user.update.post_check", false},
		{"write failed", `{}`, "rgw_user.update", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.update.post_check": tc.response}, failID: tc.failID}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/test", Parameters: map[string]any{"default_placement": "custom", "default_storage_class": "", "placement_tags_csv": "fast, raw "}})
			if tc.wantSuccess {
				if err != nil {
					t.Fatal(err)
				}
			} else if tc.failID == "rgw_user.update" {
				if err == nil {
					t.Fatal("write failure reported success")
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			calls := 2
			if tc.failID == "rgw_user.update" {
				calls = 1
			}
			if len(runner.specs) != calls {
				t.Fatalf("calls=%d want=%d", len(runner.specs), calls)
			}
			if runner.specs[0].Binary != executor.BinaryRGWAdmin || !runner.specs[0].Mutating {
				t.Fatal("expected native write")
			}
			if calls == 2 && (runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, []string{"user", "info", "--uid", "test", "--format", "json"})) {
				t.Fatal("unexpected readback")
			}
		})
	}
}

func TestRGWUserPlacementIndependentExecution(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, tc := range []struct {
		name      string
		params    map[string]any
		response  string
		writeArgs []string
	}{
		{"rule only", map[string]any{"default_placement": "custom", "default_storage_class": "ARCHIVE"}, `{"default_placement":"custom","default_storage_class":"ARCHIVE"}`, []string{"--placement-id=custom", "--storage-class=ARCHIVE"}},
		{"tags only", map[string]any{"placement_tags_csv": "fast,archive"}, `{"placement_tags":["fast","archive"]}`, []string{"--tags=fast,archive"}},
	} {
		for _, state := range []string{"unchanged", "suspend", "enable"} {
			for _, response := range []string{tc.response, `{}`} {
				t.Run(tc.name+"/"+state+"/"+response, func(t *testing.T) {
					params := make(map[string]any)
					for key, value := range tc.params {
						params[key] = value
					}
					if state != "unchanged" {
						params["suspended"] = state == "suspend"
					}
					runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.update.post_check": response}}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user", Parameters: params})
					if response == tc.response {
						if err != nil {
							t.Fatal(err)
						}
					} else {
						var actionErr *cephdomain.ActionError
						if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
							t.Fatalf("unexpected error: %v", err)
						}
					}
					want := [][]string{append(append([]string{"user", "modify", "--uid", "tenant$user"}, tc.writeArgs...), "--format", "json")}
					if state != "unchanged" {
						want = append(want, []string{"user", state, "--uid", "tenant$user", "--format", "json"})
					}
					want = append(want, []string{"user", "info", "--uid", "tenant$user", "--format", "json"})
					if len(runner.specs) != len(want) {
						t.Fatalf("calls=%d want=%d", len(runner.specs), len(want))
					}
					for index, spec := range runner.specs {
						if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (index < len(want)-1) || !reflect.DeepEqual(spec.Args, want[index]) {
							t.Fatalf("step %d: %+v", index, spec)
						}
					}
				})
			}
		}
	}
}
