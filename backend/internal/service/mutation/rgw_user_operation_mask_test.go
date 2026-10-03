package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWUserOperationMaskExecution(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, mask := range []string{"read", "write", "delete", "read,write", "read,delete", "write,delete", "read,write,delete"} {
		for _, state := range []string{"unchanged", "suspend", "enable"} {
			t.Run(mask+"/"+state, func(t *testing.T) {
				params := map[string]any{"op_mask": mask}
				if state != "unchanged" {
					params["suspended"] = state == "suspend"
				}
				suspension := "0"
				if state == "suspend" {
					suspension = "1"
				}
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.update.post_check": `{"full_user_id":"tenant$user","suspended":` + suspension + `,"op_mask":"` + strings.ReplaceAll(mask, ",", ", ") + `"}`}}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user", Parameters: params})
				if err != nil {
					t.Fatal(err)
				}
				want := [][]string{{"user", "modify", "--uid", "tenant$user", "--op-mask=" + mask, "--format", "json"}}
				if state != "unchanged" {
					want = append(want, []string{"user", state, "--uid", "tenant$user", "--format", "json"})
				}
				want = append(want, []string{"user", "info", "--uid", "tenant$user", "--format", "json"})
				if len(runner.specs) != len(want) {
					t.Fatalf("unexpected calls: %v", runner.specs)
				}
				for i, spec := range runner.specs {
					if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (i < len(want)-1) || !reflect.DeepEqual(spec.Args, want[i]) {
						t.Fatalf("step %d: %+v", i, spec)
					}
				}
			})
		}
	}
	for _, raw := range []string{`{}`, `null`, `{`, `{} {}`, `{"full_user_id":"other$user","op_mask":"read"}`, `{"full_user_id":"tenant$user","op_mask":null}`, `{"full_user_id":"tenant$user","op_mask":7}`, `{"full_user_id":"tenant$user","op_mask":"write"}`, `{"full_user_id":null,"op_mask":"read"}`} {
		t.Run(raw, func(t *testing.T) {
			service.executor = &directoryRenameExecutor{outputs: map[string]string{"rgw_user.update.post_check": raw}}
			_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user", Parameters: map[string]any{"op_mask": "read"}})
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
	for _, failID := range []string{"rgw_user.update", "rgw_user.update.post_check"} {
		runner := &directoryRenameExecutor{failID: failID}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user", Parameters: map[string]any{"op_mask": "read"}})
		if err == nil {
			t.Fatalf("accepted %s failure", failID)
		}
		if failID == "rgw_user.update" && len(runner.specs) != 1 {
			t.Fatal("read after failed write")
		}
		if failID == "rgw_user.update.post_check" {
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
				t.Fatalf("unexpected result: %v", err)
			}
		}
	}
}
