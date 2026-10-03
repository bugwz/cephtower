package operation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	externalservice "cephtower/backend/internal/service/external"
	mutationservice "cephtower/backend/internal/service/mutation"
)

type mutationExecutorFake struct {
	request mutationservice.Request
	result  cephdomain.ActionResult
	err     error
}

func (f *mutationExecutorFake) Execute(_ context.Context, request mutationservice.Request) (cephdomain.ActionResult, error) {
	f.request = request
	return f.result, f.err
}

type externalExecutorFake struct {
	request externalservice.Request
}

func (f *externalExecutorFake) Execute(_ context.Context, request externalservice.Request) (cephdomain.ActionResult, error) {
	f.request = request
	return cephdomain.ActionResult{Details: map[string]any{"external": true}}, nil
}

type reconcileExecutorFake struct {
	modules       []string
	kinds         []string
	kind          string
	refreshResult bool
	err           error
}

func (f *reconcileExecutorFake) RefreshKinds(_ context.Context, _ uint64, kinds []string) (cephdomain.ActionResult, error) {
	f.kinds = append([]string(nil), kinds...)
	return cephdomain.ActionResult{Details: map[string]any{"kinds": kinds}}, f.err
}

func (f *reconcileExecutorFake) Refresh(_ context.Context, _ uint64, modules []string) (cephdomain.ActionResult, error) {
	f.modules = append([]string(nil), modules...)
	return cephdomain.ActionResult{Details: map[string]any{"modules": modules}}, f.err
}

func (f *reconcileExecutorFake) RefreshKindIfSupported(_ context.Context, _ uint64, kind string) (bool, error) {
	f.kind = kind
	return f.refreshResult, f.err
}

