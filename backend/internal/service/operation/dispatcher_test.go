package operation

import (
	"context"
	"errors"
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