func TestActionDispatcherReconcilesNativeMutation(t *testing.T) {
	mutations := &mutationExecutorFake{result: cephdomain.ActionResult{Details: map[string]any{"exit_code": 0}}}
	reconciler := &reconcileExecutorFake{refreshResult: true}
	dispatcher := NewActionDispatcher(mutations, nil, reconciler)
	result, err := dispatcher.Execute(context.Background(), ExecutionRequest{
		ClusterID: 7, Action: "pool.create", ResourceKind: "pool", ResourceKey: "pool-a",
		Parameters: map[string]any{"name": "pool-a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if mutations.request.Action != "pool.create" || reconciler.kind != "pool" {
		t.Fatalf("mutation=%#v reconciled kind=%q", mutations.request, reconciler.kind)
	}
	if result.Details.(map[string]any)["reconciled"] != true {
		t.Fatalf("result = %#v", result)
	}
}

func TestActionDispatcherFailsWhenPostReconcileFails(t *testing.T) {
	dispatcher := NewActionDispatcher(&mutationExecutorFake{}, nil, &reconcileExecutorFake{refreshResult: true, err: errors.New("offline")})
	_, err := dispatcher.Execute(context.Background(), ExecutionRequest{ClusterID: 7, Action: "pool.create", ResourceKind: "pool"})
	var actionError *cephdomain.ActionError
	if !errors.As(err, &actionError) || actionError.Code != "post_reconcile_failed" || !actionError.Retryable {
		t.Fatalf("error = %#v", err)
	}
}

func TestGlobalScheduleDoesNotRetryAfterInventoryFailure(t *testing.T) {
	reconciler := &reconcileExecutorFake{refreshResult: true, err: errors.New("offline")}
	dispatcher := NewActionDispatcher(&mutationExecutorFake{}, nil, reconciler)
	_, err := dispatcher.Execute(context.Background(), ExecutionRequest{ClusterID: 7, Action: "rbd_mirroring.global_schedule", ResourceKind: "rbd_mirroring"})
	var failure *cephdomain.ActionError
	if !errors.As(err, &failure) || failure.Code != "post_reconcile_failed" || failure.Retryable || reconciler.kind != "rbd_mirroring" {
		t.Fatal(err, reconciler)
	}
}

func TestFilesystemRenameRefreshesAffectedStorage(t *testing.T) {
	for _, fail := range []bool{false, true} {
		mutations := &mutationExecutorFake{result: cephdomain.ActionResult{Details: map[string]any{"native_output": "renamed"}}}
		reconciler := &reconcileExecutorFake{}
		if fail {
			reconciler.err = errors.New("fixture refresh failure")
		}
		result, err := NewActionDispatcher(mutations, nil, reconciler).Execute(context.Background(), ExecutionRequest{ClusterID: 7, Action: "filesystem.rename", ResourceKind: "filesystem", ResourceKey: "filesystem/old"})
		if !reflect.DeepEqual(reconciler.kinds, []string{"filesystem", "pool"}) || reconciler.kind != "" {
			t.Fatalf("unexpected refresh: %+v", reconciler)
		}
		if fail {
			var actionError *cephdomain.ActionError
			if !errors.As(err, &actionError) || actionError.Code != "post_reconcile_failed" {
				t.Fatalf("refresh error = %v", err)
			}
		} else if err != nil || result.Details.(map[string]any)["reconciled"] != true || result.Details.(map[string]any)["native_output"] != "renamed" {
			t.Fatalf("result = %+v, error = %v", result, err)
		}
	}
}

func TestRGWAccountMigrationRefreshesAffectedResources(t *testing.T) {
	for _, fail := range []bool{false, true} {
		mutations := &mutationExecutorFake{result: cephdomain.ActionResult{Details: map[string]any{}}}
		reconciler := &reconcileExecutorFake{}
		if fail {
			reconciler.err = errors.New("refresh failed")
		}
		_, err := NewActionDispatcher(mutations, nil, reconciler).Execute(context.Background(), ExecutionRequest{ClusterID: 7, Action: "rgw_user.update", ResourceKind: "rgw_user", Parameters: map[string]any{"target_account_id": "RGW12345678901234567"}})
		if !reflect.DeepEqual(reconciler.kinds, []string{"rgw_user", "rgw_account", "rgw_bucket"}) {
			t.Fatalf("unexpected refresh: %+v", reconciler)
		}
		if fail {
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Code != "post_reconcile_failed" || actionErr.Retryable {
				t.Fatal(err)
			}
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRGWSubuserRefreshDoesNotRepeatMutation(t *testing.T) {
	for _, action := range []string{"rgw_user.delete", "rgw_user.caps", "rgw_user.subuser", "rgw_key.create", "rgw_key.update", "rgw_key.delete"} {
		for _, fail := range []bool{false, true} {
			mutations := &mutationExecutorFake{result: cephdomain.ActionResult{Details: map[string]any{}}}
			reconciler := &reconcileExecutorFake{refreshResult: true}
			if fail {
				reconciler.err = errors.New("refresh failed")
			}
			kind := "rgw_user"
			if action == "rgw_key.create" || action == "rgw_key.update" || action == "rgw_key.delete" {
				kind = "rgw_key"
			}
			_, err := NewActionDispatcher(mutations, nil, reconciler).Execute(context.Background(), ExecutionRequest{ClusterID: 7, Action: action, ResourceKind: kind})
			if reconciler.kind != "rgw_user" || len(reconciler.kinds) != 0 {
				t.Fatalf("unexpected refresh: %+v", reconciler)
			}
			if fail {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_reconcile_failed" || actionErr.Retryable {
					t.Fatal(err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestRGWUserCreationRefreshDoesNotRetryCreation(t *testing.T) {
	for _, params := range []map[string]any{nil, {"account_id": "RGW12345678901234567", "account_root": false}, {"account_id": "RGW12345678901234567", "account_root": true}} {
		for _, fail := range []bool{false, true} {
			mutations := &mutationExecutorFake{result: cephdomain.ActionResult{Details: map[string]any{"created": true}}}
			reconciler := &reconcileExecutorFake{refreshResult: true}
			if fail {
				reconciler.err = errors.New("refresh failed")
			}
			result, err := NewActionDispatcher(mutations, nil, reconciler).Execute(context.Background(), ExecutionRequest{ClusterID: 7, Action: "rgw_user.create", ResourceKind: "rgw_user", Parameters: params})
			if params == nil {
				if reconciler.kind != "rgw_user" || len(reconciler.kinds) != 0 {
					t.Fatalf("unexpected independent user refresh: %+v", reconciler)
				}
			} else if !reflect.DeepEqual(reconciler.kinds, []string{"rgw_user", "rgw_account"}) || reconciler.kind != "" {
				t.Fatalf("unexpected account user refresh: %+v", reconciler)
			}
			if fail {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_reconcile_failed" || actionErr.Retryable {
					t.Fatalf("unsafe refresh failure: %v", err)
				}
			} else if err != nil || result.Details.(map[string]any)["reconciled"] != true || result.Details.(map[string]any)["created"] != true {
				t.Fatalf("unexpected result: %+v %v", result, err)
			}
		}
	}
	mutations := &mutationExecutorFake{err: errors.New("creation failed")}
	reconciler := &reconcileExecutorFake{}
	_, err := NewActionDispatcher(mutations, nil, reconciler).Execute(context.Background(), ExecutionRequest{ClusterID: 7, Action: "rgw_user.create", ResourceKind: "rgw_user", Parameters: map[string]any{"account_id": "RGW12345678901234567"}})
	if err != mutations.err || reconciler.kind != "" || len(reconciler.kinds) != 0 {
		t.Fatal("refreshed after failed creation")
	}
}

func TestActionDispatcherRoutesRefreshAndExternalActions(t *testing.T) {
	reconciler := &reconcileExecutorFake{}
	external := &externalExecutorFake{}
	dispatcher := NewActionDispatcher(nil, external, reconciler)
	if _, err := dispatcher.Execute(context.Background(), ExecutionRequest{
		ClusterID: 7, Action: "cluster.refresh", Parameters: map[string]any{"modules": []any{"fast", "topology"}},
	}); err != nil {
		t.Fatal(err)
	}
	if len(reconciler.modules) != 2 || reconciler.modules[0] != "fast" {
		t.Fatalf("refresh modules = %v", reconciler.modules)
	}
	if _, err := dispatcher.Execute(context.Background(), ExecutionRequest{
		ClusterID: 7, Action: "cluster.refresh", Parameters: map[string]any{"kind": "pool"},
	}); err != nil {
		t.Fatal(err)
	}
	if len(reconciler.kinds) != 1 || reconciler.kinds[0] != "pool" {
		t.Fatalf("refresh kinds = %v", reconciler.kinds)
	}
	if _, err := dispatcher.Execute(context.Background(), ExecutionRequest{
		ClusterID: 7, Action: "silence.delete", ResourceKind: "silence", ResourceKey: "silence-a",
	}); err != nil {
		t.Fatal(err)
	}
	if external.request.Action != "silence.delete" || external.request.ResourceKey != "silence-a" {
		t.Fatalf("external request = %#v", external.request)
	}
}
